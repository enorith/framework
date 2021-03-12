package database

type ConnectionConfig struct {
	Driver   string `yaml:"driver" env:"DB_DRIVER" default:"mysql"`
	Host     string `yaml:"host" env:"DB_HOST" default:"localhost"`
	Username string `yaml:"username" env:"DB_DATABASE" default:"root"`
	Password string `yaml:"password" env:"DB_PASSWORD" default:"root"`
	Port     int    `yaml:"port" env:"DB_PORT" default:"3306"`
	Database string `yaml:"database" env:"DB_DATABASE" default:"enorith"`
}

type Config struct {
	Default     string                      `yaml:"default" env:"DB_CONNECTION"`
	Connections map[string]ConnectionConfig `yaml:"connections"`
}
