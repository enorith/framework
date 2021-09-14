package cache

import (
	"reflect"
	"time"

	c "github.com/enorith/cache"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	appRedis "github.com/enorith/framework/redis"
	"github.com/go-redis/cache/v8"
	gc "github.com/patrickmn/go-cache"
)

var Default *c.Manager

var RepositoryType = reflect.TypeOf((*c.Repository)(nil)).Elem()

type CacheConfig struct {
	Driver string `yaml:"driver" env:"CACHE_DRIVER" default:"go_cache"`
	Prefix string `yaml:"prefix" env:"CACHE_PREFIX" default:""`
}

type Service struct {
	cc CacheConfig
}

func (s Service) Register(app *framework.App) error {
	app.Configure("cache", &s.cc)
	c.KeyPrefix = s.cc.Prefix
	s.registerDefaultDrivers()
	Default = c.NewManager()

	app.Bind(func(ioc container.Interface) {
		resolver := func(c container.Interface) (interface{}, error) {
			return Default, nil
		}
		ioc.BindFunc(&c.Manager{}, resolver, true)
		ioc.BindFunc(RepositoryType, resolver, true)
	})

	return Default.Use(s.cc.Driver)
}

func (s Service) registerDefaultDrivers() {
	c.RegisterDriver("go_cache", func() (c.Repository, error) {
		return c.NewGoCache(gc.New(c.DefaultExpiration, c.CleanupInterval)), nil
	})
	c.RegisterDriver("redis", func() (c.Repository, error) {

		ring, e := appRedis.Manager.GetConnection()
		if e != nil {
			return nil, e
		}

		return c.NewRedisCache(&cache.Options{
			Redis:        ring,
			LocalCache:   cache.NewTinyLFU(1000, time.Minute),
			StatsEnabled: false,
		}), nil
	})
}
