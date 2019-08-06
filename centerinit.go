package center

import (
	"github.com/obase/conf"
	"time"
)

const (
	PCKEY1 = "service.center"
	PCKEY2 = "ext.service.centerAddr"

	OFF = "off"
)

const DEFAULT_TIMEOUT = 5 * time.Second

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
		Setup(&Config{Address: "", Timeout: DEFAULT_TIMEOUT})
	case string:
		if config != OFF {
			Setup(&Config{Address: config, Timeout: DEFAULT_TIMEOUT})
		}
	case map[interface{}]interface{}:
		var option *Config
		conf.Convert(config, &option)
		if option != nil && option.Timeout == 0 {
			option.Timeout = DEFAULT_TIMEOUT
		}
		Setup(option)
	}
}
