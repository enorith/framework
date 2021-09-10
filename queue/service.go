package queue

import (
	"fmt"
	"sync"

	"github.com/enorith/config"
	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
	"github.com/enorith/queue"
	"github.com/enorith/queue/connections"
	c "github.com/enorith/queue/contracts"
	"github.com/enorith/queue/std"
	"gopkg.in/yaml.v3"
)

type DriverRegister func(config yaml.Node) (c.Connection, error)

var (
	driverRegisters = make(map[string]DriverRegister)
	mu              sync.RWMutex
)

func RegisterDriver(name string, dr DriverRegister) {
	mu.Lock()
	defer mu.Unlock()

	driverRegisters[name] = dr
}

func GetDriverRegister(name string) (DriverRegister, bool) {
	mu.RLock()
	defer mu.RUnlock()

	dr, ok := driverRegisters[name]

	return dr, ok
}

type WorkerConfig struct {
	Connection  string `yaml:"connection"`
	Concurrency int    `yaml:"concurrency"`
}

type ConConf struct {
	Driver string    `yaml:"driver"`
	Config yaml.Node `yaml:"config"`
}

type Config struct {
	Listening      bool                    `yaml:"listen" env:"QUEUE_LISTEN" default:"true"`
	Connection     string                  `yaml:"connection" env:"QUEUE_CONNECTION" default:"mem"`
	RunningWorkers []string                `yaml:"running_workers"`
	Workers        map[string]WorkerConfig `yaml:"workers"`
	Connections    map[string]ConConf      `yaml:"connections"`
}

type NsqConfig struct {
	Nsqd       string `yaml:"nsqd" env:"QUEUE_NSQD"`
	Nsqlookupd string `yaml:"nsqlookupd" env:"QUEUE_NSQLOOKUPD"`
	Channel    string `yaml:"channel" env:"QUEUE_NSQCHANNEL" default:"default"`
	Topic      string `yaml:"topic" env:"QUEUE_NSQCHANNEL" default:"default"`
}

type Service struct {
	config Config
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("queue", &s.config)
	WithDefaults()
	e := s.configure()
	if e != nil {
		return e
	}
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
	for con, cf := range s.config.Connections {
		dr, ok := GetDriverRegister(cf.Driver)
		if !ok {
			return fmt.Errorf("queue driver [%s] not registerd", cf.Driver)
		}
		conf := cf.Config

		queue.DefaultManager.RegisterConnection(con, func() (c.Connection, error) {
			return dr(conf)
		})
	}

	for w, wc := range s.config.Workers {
		c, e := queue.DefaultManager.GetConnection(wc.Connection)
		if e != nil {
			return e
		}
		queue.DefaultManager.RegisterWorker(w, std.NewWorker(wc.Concurrency, c))
	}

	return nil
}

func NewService() *Service {
	return &Service{}
}

func WithDefaults() {
	RegisterDriver("nsq", func(conf yaml.Node) (c.Connection, error) {
		var c NsqConfig
		e := config.UnmarshalNode(conf, &c)
		if e != nil {
			return nil, e
		}

		return connections.NewNsqFromConfig(connections.NsqConfig{
			Nsqd:    c.Nsqd,
			Lookupd: c.Nsqlookupd,
			Channel: c.Channel,
			Topic:   c.Topic,
		}), nil
	})

	RegisterDriver("mem", func(conf yaml.Node) (c.Connection, error) {

		return connections.NewMem(), nil
	})
}
