package center

import (
	"net"
	"strconv"
)

type configClient struct {
	entries map[string][]*Service
	robin   uint32
}

func newConfigClient(cfs map[string][]string) Center {
	ret := &configClient{
		entries: make(map[string][]*Service),
	}
	for k, v := range cfs {
		ss := make([]*Service, len(v))
		for i, s := range v {
			sv := new(Service)
			h, p, _ := net.SplitHostPort(s)
			sv.Host = h
			sv.Port, _ = strconv.Atoi(p)
			ss[i] = sv
		}
		ret.entries[k] = ss
	}
	return ret
}

func (c *configClient) Register(service *Service, check *Check) (err error) {
	return
}
func (c *configClient) Deregister(serviceId string) (err error) {
	return
}
func (c *configClient) Discovery(name string) ([]*Service, error) {
	return c.entries[name], nil
}
