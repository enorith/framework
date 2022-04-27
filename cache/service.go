package cache

import (
	"reflect"
	"time"

	c "github.com/enorith/cache"
	"github.com/enorith/config"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	appRedis "github.com/enorith/framework/redis"
	"github.com/go-redis/cache/v8"
	"github.com/go-redis/redis/v8"
	gc "github.com/patrickmn/go-cache"
	"gopkg.in/yaml.v3"
)

var Default *c.Manager

var RepositoryType = reflect.TypeOf((*c.Repository)(nil)).Elem()

type StoreConfig struct {
	Driver string    `yaml:"driver"`
	Prefix string    `yaml:"prefix"`
	Config yaml.Node `yaml:"config"`
}

type CacheConfig struct {
	Store  string                 `yaml:"store" env:"CACHE_DRIVER" default:"go_cache"`
	Prefix string                 `yaml:"prefix" env:"CACHE_PREFIX" default:""`
	Stores map[string]StoreConfig `yaml:"stores"`
}

type RedisConfig struct {
	Connection string `yaml:"connection"`
}

type Service struct {
}

func (s Service) Register(app *framework.App) error {

	var cc CacheConfig
	app.Configure("cache", &cc)
	c.KeyPrefix = cc.Prefix
	s.registerStores(cc)
	Default = c.NewManager()

	app.Bind(func(ioc container.Interface) {
		resolver := func(c container.Interface) (interface{}, error) {
			return Default, nil
		}
		ioc.BindFunc(&c.Manager{}, resolver, true)
		ioc.BindFunc(RepositoryType, resolver, true)
	})

	return Default.Use(cc.Store)
}

func (s Service) registerStores(cc CacheConfig) {
	Resolve("go_cache", func(config yaml.Node, sc StoreConfig) (c.Repository, error) {
		return c.NewGoCache(gc.New(c.DefaultExpiration, c.CleanupInterval), sc.Prefix), nil
	})

	Resolve("redis", func(conf yaml.Node, sc StoreConfig) (c.Repository, error) {
		var rc RedisConfig
		config.UnmarshalNode(conf, &rc)
		var ring redis.UniversalClient
		var e error
		if rc.Connection != "" {
			ring, e = appRedis.Manager.GetConnection(rc.Connection)
		} else {
			ring, e = appRedis.Manager.GetConnection()
		}

		if e != nil {
			return nil, e
		}

		return c.NewRedisCache(&cache.Options{
			Redis:        ring,
			LocalCache:   cache.NewTinyLFU(1000, time.Minute),
			StatsEnabled: false,
		}, sc.Prefix), nil
	})

	for k, sc := range cc.Stores {
		config := sc
		c.RegisterDriver(k, func() (c.Repository, error) {
			return ResolveDriver(config.Driver, config.Config, config)
		})
	}
}
