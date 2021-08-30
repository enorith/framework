package framework

import (
	"io/fs"
	"time"

	"github.com/enorith/config"
	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/router"
	"github.com/enorith/supports/carbon"
)

type RouterRegister func(rw *router.Wrapper)

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
}

//App: framework application
type App struct {
	services        []Service
	config          Config
	configFs        fs.FS
	configService   *ConfigService
	routerRegisters []RouterRegister
}

//Register application service
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
func (app *App) Bootstrap() (*http.Server, error) {
	app.Configure(AppConfig, &app.config)
	if app.config.Timezone != "" {
		loc, e := time.LoadLocation(app.config.Timezone)
		if e == nil {
			carbon.Timezone = loc
		}
	}

	app.configService.Register(app)
	for _, s := range app.services {
		e := s.Register(app)
		if e != nil {
			return nil, e
		}
	}
	server := http.NewServer(func(request contracts.RequestContract) container.Interface {
		con := container.New()
		app.configService.Lifetime(con, request)
		for _, s := range app.services {
			s.Lifetime(con, request)
		}
		return con
	}, app.config.Debug)
	return server, nil
}

//Run application service
func (app *App) Run(at string, register http.RouterRegister) error {
	server, e := app.Bootstrap()
	if e != nil {
		return e
	}

	server.Serve(at, func(rw *router.Wrapper, k *http.Kernel) {
		register(rw, k)
		for _, rr := range app.routerRegisters {
			rr(rw)
		}
	})
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

//NewApp: new application instance
func NewApp(configFs fs.FS) *App {

	return &App{
		configFs:        configFs,
		services:        make([]Service, 0),
		configService:   &ConfigService{configs: make(map[string]interface{})},
		routerRegisters: make([]RouterRegister, 0),
	}
}
