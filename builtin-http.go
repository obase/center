package center

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"github.com/obase/conf"
	"io"
	"net"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"time"
)

const (
	REVERSE_SCHEME = "rx-scheme"
	REVERSE_HOST   = "rx-host"
	REVERSE_PATH   = "rx-path"

	ProxyBufferPool_None = "none" // 没有缓存池
	ProxyBufferPool_Sync = "sync" // 采用sync.Pool

	ProxyErrorHandler_None = "none" // 没有错误处理
	ProxyErrorHandler_Body = "body" // 将错误写到body

	HTTP_CKEY = "center.http"
)

var (
	defaultConfig       *HttpConfig
	defaultTransport    *http.Transport
	defaultClient       *http.Client
	defaultReverseProxy *httputil.ReverseProxy
)

func init() {
	var config *HttpConfig
	conf.Bind(HTTP_CKEY, &config)
	SetupHttp(config)
}

type HttpConfig struct {
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete. If Deadline is also set, it may fail
	// earlier.
	//
	// The default is no timeout.
	//
	// When using TCP and dialing a host name with multiple IP
	// addresses, the timeout may be divided between them.
	//
	// With or without a timeout, the operating system may impose
	// its own earlier timeout. For instance, TCP timeouts are
	// often around 3 minutes.
	ConnectTimeout time.Duration `json:"connectTimeout" bson:"connectTimeout" yaml:"connectTimeout"`

	// KeepAlive specifies the keep-alive period for an active
	// network connection.
	// If zero, keep-alives are enabled if supported by the protocol
	// and operating system. Network protocols or operating systems
	// that do not support keep-alives ignore this field.
	// If negative, keep-alives are disabled.
	KeepAlive time.Duration `json:"keepAlive" bson:"keepAlive" yaml:"keepAlive"`

	// MaxIdleConns controls the maximum number of idle (keep-alive)
	// connections across all hosts. Zero means no limit.
	MaxIdleConns int `json:"maxIdleConns" bson:"maxIdleConns" yaml:"maxIdleConns"`

	// MaxIdleConnsPerHost, if non-zero, controls the maximum idle
	// (keep-alive) connections to keep per-host. If zero,
	// DefaultMaxIdleConnsPerHost is used.
	MaxIdleConnsPerHost int `json:"maxIdleConnsPerHost" bson:"maxIdleConnsPerHost" yaml:"maxIdleConnsPerHost"`

	// MaxConnsPerHost optionally limits the total number of
	// connections per host, including connections in the dialing,
	// active, and idle states. On limit violation, dials will block.
	//
	// Zero means no limit.
	//
	// For HTTP/2, this currently only controls the number of new
	// connections being created at a time, instead of the total
	// number. In practice, hosts using HTTP/2 only have about one
	// idle connection, though.
	MaxConnsPerHost int `json:"maxConnsPerHost" bson:"maxConnsPerHost" yaml:"maxConnsPerHost"`

	// IdleConnTimeout is the maximum amount of time an idle
	// (keep-alive) connection will remain idle before closing
	// itself.
	// Zero means no limit.
	IdleConnTimeout time.Duration `json:"idleConnTimeout" bson:"idleConnTimeout" yaml:"idleConnTimeout"`

	// DisableCompression, if true, prevents the Transport from
	// requesting compression with an "Accept-Encoding: gzip"
	// request header when the Request contains no existing
	// Accept-Encoding value. If the Transport requests gzip on
	// its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body. However, if the user
	// explicitly requested gzip it is not automatically
	// uncompressed.
	DisableCompression bool `json:"disableCompression" bson:"disableCompression" yaml:"disableCompression"`

	// ResponseHeaderTimeout, if non-zero, specifies the amount of
	// time to wait for a server's response headers after fully
	// writing the request (including its body, if any). This
	// time does not include the time to read the response body.
	ResponseHeaderTimeout time.Duration `json:"responseHeaderTimeout" bson:"responseHeaderTimeout" yaml:"responseHeaderTimeout"`

	// ExpectContinueTimeout, if non-zero, specifies the amount of
	// time to wait for a server's first response headers after fully
	// writing the request headers if the request has an
	// "Expect: 100-continue" header. Zero means no timeout and
	// causes the body to be sent immediately, without
	// waiting for the server to approve.
	// This time does not include the time to send the request header.
	ExpectContinueTimeout time.Duration `json:"expectContinueTimeout" bson:"expectContinueTimeout" yaml:"expectContinueTimeout"`

	// MaxResponseHeaderBytes specifies a limit on how many
	// response bytes are allowed in the server's response
	// header.
	//
	// Zero means to use a default limit.
	MaxResponseHeaderBytes int64 `json:"maxResponseHeaderBytes" bson:"maxResponseHeaderBytes" yaml:"maxResponseHeaderBytes"`

	// Timeout specifies a time limit for requests made by this
	// Client. The timeout includes connection time, any
	// redirects, and reading the response body. The timer remains
	// running after Get, Head, Post, or Do return and will
	// interrupt reading of the Response.Body.
	//
	// A Timeout of zero means no timeout.
	//
	// The Client cancels requests to the underlying Transport
	// as if the Request's Context ended.
	//
	// For compatibility, the Client will also use the deprecated
	// CancelRequest method on Transport if found. New
	// RoundTripper implementations should use the Request's Context
	// for cancelation instead of implementing CancelRequest.
	RequestTimeout time.Duration `json:"requestTimeout" bson:"requestTimeout" yaml:"requestTimeout"`

	// FlushInterval specifies the flush interval
	// to flush to the client while copying the
	// response body.
	// If zero, no periodic flushing is done.
	// A negative value means to flush immediately
	// after each write to the client.
	// The FlushInterval is ignored when ReverseProxy
	// recognizes a response as a streaming response;
	// for such responses, writes are flushed to the client
	// immediately.
	ProxyFlushInterval time.Duration `json:"proxyFlushInterval" bson:"proxyFlushInterval" yaml:"proxyFlushInterval"`

	// BufferPool optionally specifies a buffer pool to
	// get byte slices for use by io.CopyBuffer when copying HTTP response bodies.
	// Values: none, sync
	ProxyBufferPool string `json:"proxyBufferPool" bson:"proxyBufferPool" yaml:"proxyBufferPool"`

	// ErrorHandler is an optional function that handles errors
	// reaching the backend or errors from ModifyResponse.
	// Values: none, body
	ProxyErrorHandler string `json:"proxyErrorHandler" bson:"proxyErrorHandler" yaml:"proxyErrorHandler"`
}

