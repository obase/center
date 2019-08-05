package httpx

import (
	"bytes"
	"encoding/json"
	"errors"
	"github.com/obase/center"
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
	X_PROXY_SCHEME = "x-proxy-scheme"
	X_PROXY_HOST   = "x-proxy-host"
	X_PROXY_PATH   = "x-proxy-path"
)

var defaultTransport = &http.Transport{
	Proxy: http.ProxyFromEnvironment,
	DialContext: (&net.Dialer{
		Timeout:   conf.OptiDuration("httpx.connectionTimeout", 30*time.Second),
		KeepAlive: conf.OptiDuration("httpx.connectionKeepalive", 30*time.Second),
	}).DialContext,
	MaxIdleConns:          conf.OptiInt("httpx.maxIdleConns", 10240),
	IdleConnTimeout:       conf.OptiDuration("httpx.idleConnTimeout", 90*time.Second),
	TLSHandshakeTimeout:   conf.OptiDuration("httpx.tlsHandshakeTimeout", 10*time.Second),
	ExpectContinueTimeout: conf.OptiDuration("httpx.expectContinueTimeout", 1*time.Second),
	MaxIdleConnsPerHost:   conf.OptiInt("httpx.maxIdleConnsPerHost", 2048),
	ResponseHeaderTimeout: conf.OptiDuration("httpx.responseHeaderTimeout", 5*time.Second),
}

// 基于828的旧参数
var defaultClient = &http.Client{
	Transport: defaultTransport,
	Timeout:   conf.OptiDuration("http.requestTimeout", 60*time.Second),
}

var defaultReverseProxy = &httputil.ReverseProxy{
	Transport: defaultTransport,
	Director: func(req *http.Request) {
		req.URL.Scheme = req.Header.Get(X_PROXY_SCHEME)
		req.URL.Host = req.Header.Get(X_PROXY_HOST)
		req.URL.Path = req.Header.Get(X_PROXY_PATH)
		if _, ok := req.Header["User-Agent"]; !ok {
			// explicitly disable User-Agent so it's not set to default value
			req.Header.Set("User-Agent", "")
		}
	},
}

var buffpool = &sync.Pool{
	New: func() interface{} {
		return bytes.NewBuffer(make([]byte, 256))
	},
}

func Request(method string, serviceName string, uri string, header map[string]string, body io.Reader) (state int, content string, err error) {
	service, err := center.Robin(serviceName)
	if err != nil {
		return
	}
	// 提早分配buf,减少内存分配
	buf := buffpool.Get().(*bytes.Buffer)
	buf.WriteString("http://")
	buf.WriteString(service.Host)
	buf.WriteString(":")
	buf.WriteString(strconv.Itoa(service.Port))
	buf.WriteString(uri)
	url := buf.String()
	buffpool.Put(buf)

	// 创建请求
	req, err := http.NewRequest(method, url, body)
	if err != nil {
		return
	}
	req.Header.Set("Content-Type", "application/json")
	for k, v := range header {
		req.Header.Set(k, v)
	}
	rsp, err := defaultClient.Do(req)
	if err != nil {
		return
	}
	defer rsp.Body.Close()

	buf.Reset()
	_, err = io.Copy(buf, rsp.Body)
	if err != nil {
		return
	}
	return rsp.StatusCode, buf.String(), nil
}

func Post(serviceName string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	data, err := json.Marshal(reqobj)
	if err != nil {
		return
	}
	status, content, err := Request(http.MethodPost, serviceName, uri, header, bytes.NewBuffer(data))
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

func Proxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := center.Robin(serviceName)
	if err != nil {
		return
	}
	request.Header.Set(X_PROXY_SCHEME, "http")
	request.Header.Set(X_PROXY_HOST, service.Host+":"+strconv.Itoa(service.Port))
	request.Header.Set(X_PROXY_PATH, uri)
	defaultReverseProxy.ServeHTTP(writer, request)
	return
}
