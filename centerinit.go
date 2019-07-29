package center

import (
	"github.com/obase/conf"
	"time"
)

const PCKEY = "service.center"

const DEFAULT_TIMEOUT = time.Minute

func init() {
	config, ok := conf.Get(PCKEY)
	if !ok {
		return
	}

	switch config := config.(type) {
	case string:
		Setup(&Option{Address: config, Timeout: DEFAULT_TIMEOUT})
	case map[string]interface{}:
		address, _ := conf.ElemString(config, "address")
		timeout, ok := conf.ElemDuration(config, "timeout")
		if !ok {
			timeout = DEFAULT_TIMEOUT
		}
		Setup(&Option{Address: address, Timeout: timeout})
	}
}
