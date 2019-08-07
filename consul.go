package center

import (
	"github.com/hashicorp/consul/api"
	"github.com/obase/log"
	"strings"
	"sync"
	"time"
)

type cacheEntry struct {
	vl []*Service
	ts int64
}

type consulCenter struct {
	*Config
	*api.Client
	ttl       int64
	now       int64
	lastIndex uint64
	cache     *sync.Map
	robin     uint32
}

func newConsulCenter(opt *Config) Center {
	config := api.DefaultConfig()
	if opt.Address != "" {
		config.Address = opt.Address
	}
	var client *api.Client
	var err error
	if client, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
		return nil
	} else {
		if _, err = client.Agent().Services(); err != nil {
			return nil
		}
	}
	return &consulCenter{
		Config: opt,
		Client: client,
		ttl:    int64(opt.Timeout.Seconds()),
		cache:  new(sync.Map),
	}
}
func (client *consulCenter) Register(service *Service, check *Check) (err error) {

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
func (client *consulCenter) Deregister(serviceId string) (err error) {
	return client.Agent().ServiceDeregister(serviceId)
}
func (client *consulCenter) Discovery(name string) ([]*Service, error) {
	var entry *cacheEntry
	now := time.Now().Unix()
	if tmp, ok := client.cache.Load(name); ok {
		if entry, ok = tmp.(*cacheEntry); ok {
			if now-entry.ts < client.ttl {
				return entry.vl, nil
			}
		}
	}

	entries, metainfo, err := client.Health().Service(name, "", true, &api.QueryOptions{
		WaitIndex: client.lastIndex,
	})
	if err != nil {
		return nil, err
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
	client.lastIndex = metainfo.LastIndex
	if entry == nil {
		entry.vl = services
		entry.ts = now
	} else {
		client.cache.Store(name, &cacheEntry{
			vl: services,
			ts: now,
		})
	}
	return services, nil
}

func (client *consulCenter) Robin(name string) (*Service, error) {
	services, err := Discovery(name)
	if err != nil {
		return nil, err
	}
	size := uint32(len(services))
	if size == 0 {
		return nil, nil
	}
	client.robin++
	return services[client.robin%size], nil
}
func (client *consulCenter) Hash(name string, key string) (*Service, error) {
	services, err := Discovery(name)
	if err != nil {
		return nil, err
	}
	size := uint32(len(services))
	if size == 0 {
		return nil, nil
	}
	idx := mmhash([]byte(key))
	return services[idx%size], nil
}
