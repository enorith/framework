package http

import (
	"fmt"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/http/rules"
	"github.com/enorith/gormdb"
	h "github.com/enorith/http"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
	"github.com/enorith/http/validation"
	"github.com/enorith/http/validation/rule"
)

type Config struct {
	Port      int  `yaml:"port" env:"HTTP_PORT" default:"8000"`
	AccessLog bool `yaml:"access_log" env:"HTTP_ACCESS_LOG" default:"false"`
}

type HttpService interface {
	//Lifetime container callback
	// usually register request lifetime instance to IoC-Container (per-request unique)
	// this function will run before every request handling
	Lifetime(ioc container.Interface, request contracts.RequestContract)
}

type RoutesRegister interface {
	RegisterRoutes(rw *router.Wrapper)
}

type Service struct {
	rg     func(rw *router.Wrapper)
	config Config
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {

	app.Configure("http", &s.config)
	validation.Register("unique", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		var field string
		if len(args) == 0 {
			return nil, fmt.Errorf("validation: table arg required (usage 'unique:table')")
		}

		if len(args) > 1 {
			field = args[1]
		} else {
			field = attribute
		}
		tx, e := gormdb.DefaultManager.GetConnection()
		if e != nil {
			return nil, fmt.Errorf("validation: unique rule initialize error, %s", e.Error())
		}

		return rules.NewUnique(tx, args[0], field), nil
	})

	app.Daemon(func(done chan struct{}) {
		config := app.GetConfig()
		services := app.Services()
		server := NewServer(func(request contracts.RequestContract) container.Interface {
			ioc := container.New()
			for _, s := range services {
				if hs, ok := s.(HttpService); ok {
					hs.Lifetime(ioc, request)
				}
			}
			return ioc
		}, config.Debug)
		server.Serve(fmt.Sprintf(":%d", s.config.Port), func(rw *router.Wrapper, k *h.Kernel) {
			k.OutputLog = s.config.AccessLog
			for _, s := range services {
				if rr, ok := s.(RoutesRegister); ok {
					rr.RegisterRoutes(rw)
				}
			}
			s.rg(rw)
		}, done)
	})

	return nil
}

func NewService(rg func(rw *router.Wrapper)) *Service {
	return &Service{rg: rg}
}
