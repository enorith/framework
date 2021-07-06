package database

import (
	"github.com/enorith/database"
	"github.com/enorith/framework"
	"github.com/enorith/framework/cache"
)

var (
	DB *database.Connection
)

type ServiceProvider struct {
	config database.Config
}

// Register preregister database connection drivers
func (s *ServiceProvider) Register(app *framework.Application) {
	s.initDB(app)
	app.Defer(func() {
		database.DefaultManager.CloseAll()
	})

	app.ConfigRuntime(func(runtime *framework.Application) {
		runtime.WithInjector(Injector{runtime: runtime})
	})
}

func (s *ServiceProvider) Boot(app *framework.Application) {
	database.Cache = cache.AppCache
}

func (s *ServiceProvider) initDB(app *framework.Application) {
	database.WithDefaultDrivers()
	app.Configure("database", &s.config)
}

func NewProvider() *ServiceProvider {
	return &ServiceProvider{}
}
