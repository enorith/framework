package injection

import (
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/supports/reflection"
)

var IocType = reflection.InterfaceType[container.Interface]()

type Service struct {
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Bind(func(ioc container.Interface) {
		ioc.Bind(IocType, ioc, true)
	})

	return nil
}

func NewService() *Service {
	return &Service{}
}
