package database

import "time"

type ConnectionConfig struct {
	Driver string `yaml:"driver" default:"mysql"`
	DSN    string `yaml:"dsn" env:"DB_DSN"`
}

type Config struct {
	Default     string                      `yaml:"default" env:"DB_CONNECTION"`
	Connections map[string]ConnectionConfig `yaml:"connections"`
	//ImplicitInjection: whether injection model to request handler implicit
	//  usage:
	//      w.Get("/users/:user", func(user models.User) models.User {
	//			return user
	//      })
	//
	ImplicitInjection bool   `yaml:"implicit_injection"`
	AuthMigrate       bool   `yaml:"migrate"`
	LogChannel        string `yaml:"log_channel"`
}

var (
	MaxIdelConns = 10
	MaxOpenConns = 100
	MaxLifeTime  = time.Hour
	MaxIdleTime  = time.Hour
)
