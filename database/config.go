package database

import "time"

type ConnectionConfig struct {
	Driver string `yaml:"driver" env:"DB_DRIVER" default:"mysql"`
	DSN    string `yaml:"dsn" env:"DB_DSN"`
}

type Config struct {
	Default     string                       `yaml:"default" env:"DB_CONNECTION"`
	Connections map[string]*ConnectionConfig `yaml:"connections"`
}

var (
	MaxIdelConns = 10
	MaxOpenConns = 100
	MaxLifeTime  = time.Hour
	MaxIdleTime  = time.Hour
)
