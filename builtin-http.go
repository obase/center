package center

import (
	"github.com/obase/kit"
	"io"
	"net/http"
	"net/http/httputil"
)

func HttpRequest(serviceName string, method string, uri string, header map[string]string, body io.Reader) (status int, content string, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("http://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	status, content, err = kit.HttpRequest(method, buf.UnsafeString(), header, body)
	kit.PutStringBuffer(buf)
	return
}

func HttpsRequest(serviceName string, method string, uri string, header map[string]string, body io.Reader) (status int, content string, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("https://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	status, content, err = kit.HttpRequest(method, buf.UnsafeString(), header, body)
	kit.PutStringBuffer(buf)
	return
}

func HttpJson(serviceName string, method string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("http://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	status, err = kit.HttpJson(method, buf.UnsafeString(), header, reqobj, rspobj)
	kit.PutStringBuffer(buf)
	return
}

func HttpsJson(serviceName string, method string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("https://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	status, err = kit.HttpJson(method, buf.UnsafeString(), header, reqobj, rspobj)
	kit.PutStringBuffer(buf)
	return
}

// 本方法纯粹为了兼容旧思路
func Post(serviceName string, uri string, param interface{}, result interface{}) (err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("https://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	_, err = kit.HttpJson(http.MethodPost, buf.UnsafeString(), nil, param, result)
	kit.PutStringBuffer(buf)
	return
}

func HttpProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {

	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("http://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	err = kit.HttpProxy(buf.UnsafeString(), writer, request)
	kit.PutStringBuffer(buf)
	return
}

func HttpsProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return
	}
	buf := kit.GetStringBuffer()
	buf.WriteString("https://")
	buf.WriteString(service.Addr)
	buf.WriteString(uri)
	err = kit.HttpProxy(buf.UnsafeString(), writer, request)
	kit.PutStringBuffer(buf)
	return
}

func HttpProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     kit.ReverseProxy.Transport,
		FlushInterval: kit.ReverseProxy.FlushInterval,
		Director: func(req *http.Request) {
			service, _ := Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "http"
				req.URL.Host = service.Addr
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   kit.ReverseProxy.BufferPool,
		ErrorHandler: kit.ReverseProxy.ErrorHandler,
	}
}

func HttpsProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     kit.ReverseProxy.Transport,
		FlushInterval: kit.ReverseProxy.FlushInterval,
		Director: func(req *http.Request) {
			service, _ := Robin(serviceName)
			if service != nil {
				req.URL.Scheme = "https"
				req.URL.Host = service.Addr
				req.URL.Path = uri
				if _, ok := req.Header["User-Agent"]; !ok {
					// explicitly disable User-Agent so it's not set to default value
					req.Header.Set("User-Agent", "")
				}
			}
		},
		BufferPool:   kit.ReverseProxy.BufferPool,
		ErrorHandler: kit.ReverseProxy.ErrorHandler,
	}
}
