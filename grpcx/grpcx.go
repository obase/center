package grpcx

import (
	"github.com/obase/center"
	"github.com/obase/conf"
	"google.golang.org/grpc"
	"google.golang.org/grpc/naming"
	"strconv"
	"time"
)

var updateSleepDuration = conf.OptiDuration("grpcx.updateSleepDuration", time.Second)

type serviceWatcher struct {
	serviceName string
	address     map[string]bool
}

func (w *serviceWatcher) Next() ([]*naming.Update, error) {
	for {
		services, err := center.Discovery(w.serviceName)
		if err != nil {
			return nil, err
		}

		address := make(map[string]bool)
		for _, service := range services {
			address[service.Host+":"+strconv.Itoa(service.Port)] = true
		}

		var updates []*naming.Update
		for addr := range w.address {
			if _, ok := address[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Delete, Addr: addr})
			}
		}

		for addr := range address {
			if _, ok := w.address[addr]; !ok {
				updates = append(updates, &naming.Update{Op: naming.Add, Addr: addr})
			}
		}

		if len(updates) != 0 {
			w.address = address
			return updates, nil
		}
		time.Sleep(updateSleepDuration)
	}
}

func (w *serviceWatcher) Close() {
	// nothing to do
}

func (r *serviceWatcher) Resolve(target string) (naming.Watcher, error) {
	return r, nil
}

func Dial(serviceName string) (*grpc.ClientConn, error) {
	return grpc.Dial("", grpc.WithInsecure(), grpc.WithBlock(), grpc.WithBalancer(
		grpc.RoundRobin(&serviceWatcher{
			serviceName: serviceName,
		})))
}
