package center

import (
	"github.com/obase/kit"
	"strings"
)

var robin uint

func Robin(name string) (*Service, error) {
	services, _, err := FetchService(name)
	if err != nil {
		return nil, err
	}
	size := len(services)
	if size == 0 {
		return nil, nil
	}
	robin++
	return services[robin%uint(size)], nil
}
func Hash(name string, key string) (*Service, error) {
	services, _, err := FetchService(name)
	if err != nil {
		return nil, err
	}
	size := len(services)
	if size == 0 {
		return nil, nil
	}
	idx := kit.MMHash32([]byte(key))
	return services[idx%uint32(size)], nil
}

func HttpName(name string) string {
	if strings.HasSuffix(name, ".http") {
		return name
	}
	return name + ".http"
}

func GrpcName(name string) string {
	if strings.HasSuffix(name, ".grpc") {
		return name
	}
	return name + ".grpc"
}
