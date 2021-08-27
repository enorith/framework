package guards

import "github.com/enorith/http/contracts"

type JWTConfig struct {
	Key string `yaml:"key" env:"AUTH_JWT_KEY"`
	TTL int    `yaml:"ttl" env:"AUTH_JWT_TTL"`
}

type TokenProvider struct {
	Request contracts.RequestContract
}

func (tp TokenProvider) GetAccessToken() ([]byte, error) {
	return tp.Request.BearerToken()
}
