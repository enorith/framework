package cache

import (
	. "github.com/enorith/cache"
	"github.com/enorith/environment"
	"github.com/enorith/framework/kernel"
	appRedis "github.com/enorith/framework/redis"
	"github.com/go-redis/cache"
	gc "github.com/patrickmn/go-cache"
	"github.com/vmihailenco/msgpack"
)

var AppCache *Manager

type ServiceProvider struct {
}

func (s *ServiceProvider) Register(app *kernel.Application) {
	s.registerDefaultDrivers()
}

func (s *ServiceProvider) Boot(app *kernel.Application) {
	AppCache = NewManager(rithenv.GetString("CACHE_DRIVER", "go_cache"))
}

func (s *ServiceProvider) registerDefaultDrivers() {
	RegisterDriver("go_cache", func() Repository {
		return NewGoCache(gc.New(DefaultExpiration, CleanupInterval))
	})
	RegisterDriver("redis", func() Repository {
		codec := &cache.Codec{
			Redis: appRedis.GetClient(),

			Marshal: func(v interface{}) ([]byte, error) {
				return msgpack.Marshal(v)
			},
			Unmarshal: func(b []byte, v interface{}) error {
				return msgpack.Unmarshal(b, v)
			},
		}

		return NewRedisCache(codec)
	})
}
