package redis

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
	rds "github.com/go-redis/redis/v8"
)

var Client rds.Cmdable

var redisType = reflect.TypeOf((*rds.Cmdable)(nil)).Elem()

type RedisConfig struct {
	Hosts string `yaml:"hosts" default:"127.0.0.1:6379"`
	DB    int    `yaml:"database" default:"0"`
}

type Service struct {
}

func (s Service) Register(app *framework.App) error {
	var rc RedisConfig
	app.Configure("redis", &rc)
	addresses := s.parseAddress(rc.Hosts)
	if len(addresses) > 0 {
		addrs := make(map[string]string)
		for k, v := range addresses {
			addrs[fmt.Sprintf("redis%d", k)] = v
		}
		Client = rds.NewRing(&rds.RingOptions{
			Addrs: addrs,
			DB:    rc.DB,
		})
	} else {
		Client = rds.NewClient(&rds.Options{
			Addr: addresses[0],
			DB:   rc.DB,
		})
	}
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(redisType, func(c container.Interface) (reflect.Value, error) {
		return reflect.ValueOf(Client), nil
	}, true)
}

func (s *Service) parseAddress(address string) []string {
	return strings.Split(address, ",")
}
