# package httpx
http扩展客户端

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
# http扩展配置
httpx:
  # 连接超时, 默认30秒
  connectTimeout: "30s"
  # 连接keepalive, 默认30秒
  keepAlive: "30s"
  # 最大空闲,默认10240
  maxIdleConns: 10240
  # 每个主机最大连接数, 该值直接影响并发QPS
  maxIdleConnsPerHost: 2048
  # 每机最大连接数
  maxConnsPerHost: 0
  # 空闲超时, 默认90秒
  idleConnTimeout: "90s"
  # 是否禁用压缩
  disableCompression: false
  # 响应头超时, 默认5秒
  responseHeaderTimeout: "5s"
  # 期望再超时, 默认1秒
  expectContinueTimeout: "1s"
  # 最大响应大字节数
  maxResponseHeaderBytes: 0
  # 请求超时.默认60秒
  requestTimeout: "60s"
  # 反向代理刷新间隔, 0表示默认, 负表示立即刷新
  proxyFlushInterval: 0
  # 反向代理Buff池策略, none表示没有,sync表示用sync.Pool
  proxyBufferPool: "none"
  # 反向代理错误解句柄, none表示没有,body表示将错误写在响应内容体
  proxyErrorHandler: "none"

```
# Index
- Constants
- Variables
- func Request
```
func Request(method string, serviceName string, uri string, header map[string]string, body io.Reader) (state int, content string, err error)
```
各参数意义:
```
- method: 请求方法, 例如GET, POST, PUT, DELETE, HEAD等
- serviceName: 注册中心的服务名称
- uri: 请求资源名称
- header: 请求头
- body: 请求体内容
- state: 响应状态码, 一般200~299是合法返回.
- content: 响应内容
- err: 响应错误
```

- func Post
```
func Post(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) 
```
POST+JSON请求快捷方法, 各参数意义:
```
- serviceName: 注册中心的服务名称
- uri: 请求资源名称 
- header: 请求头
- reqobj: 请求体对象, 自动序列为JSON数据
- rspobj: 响应体对象指针(必须是指针) 自动反序列化JSON对象. 只有status为200~299才会结果对象. 
- status: 响应状态码
- err: 响应错误. 当status不是200~299则将响应内容作为错误内容.
```

- func Proxy
```
func Proxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error)
```
代理转发HTTP请求, 典型应用场景是反向代理网关. 各参数意义:
```
- serviceName: 注册中心的服务名称
- uri: 请求资源名称 
- writer: 原始请求Writer
- request: 原始请求Request
- err: 代理错误
```

- func ProxyTLS
```
func ProxyTLS(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error)
```
代理转发HTTPS请求, 典型应用场景是反向代理网关. 各参数意义:
```
- serviceName: 注册中心的服务名称
- uri: 请求资源名称 
- writer: 原始请求Writer
- request: 原始请求Request
- err: 代理错误
```

- func ProxyHandler
```
func ProxyHandler(serviceName string, uri string) *httputil.ReverseProxy
```
代理转发HTTP请求

- func ProxyHandlerTLS
```
func ProxyHandlerTLS(serviceName string, uri string) *httputil.ReverseProxy
```
代理转发HTTPS请求

# Example