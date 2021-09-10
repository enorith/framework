package framework

import (
	"io/fs"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/enorith/config"
	"github.com/enorith/http/router"
	"github.com/enorith/supports/carbon"
)

type RouterRegister func(rw *router.Wrapper)

type DaemonFn func(done chan struct{})

//AppConfig: default app config name
var AppConfig = "app"

//ConfigExt: default config extension
var ConfigExt = ".yaml"

//Config of application
type Config struct {
	Env      string `yaml:"env" env:"APP_ENV" default:"production"`
	Debug    bool   `yaml:"debug" env:"APP_DEBUG" default:"false"`
	Locale   string `yaml:"locale" env:"APP_LOCALE" default:"en"`
	Url      string `yaml:"url" env:"APP_URL" default:"http://localhost"`
	Timezone string `yaml:"timezone" env:"APP_TIMEZONE" default:""`
	Key      string `yaml:"timezone" env:"APP_KEY" default:"somerandomkey!!!"`
}

//App: framework application
type App struct {
	services        []Service
	config          Config
	configFs        fs.FS
	configService   *ConfigService
	routerRegisters []RouterRegister
	defers          []func()
	daemons         []DaemonFn
}

//Register application service provider
// service
func (app *App) Register(service Service) *App {
	app.services = append(app.services, service)
	return app
}

//Configure: load config instance and add to config service
func (app *App) Configure(name string, value interface{}) *App {
	config.UnmarshalFS(app.configFs, name+ConfigExt, value)
	app.configService.Add(name, value)

	return app
}

//ConfigureFS: load config from filesystem
func (app *App) ConfigureFS(fs fs.FS, name string, value interface{}) *App {
	config.UnmarshalFS(fs, name+ConfigExt, value)
	app.configService.Add(name, value)

	return app
}

//GetConfig: get app config instance
func (app *App) GetConfig() Config {
	return app.config
}

//Bootstrap application, will call before app run
func (app *App) Bootstrap() error {
	app.Configure(AppConfig, &app.config)
	if app.config.Timezone != "" {
		loc, e := time.LoadLocation(app.config.Timezone)
		if e == nil {
			carbon.Timezone = loc
		}
	}
	app.Register(app.configService)
	// app.configService.Register(app)
	for _, s := range app.services {
		e := s.Register(app)
		if e != nil {
			return e
		}
	}

	return nil
}

//Run application service
func (app *App) Run() error {
	e := app.Bootstrap()
	if e != nil {
		return e
	}
	wg := new(sync.WaitGroup)
	app.RunDaemons(wg)
	wg.Wait()
	app.RunDefers()
	return nil
}

//RunWithoutServer, run background services without http server
func (app *App) RunWithoutServer() error {
	e := app.Bootstrap()
	if e != nil {
		return e
	}
	wg := new(sync.WaitGroup)
	app.RunDaemons(wg)
	wg.Wait()
	app.RunDefers()
	return nil
}

//GetEnv: get app env
func (app *App) GetEnv() string {
	return app.config.Env
}

//RegisterRoutes: register routes of http service
func (app *App) RegisterRoutes(rr RouterRegister) *App {

	app.routerRegisters = append(app.routerRegisters, rr)
	return app
}

//Defer run function, run after server shutdown
func (app *App) Defer(f func()) *App {

	app.defers = append(app.defers, f)
	return app
}

//Daemon run function, start daemon service before http service started
func (app *App) Daemon(f DaemonFn) *App {

	app.daemons = append(app.daemons, f)
	return app
}

func (app *App) RunDefers() {
	for _, f := range app.defers {
		f()
	}
}

//RunDaemons, run background services
func (app *App) RunDaemons(wg *sync.WaitGroup, daemon ...bool) {
	d := true
	if len(daemon) > 0 {
		d = daemon[0]
	}

	lenDaemon := len(app.daemons)
	done := make(chan struct{}, lenDaemon)

	kill := make(chan os.Signal, 1)
	signal.Notify(kill, os.Interrupt, syscall.SIGTERM, syscall.SIGINT)

	for _, f := range app.daemons {
		wg.Add(1)

		go func(f DaemonFn) {
			defer wg.Done()
			f(done)
		}(f)
	}
	wait := func() {
		<-kill
		i := 0
		for i < lenDaemon {
			done <- struct{}{}
			i++
		}
	}

	if d {
		go wait()
	} else {
		wait()
	}
}
func (app *App) Services() []Service {
	return app.services
}

//NewApp: new application instance
func NewApp(configFs fs.FS) *App {

	return &App{
		configFs:        configFs,
		services:        make([]Service, 0),
		configService:   &ConfigService{configs: make(map[string]interface{})},
		routerRegisters: make([]RouterRegister, 0),
		defers:          make([]func(), 0),
		daemons:         make([]DaemonFn, 0),
	}
}
