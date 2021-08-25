package framework

import (
	"io/fs"

	"github.com/enorith/config"
	"github.com/enorith/container"
	"github.com/enorith/http"
	"github.com/enorith/http/contracts"
)

//AppConfig: default app config name
var AppConfig = "app"

//ConfigExt: default config extension
var ConfigExt = ".yaml"

//Config of application
type Config struct {
	Env   string `yaml:"env" env:"APP_ENV" default:"production"`
	Debug bool   `yaml:"debug" env:"APP_DEBUG" default:"false"`
}

//App: framework application
type App struct {
	services      []Service
	config        Config
	configFs      fs.FS
	configService *ConfigService
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

//GetConfig: get app config instance
func (app *App) GetConfig() Config {
	return app.config
}

//Bootstrap application, will call before app run
func (app *App) Bootstrap() (*http.Server, error) {
	app.Configure(AppConfig, &app.config)

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

	server.Serve(at, register)
	return nil
}

//NewApp: new application instance
func NewApp(configFs fs.FS) *App {

	return &App{
		configFs:      configFs,
		services:      make([]Service, 0),
		configService: &ConfigService{configs: make(map[string]interface{})},
	}
}
