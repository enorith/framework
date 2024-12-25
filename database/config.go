package database

import "time"

type ConnectionConfig struct {
	Driver     string `yaml:"driver" default:"mysql"`
	DSN        string `yaml:"dsn"`
	LogChannel string `yaml:"log_channel"`
	LogLevel   string `yaml:"log_level" default:"info"`
}

type Config struct {
	Default     string                       `yaml:"default" env:"DB_CONNECTION"`
	Connections map[string]*ConnectionConfig `yaml:"connections"`
	//ImplicitInjection: whether injection model to request handler implicit
	//  usage:
	//      w.Get("/users/:user", func(user models.User) models.User {
	//			return user
	//      })
	//
	ImplicitInjection    bool   `yaml:"implicit_injection"`
	AuthMigrate          bool   `yaml:"migrate" env:"DB_MIGRATE"`
	LogChannel           string `yaml:"log_channel"`
	WithForeignKey       bool   `yaml:"with_foreign_key" default:"false"`
	WithoutRelationships bool   `yaml:"without_relationships" default:"false"`
	WithMigrationLog     bool   `yaml:"with_migration_log" default:"false"`
}

var (
	MaxIdelConns = 10
	MaxOpenConns = 100
	MaxLifeTime  = time.Hour
	MaxIdleTime  = time.Hour
)
