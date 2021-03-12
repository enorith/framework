package cache

import (
	"time"

	. "github.com/enorith/cache"
	"github.com/enorith/framework/kernel"
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
}

func (s *ServiceProvider) Register(app *kernel.Application) {
	s.registerDefaultDrivers()
}

func (s *ServiceProvider) Boot(app *kernel.Application) {
	var cc CacheConfig
	app.Configure("cache", &cc)
	KeyPrefix = cc.Prefix

	AppCache = NewManager(cc.Driver)
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
