package center

import (
	"github.com/obase/conf"
)

const (
	PCKEY1 = "service.center"
	PCKEY2 = "ext.service.centerAddr"

	OFF = "off"
)

func init() {
	config, ok := conf.Get(PCKEY1)
	if !ok {
		config, ok = conf.Get(PCKEY2)
	}
	/*
	   为了兼容旧的828逻辑, 在没有声明center或centerAddr的情况默认连接本地. 所以声明一个特殊值"none","off"表示关闭
	*/
	switch config := config.(type) {
	case nil:
		Setup(&Config{Address: ""})
	case string:
		if config != OFF {
			Setup(&Config{Address: config})
		}
	case map[interface{}]interface{}:
		service := make(map[string][]string)
		for k, v := range config {
			service[k.(string)] = v.([]string)
		}
		Setup(&Config{Service: service})
	}
}
