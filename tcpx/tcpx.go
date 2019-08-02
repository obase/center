package tcpx

import (
	"github.com/obase/center"
	"github.com/obase/conf"
	"net"
	"strconv"
	"time"
)

var (
	connectionTimeout = conf.OptiDuration("tcpx.connectionTimeout", 20*time.Second)
	keepalivePeriod   = conf.OptiDuration("tcpx.keepalivePeriod", 0)
)

func Dial(serviceName string) (tcon *net.TCPConn, err error) {
	service, err := center.Robin(serviceName)
	if err != nil {
		return
	}
	con, err := net.DialTimeout("tcp", service.Host+":"+strconv.Itoa(service.Port), connectionTimeout)
	if err != nil {
		return
	}
	tcon = con.(*net.TCPConn)
	if keepalivePeriod > 0 {
		tcon.SetKeepAlive(true)
		tcon.SetKeepAlivePeriod(keepalivePeriod)
	}
	return
}
