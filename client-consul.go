package center

import (
	"github.com/hashicorp/consul/api"
	"github.com/obase/log"
	"strings"
	"sync"
	"time"
)

type consulEntry struct {
	sync.RWMutex
	Index   uint64     // 最后刷新的下标
	Service []*Service // 服务项
}

type consulClient struct {
	*api.Client
	sync.RWMutex
	Entries map[string]*consulEntry
}

func newConsulClient(opt *Config) Center {
	var client *api.Client
	var err error

	config := api.DefaultConfig()
	if opt.Address != LOCAL {
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
	}
	// 启动刷新后台,定期更新entires的数据
	if opt.Expired > 0 {
		go func() {
			_expire, _refresh := opt.Expired, opt.Refresh
			for _ = range time.Tick(time.Duration(_expire) * time.Second) {
				ret.RefreshEntries(_refresh)
			}
		}()
	}

	return ret
}

func (c *consulClient) RefreshEntries(refresh int) {

	defer func() {
		if perr := recover(); perr != nil {
			log.ErrorStack("refreshConsulClientEntries panic: %v", perr)
		}
	}()

	if len(c.Entries) <= refresh {
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
		set := make([]map[string]*consulEntry, refresh)
		cnt := 0
		c.RWMutex.RLock()
		for name, entry := range c.Entries {
			idx := cnt % refresh
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
		service, index, err = c.refresh(name, entry) // FIXBUG: 在refresh已经加锁,外层不再需要锁
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
