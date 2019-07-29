package center

import (
	"errors"
	"github.com/obase/log"
	"time"
)

var ErrInvalidClient = errors.New("invalid consul client")

type Option struct {
	Address string        // 远程地址
	Timeout time.Duration // 本地缓存过期时间
}

// 根据consul的服务项设计
type Check struct {
	Type     string `json:"type,omitempty"`
	Target   string `json:"target,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Interval string `json:"interval,omitempty"`
}

type Service struct {
	Id   string `json:"id,omitempty"` // 如果为空则自动生成
	Kind string `json:"kind,omitempty"`
	Name string `json:"name,omitempty"`
	Host string `json:"host,omitempty"`
	Port int    `json:"port,omitempty"`
}

type Center interface {
	Register(service *Service, check *Check) (err error)
	Deregister(serviceId string) (err error)
	Discovery(name string) ([]*Service, error)
}

var Default Center

func Setup(opt *Option) {
	Default = newConsulCenter(opt.Address, opt.Timeout)
}

func Register(service *Service, check *Check) (err error) {
	if Default == nil {
		log.Warnf(nil, "Register service failed: invalid consul client, %v", service)
		log.Flushf()
		return ErrInvalidClient
	}
	return Default.Register(service, check)
}
func Deregister(serviceId string) (err error) {
	if Default != nil {
		log.Warnf(nil, "Deregister service failed: invalid consul client, %v", serviceId)
		log.Flushf()
		return ErrInvalidClient
	}
	return Default.Deregister(serviceId)
}
func Discovery(name string) ([]*Service, error) {
	if Default == nil {
		log.Warnf(nil, "Disconvery service failed: invalid consul client, %v", name)
		return nil, ErrInvalidClient
	}
	return Default.Discovery(name)
}
