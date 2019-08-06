# package grpcx
grpc客户端扩展. 提透明容错, 负载均衡等高可用特性

依赖:
```
"google.golang.org/grpc"
"google.golang.org/grpc/naming"
```

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
grpcx:
  # 空询稍等时间
  updateSleepDuration: "1s"
```

# Index
- func Dial
```
func Dial(serviceName string) (*grpc.ClientConn, error) 
```
返回grpc客户端连接, 各参数意义:
```
- serviceName: 注册中心的服务名称

```
# Examples
```
    
package main

import (
	"context"
	"fmt"
	"github.com/obase/apix"
	"github.com/obase/demo/api"
	"github.com/obase/log"
	"google.golang.org/grpc"
	"os"
	"strconv"
)

func main() {
	cc, err := grpcx.Dial("pvpbroker")
	if err != nil {
		log.Errorf(context.Background(), "dial error: %v", err)
		os.Exit(1)
	}
	defer cc.Close()

	cl := api.NewIPlayerClient(cc)
	for i := 0; i < 10; i++ {
		_, err = cl.Add(context.Background(), &api.Player{
			Id:   "ID_" + strconv.Itoa(i),
			Name: "Tomcat",
		})
		if err != nil {
			panic(err)
		}
	}
	lst, err := cl.List(context.Background(), apix.None)
	if err != nil {
		panic(err)
	}
	for i, v := range lst.Players {
		fmt.Printf("%v : %v\n", i, v)
	}
}

```