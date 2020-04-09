package center

import (
	"errors"
)

var ErrInvalidClient = errors.New("invalid consul client")

type Config struct {
	Address string              // 远程地址
	Expired int64               // 缓存过期时间(秒)
	Service map[string][]string // 本地配置服务
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
	FetchService(name string) ([]*Service, uint64, error)               // 有缓存
	WatchService(name string, index uint64) ([]*Service, uint64, error) // 无缓存
}

var instance Center

func Setup(opt *Config) {
	if len(opt.Service) > 0 {
		instance = newLocalClient(opt.Service)
	} else {
		instance = newConsulClient(opt)
	}
}

func Register(service *Service, check *Check) (err error) {
	if instance == nil {
		return ErrInvalidClient
	}
	return instance.Register(service, check)
}
func Deregister(serviceId string) (err error) {
	if instance == nil {
		return ErrInvalidClient
	}
	return instance.Deregister(serviceId)
}
func FetchService(name string) ([]*Service, uint64, error) {
	if instance == nil {
		return nil, 0, ErrInvalidClient
	}
	return instance.FetchService(name)
}

func WatchService(name string, index uint64) ([]*Service, uint64, error) {
	if instance == nil {
		return nil, 0, ErrInvalidClient
	}
	return instance.WatchService(name, index)
}
