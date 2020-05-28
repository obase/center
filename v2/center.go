package v2

import "errors"

// 未作初始化
var ErrInvalidState = errors.New("invalid state")

type Config struct {
	Address string              // 远程地址
	Expired int64               // 缓存过期时间(秒)
	Service map[string][]string // 本地配置服务
}

var ()

func Setup() {

}
