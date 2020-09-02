package redis

import (
	"fmt"
	env "github.com/enorith/environment"
	"github.com/enorith/framework/kernel"
	rds "github.com/go-redis/redis"
	"strings"
)

var appRedis rds.Cmdable

var GetClient func() rds.Cmdable

type ServiceProvider struct {
}

func (s *ServiceProvider) Register(app *kernel.Application) {
	addresses := s.parseAddress(env.GetString("REDIS_HOSTS", ":6379"))
	db := env.GetInt("REDIS_DATABASE", 0)

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
