package center

import (
	"github.com/obase/kit"
	"io"
	"net/http"
	"net/http/httputil"
)

func HttpRequest(serviceName string, method string, uri string, header map[string]string, body io.Reader) (int, string, error) {
	service, err := Robin(serviceName)
	if err != nil {
		return 0, "", err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("http://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	// 创建请求
	return kit.HttpRequest(method, url, header, body)
}

func HttpsRequest(serviceName string, method string, uri string, header map[string]string, body io.Reader) (int, string, error) {
	service, err := Robin(serviceName)
	if err != nil {
		return 0, "", err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("https://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	// 创建请求
	return kit.HttpRequest(method, url, header, body)
}

func HttpJson(serviceName string, method string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return 0, err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("http://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	return kit.HttpJson(method, url, header, reqobj, rspobj)
}

func HttpsJson(serviceName string, method string, uri string, header map[string]string, reqobj interface{}, rspobj interface{}) (status int, err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return 0, err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("https://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	return kit.HttpJson(method, url, header, reqobj, rspobj)
}

// 本方法纯粹为了兼容旧思路
func Post(serviceName string, uri string, param interface{}, result interface{}) error {
	service, err := Robin(serviceName)
	if err != nil {
		return err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("http://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	_, err = kit.HttpJson(http.MethodPost, url, nil, param, result)
	return err
}

func HttpProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("http://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	return kit.HttpProxy(url, writer, request)
}

func HttpsProxy(serviceName string, uri string, writer http.ResponseWriter, request *http.Request) (err error) {
	service, err := Robin(serviceName)
	if err != nil {
		return err
	}

	var url string
	{
		buf := kit.BorrowBuffer()
		buf.WriteString("https://")
		buf.WriteString(service.Addr)
		buf.WriteString(uri)
		url = buf.String()
		kit.ReturnBuffer(buf)
	}
	return kit.HttpProxy(url, writer, request)
}

func HttpProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     kit.DefaultReverseProxy.Transport,
		FlushInterval: kit.DefaultReverseProxy.FlushInterval,
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
		BufferPool:   kit.DefaultReverseProxy.BufferPool,
		ErrorHandler: kit.DefaultReverseProxy.ErrorHandler,
	}
}

func HttpsProxyHandler(serviceName string, uri string) *httputil.ReverseProxy {
	return &httputil.ReverseProxy{
		Transport:     kit.DefaultReverseProxy.Transport,
		FlushInterval: kit.DefaultReverseProxy.FlushInterval,
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
		BufferPool:   kit.DefaultReverseProxy.BufferPool,
		ErrorHandler: kit.DefaultReverseProxy.ErrorHandler,
	}
}
