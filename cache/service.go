package cache

import (
	"reflect"
	"time"

	. "github.com/enorith/cache"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	appRedis "github.com/enorith/framework/redis"
	"github.com/go-redis/cache/v8"
	gc "github.com/patrickmn/go-cache"
)

var AppCache *Manager

type CacheConfig struct {
	Driver string `yaml:"driver" env:"CACHE_DRIVER" default:"go_cache"`
	Prefix string `yaml:"prefix" env:"CACHE_PREFIX" default:""`
}

type ServiceProvider struct {
	cc CacheConfig
}

func (s *ServiceProvider) Register(app *framework.Application) {
	app.Configure("cache", &s.cc)
	KeyPrefix = s.cc.Prefix
	s.registerDefaultDrivers()

	app.BindRuntimeFunc(&Manager{}, func(c container.Interface) reflect.Value {
		return reflect.ValueOf(NewManager(s.cc.Driver))
	}, true)
}

func (s *ServiceProvider) Boot(app *framework.Application) {
	AppCache = NewManager(s.cc.Driver)
}

func (s *ServiceProvider) registerDefaultDrivers() {
	RegisterDriver("go_cache", func() Repository {
		return NewGoCache(gc.New(DefaultExpiration, CleanupInterval))
	})
	RegisterDriver("redis", func() Repository {

		ring := appRedis.GetClient()

		return NewRedisCache(&cache.Options{
			Redis:        ring,
			LocalCache:   cache.NewTinyLFU(1000, time.Minute),
			StatsEnabled: false,
		})
	})
}
