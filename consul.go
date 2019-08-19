package center

import (
	"github.com/hashicorp/consul/api"
	"strings"
	"sync"
	"time"
)

const DEFAULT_EXPIRED int64 = 5 //与828 center默认值相同

type consulEntry struct {
	sync.RWMutex
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

	ret := &consulClient{
		Client:  client,
		Entries: make(map[string]*consulEntry),
		Expired: nvl(opt.Expired, DEFAULT_EXPIRED),
	}
	// 启动刷新后台,定期更新entires的数据
	go refreshConsulClientEntries(ret)

	return ret
}

const maxprocs = 8 // 最多起8个协程处理后台更新
func refreshConsulClientEntries(c *consulClient) {
	for _ = range time.Tick(time.Duration(c.Expired)*time.Second) {
		if len(c.Entries) <= maxprocs {
			// 数量不超maxprocs不需分组
			wg := new(sync.WaitGroup)
			c.RWMutex.RLock()
			for name, entry := range c.Entries {
				wg.Add(1)
				go func() {
					defer wg.Done()
					c.refresh(name, entry)
				}()
			}
			c.RWMutex.RUnlock()
			wg.Wait()
		} else {
			// 数量超过maxprocs需要分组
			set := make([]map[string]*consulEntry, maxprocs)
			cnt := 0
			c.RWMutex.RLock()
			for name, entry := range c.Entries {
				idx := cnt % maxprocs
				if set[idx] == nil {
					set[idx] = make(map[string]*consulEntry)
				}
				set[idx][name] = entry
				cnt++
			}
			c.RWMutex.RUnlock()

			wg := new(sync.WaitGroup)
			for _, part := range set {
				if len(part) > 0 {
					wg.Add(1)
					go func() {
						defer wg.Done()
						for name, entry := range part {
							c.refresh(name, entry)
						}
					}()
				}
			}
			wg.Wait()
		}

	}
}

func (c *consulClient) Register(service *Service, check *Check) (err error) {

	var consulCheck *api.AgentServiceCheck
	var consulService *api.AgentServiceRegistration

	if check != nil {
		switch strings.ToLower(check.Type) {
		case "http":
			consulCheck = &api.AgentServiceCheck{
				HTTP:                           check.Target,
				Timeout:                        check.Timeout,
				Interval:                       check.Interval,
				DeregisterCriticalServiceAfter: check.Interval,
			}
		case "grpc":
			consulCheck = &api.AgentServiceCheck{
				GRPC:                           check.Target,
				Timeout:                        check.Timeout,
				Interval:                       check.Interval,
				DeregisterCriticalServiceAfter: check.Interval,
			}
		case "tcp":
			consulCheck = &api.AgentServiceCheck{
				TCP:                            check.Target,
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

func (c *consulClient) FetchService(name string) (service []*Service, index uint64, err error) {
	var (
		entry *consulEntry
		ok    bool
	)
	c.RWMutex.RLock()
	entry, ok = c.Entries[name]
	c.RWMutex.RUnlock()
	if !ok {
		c.RWMutex.Lock()
		entry, ok = c.Entries[name]
		if !ok { // 二重检测
			entry = new(consulEntry)
			c.Entries[name] = entry
		}
		c.RWMutex.Unlock()
	}
	entry.RWMutex.RLock()
	service, index = entry.Service, entry.Index
	entry.RWMutex.RUnlock()
	if service == nil {
		entry.RWMutex.Lock()
		service, index = entry.Service, entry.Index
		if service == nil { // 二重检测
			service, index, err = c.refresh(name, entry)
		}
		entry.RWMutex.Unlock()
	}
	return
}

func (c *consulClient) refresh(name string, entry *consulEntry) (service []*Service, index uint64, err error) {
	if service, index, err = c.WatchService(name, 0); err == nil {
		if index != entry.Index {
			entry.RWMutex.Lock()
			entry.Service, entry.Index = service, index
			entry.RWMutex.Unlock()
		}
	}
	return
}

func (c *consulClient) WatchService(name string, index uint64) ([]*Service, uint64, error) {
	var options *api.QueryOptions
	if index > 0 {
		options = &api.QueryOptions{
			WaitIndex: index,
		}
	}
	entries, metainfo, err := c.Client.Health().Service(name, "", true, options)
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
