package center

import (
	"github.com/obase/conf"
)

const (
	CKEY  = "center"
	CKEY2 = "service.center" // 兼容旧版本,新版本建议放置在顶层
	LOCAL = "local"
)

/*
不再兼容828旧逻辑,没有"service.center"就不会自动启用!
*/
func init() {
	config, ok := conf.Get(CKEY)
	if !ok {
		// 兼容旧版本,新版本建议放置在顶层
		config, ok = conf.Get(CKEY2)
		if !ok {
			return
		}
	}
	switch config := config.(type) {
	case nil:
		Setup(&Config{Address: LOCAL})
	case string:
		Setup(&Config{Address: config})
	case map[interface{}]interface{}:
		var service map[string][]string
		address, ok := conf.ElemString(config, "address")
		expired, ok := conf.ElemInt64(config, "expired")
		tmp, ok := conf.ElemMap(config, "service")
		if ok {
			service = make(map[string][]string)
			for k, v := range tmp {
				service[k] = conf.ToStringSlice(v)
			}
		}
		refresh, ok := conf.ElemInt(config, "refresh")
		Setup(&Config{
			Address: address,
			Expired: expired,
			Service: service,
			Refresh: refresh,
		})
	}
}
