# package center
服务注册中心客户端, 目前只支持consul. 
配置consul的命令如下:

## master
```
consul agent \
-server \
-bootstrap-expect 1 -ui \
-bind 10.11.165.44 \
-dns-port 18600 \
-http-port 18500 \
-serf-lan-port 18301 \
-serf-wan-port 18302 \
-server-port 18300 \
-data-dir /data/consul/data18500 \
-log-file /data/consul/logs/18500 \
-pid-file /data/consul/pids/18500 \
-node 18500 \
-datacenter jx3 \
-client 0.0.0.0 &
```

## slave
```
consul agent \
-server \
-retry-join 10.11.165.44:18301 \
-bind 10.11.165.44 \
-dns-port 28600 \
-http-port 28500 \
-serf-lan-port 28301 \
-serf-wan-port 28302 \
-server-port 28300 \
-data-dir /data/consul/data28500 \
-log-file /data/consul/logs/28500 \
-pid-file /data/consul/pids/28500 \
-node 28500 \
-datacenter jx3 \
-client 0.0.0.0 &
```

## slave
```
consul agent \
-server \
-retry-join 10.11.165.44:18301 \
-bind 10.11.165.44 \
-dns-port 38600 \
-http-port 38500 \
-serf-lan-port 38301 \
-serf-wan-port 38302 \
-server-port 38300 \
-data-dir /data/consul/data38500 \
-log-file /data/consul/logs/38500 \
-pid-file /data/consul/pids/38500 \
-node 38500 \
-datacenter jx3 \
-client 0.0.0.0 &
```

## client
```
consul agent -bind 10.11.163.127 -retry-join 10.11.165.44:18301 -data-dir F:\consul -datacenter jx3 -node c1 -config-file F:\consul.json

----consul.json----
{
"disable_update_check": true
}
```

注意: windows下不能用"*.js"后缀

# Installation
- go get
```
go get -u github.com/obase/center
```
- go mod
```
go mod edit -require=github.com/obase/center@latest
```

# Configuration

conf.yml
```
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center: "127.0.0.1:8500"

或者定制缓存超时
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center:
    address: "127.0.0.1:8500"
    timeout: "1m"

或者设置本地缓存
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center:
    configs:
      pvpbroker: ["172.31.0.5:9000","172.31.0.19:9000"]

或者禁止center
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center: off
```
如无配置,默认127.0.0.1:8500. 使用off可以关闭center功能! 另外, 兼容旧的828配置
```
ext:
  service:
    centerAddr: 
```
# Constants

# Variables
```
var ErrInvalidClient = errors.New("invalid consul client")
```
# Index
- type Config
```
type Config struct {
	Address string              // 远程地址
	Expired int64               // 本地缓存过期秒数
	Configs map[string][]string // 本地配置address
}

```

- type Check 
```
type Check struct {
	Type     string `json:"type,omitempty"`
	Target   string `json:"target,omitempty"`
	Timeout  string `json:"timeout,omitempty"`
	Interval string `json:"interval,omitempty"`
}
```

- type Center 
```
type Center interface {
	Register(service *Service, check *Check) (err error)
	Deregister(serviceId string) (err error)
	FetchService(name string) ([]*Service, uint64, error)               // 有缓存
	WatchService(name string, index uint64) ([]*Service, uint64, error) // 无缓存
}

```
- func Register
```
func Register(service *Service, check *Check) (err error) 
```
注册服务及心跳检查

- func Deregister
```
func Deregister(serviceId string) (err error) 
```
反注册服务

- func FetchService
```
func FetchService(name string) ([]*Service, uint64, error) 
```
获取服务(非阻塞)

- func WatchService
```
func WatchService(name string, index uint64) ([]*Service, uint64, error) 
```
获取服务(阻塞直到index后有更新). 同时返回最新的index

- func Robin
```
func Robin(name string) (*Service, error) 
```
轮询返回服务地址

- func Hash
``` 
func Hash(name string, key string) (*Service, error)
```
根据key哈希值返回固定服务. 所用算法是murmurhash32

- func HttpName
```
func HttpName(name string) string
```
获取name对应的http name

- func GrpcName
```
func GrpcName(name string) string
```
获取name对应的grpc name

- func TcpName
```
func TcpName(name string) string
```
获取name对应的tcp name

# Examples

conf.yml
```
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center: "127.0.0.1:8500"

或者定制缓存超时
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center:
    address: "127.0.0.1:8500"
    expired: 60

或者设置本地缓存
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center:
    configs:
      pvpbroker: ["172.31.0.5:9000","172.31.0.19:9000"]

或者禁止center
service:
  # consul注册中心,默认127.0.0.1:8500. 不设置则忽略!
  center: off
```
codes:
```
func TestDiscovery(t *testing.T) {
	ss, err := Discovery("pvpbroker")
	if err != nil {
		t.Fatal(err)
	}
	fmt.Println(ss)
}

```