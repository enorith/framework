package authentication

import (
	"github.com/enorith/authenticate"
	"github.com/enorith/authenticate/jwt"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/authentication/guards"
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/reflection"
)

var GuardType = reflection.InterfaceType[authenticate.Guard]()

type Service struct {
	config Config
}

// Register service when app starting, before http server start
// you can configure service, initialize global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("auth", &s.config)
	s.withJWT(app)

	return nil
}
func (s *Service) withJWT(app *framework.App) {
	var jwtConfig guards.JWTConfig
	app.Configure("jwt", &jwtConfig)
	AuthManager.RegisterDriver("jwt", func(up AuthProvider, r contracts.RequestContract) (authenticate.Guard, error) {
		return jwt.NewJwtGuard(guards.TokenProvider{Request: r}, up, []byte(jwtConfig.Key), jwtConfig.TTL), nil
	})
}

// Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	gc, ok := s.config.Guards[s.config.Default]

	if ok {
		guardResolver := func(c container.Interface) (interface{}, error) {
			return AuthManager.GetGuard(gc.Driver, gc.Provider, request)
		}
		ioc.BindFunc("auth.guard", guardResolver, true)
		ioc.BindFunc(GuardType, guardResolver, true)
	}
}

func NewAuthService() *Service {
	return &Service{}
}
