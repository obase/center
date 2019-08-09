package center

import (
	"net"
	"strconv"
)

type localClient struct {
	entries map[string][]*Service
	robin   uint32
}

func newLocalClient(cfs map[string][]string) Center {
	ret := &localClient{
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

func (c *localClient) Register(service *Service, check *Check) (err error) {
	return
}
func (c *localClient) Deregister(serviceId string) (err error) {
	return
}
func (c *localClient) FetchService(name string) ([]*Service, uint64, error) {
	return c.entries[name], 0, nil
}
func (c *localClient) WatchService(name string, index uint64) ([]*Service, uint64, error) {
	return c.entries[name], 0, nil
}
