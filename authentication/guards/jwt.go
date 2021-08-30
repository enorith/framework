package guards

import "github.com/enorith/http/contracts"

//JWTConfig jwt auth guard config
type JWTConfig struct {
	Key string `yaml:"key" env:"AUTH_JWT_KEY"`
	TTL int    `yaml:"ttl" env:"AUTH_JWT_TTL"`
}

//TokenProvider of jwt auth guard
type TokenProvider struct {
	Request contracts.RequestContract
}

//GetAccessToken: get request token (bearer)
func (tp TokenProvider) GetAccessToken() ([]byte, error) {
	return tp.Request.BearerToken()
}
