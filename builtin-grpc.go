package center

import (
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"strconv"
)

type serviceWatcher struct {
	Name  string
	Index uint64
	Cache map[string]interface{}
}

// 阻塞式访问
func (w *serviceWatcher) Next() ([]*naming.Update, error) {
	for {
		services, index, err := WatchService(w.Name, w.Index)
		if err != nil {
			return nil, err
		}
		w.Index = index

		cache := make(map[string]interface{})
		for _, service := range services {
			cache[service.Host+":"+strconv.Itoa(service.Port)] = nil
		}

		var updates []*naming.Update
		for addr := range w.Cache {
			if _, ok := cache[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
			}
		}

		for addr := range cache {
			if _, ok := w.Cache[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
			}
		}

		if len(updates) != 0 {
			w.Cache = cache
			return updates, nil
		}
	}
}

func (w *serviceWatcher) Close() {
	// nothing to do
}

func (r *serviceWatcher) Resolve(target string) (naming.Watcher, error) {
	return r, nil
}

func GrpcDial(serviceName string) (*grpc.ClientConn, error) {
	return grpc.Dial("", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithBalancer(
		grpc.RoundRobin(&serviceWatcher{
			Name: serviceName,
		})))
}
