package httpx

import (
	"fmt"
	"github.com/obase/conf"
	"net/http"
	"net/http/httputil"
	"sync"
	"time"
)

type Config struct {
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

const (
	ProxyBufferPool_None = "none" // 没有缓存池
	ProxyBufferPool_Sync = "sync" // 采用sync.Pool

	ProxyErrorHandler_None = "none" // 没有错误处理
	ProxyErrorHandler_Body = "body" // 将错误写到body
)

const CKEY = "httpx"

func LoadConfig() *Config {
	var config *Config
	if ok := conf.Scan(CKEY, &config); !ok {
		return nil
	}
	return config
}

func mergeConfig(c *Config) *Config {
	if c == nil {
		c = new(Config)
	}
	if c.ProxyBufferPool == "" {
		c.ProxyBufferPool = ProxyBufferPool_None
	}
	if c.ProxyErrorHandler == "" {
		c.ProxyErrorHandler = ProxyErrorHandler_None
	}
	return c
}

type syncPool struct {
	*sync.Pool
}

func newSyncPool() *syncPool {
	return &syncPool{
		Pool: &sync.Pool{
			New: func() interface{} {
				return make([]byte, 32*1024)
			},
		},
	}
}

func (s *syncPool) Get() []byte {
	return s.Pool.Get().([]byte)
}
func (s *syncPool) Put(v []byte) {
	s.Pool.Put(v)
}

func proxyBufferPool(name string) httputil.BufferPool {
	switch name {
	case ProxyBufferPool_None:
		return nil
	case ProxyBufferPool_Sync:
		return newSyncPool()
	}
	panic("invalid proxy buffer pool type: " + name)
}

func bodyErrorHandler(w http.ResponseWriter, r *http.Request, err error) {
	fmt.Fprintf(w, " proxy error: %v", err)
	w.WriteHeader(http.StatusBadGateway)
}

func proxyErrorHandler(name string) func(w http.ResponseWriter, r *http.Request, err error) {
	switch name {
	case ProxyErrorHandler_None:
		return nil
	case ProxyErrorHandler_Body:
		return bodyErrorHandler
	}
	panic("invalid proxy error handler type: " + name)
}