func mergeHttpConfig(c *HttpConfig) *HttpConfig {
	if c == nil {
		c = new(HttpConfig)
	}
	if c.ProxyBufferPool == "" {
		c.ProxyBufferPool = ProxyBufferPool_Sync
	}
	if c.ProxyErrorHandler == "" {
		c.ProxyErrorHandler = ProxyErrorHandler_Body
	}
	return c
}

func SetupHttp(hc *HttpConfig) {
	defaultConfig = mergeHttpConfig(hc)

	defaultTransport = &http.Transport{
		Proxy: http.ProxyFromEnvironment,
		DialContext: (&net.Dialer{
			Timeout:   defaultConfig.ConnectTimeout,
			KeepAlive: defaultConfig.KeepAlive,
		}).DialContext,
		MaxIdleConns:           defaultConfig.MaxIdleConns,
		MaxIdleConnsPerHost:    defaultConfig.MaxIdleConnsPerHost,
		MaxConnsPerHost:        defaultConfig.MaxConnsPerHost,
		IdleConnTimeout:        defaultConfig.IdleConnTimeout,
		DisableCompression:     defaultConfig.DisableCompression,
		ResponseHeaderTimeout:  defaultConfig.ResponseHeaderTimeout,
		ExpectContinueTimeout:  defaultConfig.ExpectContinueTimeout,
		MaxResponseHeaderBytes: defaultConfig.MaxResponseHeaderBytes,
	}
	defaultClient = &http.Client{
		Transport: defaultTransport,
		Timeout:   defaultConfig.RequestTimeout,
	}
	defaultReverseProxy = &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			req.URL.Scheme = req.Header.Get(REVERSE_SCHEME)
			req.URL.Host = req.Header.Get(REVERSE_HOST)
			req.URL.Path = req.Header.Get(REVERSE_PATH)
			if _, ok := req.Header["User-Agent"]; !ok {
				// explicitly disable User-Agent so it's not set to default value
				req.Header.Set("User-Agent", "")
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}

type bytesPool struct {
	sync.Pool
}

func (s *bytesPool) Get() []byte {
	return s.Pool.Get().([]byte)
}
func (s *bytesPool) Put(v []byte) {
	s.Pool.Put(v)
}

func proxyBufferPool(name string) httputil.BufferPool {
	switch name {
	case ProxyBufferPool_None:
		return nil
	case ProxyBufferPool_Sync:
		return &bytesPool{
			Pool: sync.Pool{
				New: func() interface{} {
					return make([]byte, 32*1024)
				},
			},
		}
	}
	panic("invalid proxy buffer pool type: " + name)
}

func proxyErrorHandler(name string) func(w http.ResponseWriter, r *http.Request, err error) {
	switch name {
	case ProxyErrorHandler_None:
		return nil
	case ProxyErrorHandler_Body:
		return func(w http.ResponseWriter, r *http.Request, err error) {
			w.WriteHeader(http.StatusBadGateway)
			fmt.Fprintf(w, " proxy error: %v", err)
		}
	}
	panic("invalid proxy error handler type: " + name)
}

var bytesBufferPool = sync.Pool{
	New: func() interface{} {
		return new(bytes.Buffer)
	},
}

func HttpRequest(serviceName string, method string, https bool, uri string, header map[string]string, body io.Reader) (int, string, error) {
	service, err := Robin(serviceName)
	if err != nil {
		return 0, "", err
	}

	var url string
	{
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		if https {
			buf.WriteString("https://")
		} else {
			buf.WriteString("http://")
		}
		buf.WriteString(service.Host)
		buf.WriteByte(':')
		buf.WriteString(strconv.Itoa(service.Port))
		buf.WriteString(uri)
		url = buf.String()
		bytesBufferPool.Put(buf)
	}
	// 创建请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return 0, "", err
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rsp, err := defaultClient.Do(req)
	if err != nil {
		return 0, "", err
	}
	defer rsp.Body.Close()

	var content string
	{
		buf := bytesBufferPool.Get().(*bytes.Buffer)
		_, err = io.Copy(buf, rsp.Body)
		if err != nil {
			bytesBufferPool.Put(buf)
			return 0, "", err
		} else {
			content = buf.String()
			bytesBufferPool.Put(buf)
		}
	}
	return rsp.StatusCode, content, nil
}

func HttpPost(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	var body io.Reader
	if reqobj != nil {
		var data []byte
		data, err = json.Marshal(reqobj)
		if err != nil {
			return
		}
		body = bytes.NewReader(data)
	}
	status, content, err := HttpRequest(serviceName, http.MethodPost, false, uri, header, body)
	if err != nil {
		return
	}
	if status < 200 || status > 299 {
		err = errors.New(content)
	} else {
		err = json.Unmarshal([]byte(content), &rspobj)
	}
	return
}

func HttpsPost(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	var body io.Reader
	if reqobj != nil {
		var data []byte
		data, err = json.Marshal(reqobj)
		if err != nil {
			return
		}
		body = bytes.NewReader(data)
	}
	status, content, err := HttpRequest(serviceName, http.MethodPost, true, uri, header, body)
	if err != nil {
		return
	}
	if status < 200 || status > 299 {
		err = errors.New(content)
	} else {
		err = json.Unmarshal([]byte(content), &rspobj)
	}
	return
}

func HttpProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := Robin(serviceName)
	if service != nil && err == nil {
		request.Header.Set(REVERSE_SCHEME, "http")
		request.Header.Set(REVERSE_HOST, service.Host+":"+strconv.Itoa(service.Port))
		request.Header.Set(REVERSE_PATH, uri)
		defaultReverseProxy.ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusBadGateway)
	}
	return
}

func HttpsProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := Robin(serviceName)
	if service != nil && err == nil {
		request.Header.Set(REVERSE_SCHEME, "https")
		request.Header.Set(REVERSE_HOST, service.Host+":"+strconv.Itoa(service.Port))
		request.Header.Set(REVERSE_PATH, uri)
		defaultReverseProxy.ServeHTTP(writer, request)
	} else {
		writer.WriteHeader(http.StatusBadGateway)
	}
	return
}

func HttpProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			service, _ := Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "http"
				req.URL.Host = service.Host + ":" + strconv.Itoa(service.Port)
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}

func HttpsProxyHandlerTLS(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     defaultTransport,
		FlushInterval: defaultConfig.ProxyFlushInterval,
		Director: func(req *http.Request) {
			service, _ := Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "https"
				req.URL.Host = service.Host + ":" + strconv.Itoa(service.Port)
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   proxyBufferPool(defaultConfig.ProxyBufferPool),
		ErrorHandler: proxyErrorHandler(defaultConfig.ProxyErrorHandler),
	}
}
