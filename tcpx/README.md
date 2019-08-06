# package grpcx
tcp客户端扩展.

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
```

```

# Index
- func Dial
```
func Dial(serviceName string) (tcon *net.TCPConn, err error) 
```
返回tcp客户端连接, 各参数意义:
```
- serviceName: 注册中心的服务名称

```
# Examples
```

```