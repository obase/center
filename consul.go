package center

import (
	"github.com/hashicorp/consul/api"
	"strings"
	"sync"
	"time"
)

const DEFAULT_EXPIRED int64 = 5 //与828 center默认值相同

type consulEntry struct {
	Mtime   int64      // 最后修改时间戳
	Index   uint64     // 最后刷新的下标
	Service []*Service // 服务项
}

type consulClient struct {
	*api.Client
	sync.RWMutex
	Entries map[string]*consulEntry
	Expired int64 // 缓存过期时间
}

func newConsulClient(opt *Config) Center {
	var client *api.Client
	var err error

	config := api.DefaultConfig()
	if opt.Address != "" {
		config.Address = opt.Address
	}
	if client, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
		return nil
	} else {
		if _, err = client.Agent().Self(); err != nil {
			return nil
		}
	}

	return &consulClient{
		Client:  client,
		Entries: make(map[string]*consulEntry),
		Expired: nvl(opt.Expired, DEFAULT_EXPIRED),
	}
}

func (c *consulClient) Register(service *Service, check *Check) (err error) {

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

	return c.Agent().ServiceRegister(consulService)
}
func (c *consulClient) Deregister(serviceId string) (err error) {
	return c.Agent().ServiceDeregister(serviceId)
}

func (c *consulClient) FetchService(name string) ([]*Service, uint64, error) {
	var (
		entry *consulEntry
		ok    bool
		err   error
		now   = time.Now().Unix()
	)
	c.RWMutex.RLock()
	entry, ok = c.Entries[name]
	c.RWMutex.RUnlock()
	if !ok || now-entry.Mtime > c.Expired {
		c.RWMutex.Lock()
		entry, ok = c.Entries[name]
		if !ok || now-entry.Mtime > c.Expired { // 二次检测
			entry = new(consulEntry)
			entry.Mtime = now
			entry.Service, entry.Index, err = c.WatchService(name, entry.Index)
			if err == nil {
				c.Entries[name] = entry
			}
		}
		c.RWMutex.Unlock()
	}
	return entry.Service, entry.Index, err
}
func (c *consulClient) WatchService(name string, index uint64) ([]*Service, uint64, error) {
	entries, metainfo, err := c.Client.Health().Service(name, "", true, &api.QueryOptions{
		WaitIndex: index,
	})
	if err != nil {
		return nil, 0, err
	}

	services := make([]*Service, len(entries))
	for i, entry := range entries {
		services[i] = &Service{
			Id:   entry.Service.ID,
			Kind: string(entry.Service.Kind),
			Name: name,
			Host: entry.Service.Address,
			Port: entry.Service.Port,
		}
	}
	return services, metainfo.LastIndex, nil
}
