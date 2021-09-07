package queue

import (
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
	"github.com/enorith/queue"
	"github.com/enorith/queue/connections"
	c "github.com/enorith/queue/contracts"
	"github.com/enorith/queue/std"
)

type WorkerConfig struct {
	Connection  string `yaml:"connection"`
	Concurrency int    `yaml:"concurrency"`
}

type Config struct {
	Listening      bool                              `yaml:"listen" env:"QUEUE_LISTEN" default:"true"`
	Connection     string                            `yaml:"connection" env:"QUEUE_CONNECTION" default:"mem"`
	RunningWorkers []string                          `yaml:"running_workers"`
	Workers        map[string]WorkerConfig           `yaml:"workers"`
	Connections    map[string]map[string]interface{} `yaml:"connections"`
}

type Service struct {
	config Config
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("queue", &s.config)
	queue.DefaultManager.RegisterConnection("nsq", func(config map[string]interface{}) c.Connection {
		return connections.NewNsq(config)
	})

	queue.DefaultManager.RegisterConnection("mem", func(config map[string]interface{}) c.Connection {
		return connections.NewMem()
	})
	s.configure()
	if s.config.Listening {
		app.Daemon(func(done chan struct{}) {
			queue.DefaultManager.Work(done, s.config.RunningWorkers...)
		})
	}

	app.Defer(func() {
		queue.DefaultManager.Close(s.config.RunningWorkers...)
	})

	queue.DefaultDispatcher = queue.Dispatcher{DefaultConnection: s.config.Connection}
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request handling
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(queue.Dispatcher{}, func(c container.Interface) (interface{}, error) {
		return queue.DefaultDispatcher, nil
	}, false)
}

func (s *Service) configure() error {
	for w, wc := range s.config.Workers {
		t, conConfig := s.connectionConfig(wc.Connection)
		c, e := queue.DefaultManager.ResolveConnection(t, conConfig)
		if e != nil {
			return e
		}
		queue.DefaultManager.RegisterWorker(w, std.NewWorker(wc.Concurrency, c))
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
