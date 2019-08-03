package center

import (
	"github.com/obase/conf"
	"github.com/obase/log"
	"time"
)

const PCKEY = "service.center"

const DEFAULT_TIMEOUT = 5 * time.Second

func init() {
	config, ok := conf.Get(PCKEY)
	if !ok {
		return
	}

	switch config := config.(type) {
	case string:
		Setup(&Option{Address: config, Timeout: DEFAULT_TIMEOUT})
	case map[interface{}]interface{}:
		var option *Option
		conf.Scan(PCKEY, &option)
		if option != nil && option.Timeout == 0 {
			option.Timeout = DEFAULT_TIMEOUT
		}
		Setup(option)
	default:
		log.Errorf(nil, "Invalid config option of "+PCKEY)
	}
}
