package center

import (
	"fmt"
	"github.com/hashicorp/consul/api"
	"os"
	"strings"
	"sync"
	"time"
)

const (
	DefaultRefresh = 60 * time.Second
	ExpiredSeconds = 3600
)

type consulEntry struct {
	value []*Service // 服务项
	atime int64      // 最后时间戳(秒)
	index uint64
}

type consulClient struct {
	*api.Client
	*sync.RWMutex
	Service map[string]*consulEntry
}

func newConsulClient(opt *Config) Center {
	config := api.DefaultConfig()
	if opt.Address != "" {
		config.Address = opt.Address
	}
	var client *api.Client
	var err error
	if client, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
		return nil
	} else {
		if _, err = client.Agent().Self(); err != nil {
			return nil
		}
	}

	ret := &consulClient{
		Client:  client,
		RWMutex: new(sync.RWMutex),
		Service: make(map[string]*consulEntry),
	}
	// 启动后台刷新线程
	go func(interval time.Duration) {
		for _ = range time.Tick(interval) {
			ret.refresh()
		}
	}(nvl(opt.Refresh, DefaultRefresh))

	return ret
}

func (client *consulClient) refresh() {
	defer func() {
		if perr := recover(); perr != nil {
			fmt.Fprint(os.Stderr, "consul service refresh error: %v", perr)
		}
	}()

	var newdata map[string]*consulEntry = make(map[string]*consulEntry)
	var now = time.Now().Unix()
	client.RWMutex.RLock()
	for k, v := range client.Service {
		if now-v.atime < ExpiredSeconds {
			if ret, idx, err := client.discovery(k, v.index); err != nil && len(ret) > 0 {
				newdata[k] = &consulEntry{
					value: ret,
					atime: v.atime,
					index: idx,
				}
			} else {
				newdata[k] = v
			}
		}
	}
	client.RWMutex.RUnlock()

	// 直接切换掉
	client.RWMutex.Lock()
	client.Service = newdata
	client.RWMutex.Unlock()
}

func (client *consulClient) discovery(name string, fromLastIndex uint64) (services []*Service, lastIndex uint64, err error) {
	entries, metainfo, err := client.Health().Service(name, "", true, &api.QueryOptions{
		WaitIndex: fromLastIndex,
		WaitTime:  100 * time.Millisecond,
	})
	if err != nil {
		return
	}

	services = make([]*Service, len(entries))
	for i, entry := range entries {
		services[i] = &Service{
			Id:   entry.Service.ID,
			Kind: string(entry.Service.Kind),
			Name: name,
			Host: entry.Service.Address,
			Port: entry.Service.Port,
		}
	}
	lastIndex = metainfo.LastIndex

	return
}

func (client *consulClient) Register(service *Service, check *Check) (err error) {

	var consulCheck *api.AgentServiceCheck
	var consulService *api.AgentServiceRegistration

	if check != nil {
		switch strings.ToUpper(check.Type) {
		case "HTTP":
			consulCheck = &api.AgentServiceCheck{
				HTTP:                           check.Target,
				Timeout:                        check.Timeout,
				Interval:                       check.Interval,
				DeregisterCriticalServiceAfter: check.Interval,
			}
		case "GRPC":
			consulCheck = &api.AgentServiceCheck{
				GRPC:                           check.Target,
				Timeout:                        check.Timeout,
				Interval:                       check.Interval,
				DeregisterCriticalServiceAfter: check.Interval,
			}
		}
	}
	consulService = &api.AgentServiceRegistration{
		Kind:    api.ServiceKind(service.Kind),
		ID:      service.Id,
		Name:    service.Name,
		Address: service.Host,
		Port:    service.Port,
		Check:   consulCheck,
	}

	return client.Agent().ServiceRegister(consulService)
}
func (client *consulClient) Deregister(serviceId string) (err error) {
	return client.Agent().ServiceDeregister(serviceId)
}
func (client *consulClient) Discovery(name string) (ret []*Service, err error) {

	client.RWMutex.RLock()
	entry, ok := client.Service[name]
	client.RWMutex.RUnlock()

	if ok {
		ret = entry.value
		return
	}

	ret, idx, err := client.discovery(name, 0)
	if err == nil {
		client.RWMutex.Lock()
		client.Service[name] = &consulEntry{
			value: ret,
			atime: time.Now().Unix(),
			index: idx,
		}
		client.RWMutex.Unlock()
	}
	return
}
