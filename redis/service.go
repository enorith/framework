package redis

import (
	"reflect"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/redis"
	rds "github.com/go-redis/redis/v8"
)

var Manager *redis.Manager

var redisType = reflect.TypeOf((*rds.UniversalClient)(nil)).Elem()

type ConnectionConfig struct {
	Addrs      []string `yaml:"addrs"`
	DB         int      `yaml:"database" env:"REDIS_DB"`
	Username   string   `yaml:"username"`
	Password   string   `yaml:"password"`
	MasterName string   `yaml:"master_name"`
	PoolSize   int      `yaml:"pool_size" default:"1000"`
	MaxIdle    int      `yaml:"max_idle" default:"10"`
}

type RedisConfig struct {
	Default     string `yaml:"default" env:"REDIS_CONNECTION"`
	Hosts       string `yaml:"hosts" default:"127.0.0.1:6379" env:"REDIS_HOSTS"`
	DB          int    `yaml:"database" default:"0" env:"REDIS_DB"`
	Username    string `yaml:"username"`
	Password    string `yaml:"password"`
	Connections map[string]ConnectionConfig
}

type Service struct {
}

func (s Service) Register(app *framework.App) error {
	var rc RedisConfig
	app.Configure("redis", &rc)
	Manager = redis.NewManager()

	s.loadConfig(rc)

	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(redisType, func(c container.Interface) (interface{}, error) {
			return Manager.GetConnection()
		}, true)
	})
	return nil
}

func (s *Service) loadConfig(rc RedisConfig) {
	Manager.Use(rc.Default)

	for k, cc := range rc.Connections {
		conf := cc
		Manager.Register(k, func() rds.UniversalClient {
			return rds.NewUniversalClient(&rds.UniversalOptions{
				Addrs:        conf.Addrs,
				DB:           conf.DB,
				MasterName:   conf.MasterName,
				PoolSize:     conf.PoolSize,
				MinIdleConns: conf.MaxIdle,
				Username:     conf.Username,
				Password:     conf.Password,
			})
		})
	}
}
