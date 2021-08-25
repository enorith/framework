package cache

import (
	"reflect"
	"time"

	c "github.com/enorith/cache"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	appRedis "github.com/enorith/framework/redis"
	"github.com/enorith/http/contracts"
	"github.com/go-redis/cache/v8"
	gc "github.com/patrickmn/go-cache"
)

var AppCache *c.Manager

type CacheConfig struct {
	Driver string `yaml:"driver" env:"CACHE_DRIVER" default:"go_cache"`
	Prefix string `yaml:"prefix" env:"CACHE_PREFIX" default:""`
}

type Service struct {
	cc CacheConfig
}

func (s *Service) Register(app *framework.App) error {
	app.Configure("cache", &s.cc)
	c.KeyPrefix = s.cc.Prefix
	s.registerDefaultDrivers()
	AppCache = c.NewManager(s.cc.Driver)
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(&c.Manager{}, func(c container.Interface) (reflect.Value, error) {
		return reflect.ValueOf(AppCache), nil
	}, true)
}

func (s *Service) registerDefaultDrivers() {
	c.RegisterDriver("go_cache", func() c.Repository {
		return c.NewGoCache(gc.New(c.DefaultExpiration, c.CleanupInterval))
	})
	c.RegisterDriver("redis", func() c.Repository {

		ring := appRedis.Client

		return c.NewRedisCache(&cache.Options{
			Redis:        ring,
			LocalCache:   cache.NewTinyLFU(1000, time.Minute),
			StatsEnabled: false,
		})
	})
}
