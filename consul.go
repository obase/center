package consul

import (
	"context"
	"errors"
	"github.com/hashicorp/consul/api"
	"github.com/obase/conf"
	"github.com/obase/log"
)

var (
	client           *api.Client
	ErrInvalidClient = errors.New("invalid consul client")
)

const (
	KEY_AGENT     = "service.consulAgent"
	KEY_AGENT_828 = "ext.service.centerAddr"
)

func init() {

	consulAddress, ok := conf.GetString(KEY_AGENT)
	if !ok {
		if consulAddress, ok = conf.GetString(KEY_AGENT_828); !ok {
			return
		}
	}

	config := api.DefaultConfig()
	if consulAddress != "" {
		config.Address = consulAddress
	}
	var err error
	if client, err = api.NewClient(config); err != nil { // 兼容旧的逻辑
		log.Errorf(context.Background(), "Connect consul error: %s, %v", consulAddress, err)
		log.Flushf()
	} else {
		if _, err = client.Agent().Services(); err != nil {
			log.Errorf(context.Background(), "Connect consul error: %s, %v", consulAddress, err)
			log.Flushf()
		} else {
			log.Inforf(context.Background(), "Connect consul success: %s", consulAddress)
			log.Flushf()
		}
	}

}

func RegisterService(service *api.AgentServiceRegistration) (err error) {
	if client != nil {
		log.Warnf(nil, "Register service failed: invalid consul client, %v", service)
		log.Flushf()
		return ErrInvalidClient
	}
	if err = client.Agent().ServiceRegister(service); err != nil {
		log.Errorf(nil, "Register service error: %v, %v", service, err)
		log.Flushf()
	} else {
		log.Inforf(nil, "Register service success: %v", service)
		log.Flushf()
	}
	return
}

func DeregisterService(serviceId string) (err error) {
	if client == nil {
		log.Warnf(nil, "Deregister service failed: invalid consul client, %v", serviceId)
		log.Flushf()
		return ErrInvalidClient
	}
	if err = client.Agent().ServiceDeregister(serviceId); err != nil {
		log.Errorf(context.Background(), "Deregister service error: %v, %v", serviceId, err)
		log.Flushf()
	} else {
		log.Inforf(context.Background(), "Deregister service success: %v", serviceId)
		log.Flushf()
	}
	return
}

func DiscoveryService(lastIndex uint64, serviceId string) ([]*api.ServiceEntry, *api.QueryMeta, error) {
	if client == nil {
		log.Warnf(nil, "Disconvery service failed: invalid consul client, %v", serviceId)
		return nil, nil, ErrInvalidClient
	}
	return client.Health().Service(serviceId, "", true, &api.QueryOptions{
		WaitIndex: lastIndex,
	})
}
