package queue

import (
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/queue/connections"
	c "github.com/enorith/framework/queue/contracts"
	"github.com/enorith/framework/queue/std"
	"github.com/enorith/http/contracts"
)

var DefaultDispatcher Dispatcher

type Service struct {
	config Config
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("queue", &s.config)
	DefaultManager.RegisterConnection("nsq", func(config map[string]interface{}) c.Connection {
		return connections.NewNsq(config)
	})

	DefaultManager.RegisterConnection("mem", func(config map[string]interface{}) c.Connection {
		return connections.NewMem()
	})
	s.configure()
	if s.config.Listening {
		app.Daemon(func(done chan struct{}) {
			DefaultManager.Work(done, s.config.RunningWorkers...)
		})
	}

	app.Defer(func() {
		DefaultManager.Close(s.config.RunningWorkers...)
	})

	DefaultDispatcher = Dispatcher{on: s.config.Connection}
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request handling
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(Dispatcher{}, func(c container.Interface) (interface{}, error) {
		return DefaultDispatcher, nil
	}, false)
}

func (s *Service) configure() error {
	for w, wc := range s.config.Workers {
		t, conConfig := s.connectionConfig(wc.Connection)
		c, e := DefaultManager.ResolveConnection(t, conConfig)
		if e != nil {
			return e
		}

		DefaultManager.RegisterWorker(w, std.NewWorker(wc.Concurrency, c))
	}

	return nil
}

func (s *Service) connectionConfig(con string) (string, map[string]interface{}) {
	conConfig := s.config.Connections[con]
	if t, ok := conConfig["type"].(string); ok {
		return t, conConfig
	}

	return s.config.Connection, conConfig
}

func NewService() *Service {
	return &Service{}
}
