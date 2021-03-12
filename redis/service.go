package redis

import (
	"fmt"
	"strings"

	"github.com/enorith/framework/kernel"
	rds "github.com/go-redis/redis"
)

var appRedis rds.Cmdable

var GetClient func() rds.Cmdable

type RedisConfig struct {
	Hosts string `yaml:"hosts" default:"127.0.0.1:6379"`
	DB    int    `yaml:"database" default:"0"`
}

type ServiceProvider struct {
}

func (s *ServiceProvider) Register(app *kernel.Application) {
	var rc RedisConfig
	app.Configure("redis", &rc)
	addresses := s.parseAddress(rc.Hosts)
	db := rc.DB

	GetClient = func() rds.Cmdable {
		if appRedis != nil {
			return appRedis
		}

		if len(addresses) > 0 {
			addrs := make(map[string]string)
			for k, v := range addresses {
				addrs[fmt.Sprintf("redis%d", k)] = v
			}
			appRedis = rds.NewRing(&rds.RingOptions{
				Addrs: addrs,
				DB:    db,
			})
		} else {
			appRedis = rds.NewClient(&rds.Options{
				Addr: addresses[0],
				DB:   db,
			})
		}
		return appRedis
	}
}

func (s *ServiceProvider) Boot(app *kernel.Application) {

}

func (s *ServiceProvider) parseAddress(address string) []string {
	return strings.Split(address, ",")
}
