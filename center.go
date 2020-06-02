package center

import (
	"errors"
	"github.com/obase/conf"
)

var ErrInvalidClient = errors.New("invalid consul client")

const (
	DEFAULT_REFRESH = 8 // 最多起8个协程处理后台更新
	LOCAL           = "local"
)

const CKEY = "center"

type Config struct {
	Address string              // 远程地址
	Service map[string][]string // 本地配置服务
	Expired int64               // 缓存过期时间(秒)
	Refresh int                 // 并发刷新协程数量
}

func init() {
	config, ok := conf.Get(CKEY)
	if !ok {
		return
	}
	switch config := config.(type) {
	case nil:
		Setup(&Config{Address: LOCAL})
	case string:
		Setup(&Config{Address: config})
	case map[interface{}]interface{}:
		var service map[string][]string
		address, ok := conf.ElemString(config, "address")
		expired, ok := conf.ElemInt64(config, "expired")
		tmp, ok := conf.ElemMap(config, "service")
		if ok {
			service = make(map[string][]string)
			for k, v := range tmp {
				service[k] = conf.ToStringSlice(v)
			}
		}
		refresh, ok := conf.ElemInt(config, "refresh")
		Setup(&Config{
			Address: address,
			Expired: expired,
			Service: service,
			Refresh: refresh,
		})
	}
}

// 根据consul的服务项设计
type Check struct {
	Type     string
	Target   string
	Timeout  string
	Interval string
}

type Service struct {
	Id   string
	Kind string
	Name string
	Host string
	Port int
	Addr string // 是host:port,避免反复拼接
}

type Center interface {
	Register(service *Service, check *Check) (err error)
	Deregister(serviceId string) (err error)
	FetchService(name string) ([]*Service, uint64, error)               // 有缓存
	WatchService(name string, index uint64) ([]*Service, uint64, error) // 无缓存
}

var instance Center

func Setup(c *Config) {

	if c == nil {
		c = new(Config)
	}
	if c.Refresh == 0 {
		c.Refresh = DEFAULT_REFRESH
	}

	if len(c.Service) > 0 {
		instance = newLocalClient(c.Service)
	} else {
		instance = newConsulClient(c)
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
