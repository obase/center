package center

import (
	"errors"
	"time"
)

var ErrInvalidClient = errors.New("invalid consul client")

type Config struct {
	Address string              // 远程地址
	Timeout time.Duration       // 本地缓存过期时间
	Configs map[string][]string // 本地配置address
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

var instance Center

func Setup(opt *Config) {
	if len(opt.Configs) > 0 {
		instance = newConfigClient(opt.Configs)
	} else {
		instance = newConsulCenter(opt)
	}
}

func Register(service *Service, check *Check) (err error) {
	if instance == nil {
		return ErrInvalidClient
	}
	return instance.Register(service, check)
}
func Deregister(serviceId string) (err error) {
	if instance != nil {
		return ErrInvalidClient
	}
	return instance.Deregister(serviceId)
}
func Discovery(name string) ([]*Service, error) {
	if instance == nil {
		return nil, ErrInvalidClient
	}
	return instance.Discovery(name)
}
