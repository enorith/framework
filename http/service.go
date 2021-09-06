package http

import (
	"fmt"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/http/rules"
	"github.com/enorith/gormdb"
	h "github.com/enorith/http"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/validation"
	"github.com/enorith/http/validation/rule"
)

type HttpService interface {
	//Lifetime container callback
	// usually register request lifetime instance to IoC-Container (per-request unique)
	// this function will run before every request handling
	Lifetime(ioc container.Interface, request contracts.RequestContract)
}

type Service struct {
	rg h.RouterRegister
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
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
		server.Serve(fmt.Sprintf(":%d", config.Port), s.rg, done)
	})

	return nil
}

func NewService(rg h.RouterRegister) *Service {
	return &Service{rg: rg}
}
