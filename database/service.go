package database

import (
	"github.com/enorith/database"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/kernel"
)

var (
	DB *database.Connection
)

type ServiceProvider struct {
	config database.Config
}

// Register preregister database connection drivers
func (s *ServiceProvider) Register(app *kernel.Application) {
	s.initDB(app)
	app.Defer(func() {
		database.DefaultManager.CloseAll()
	})

	app.ConfigRuntime(func(runtime *kernel.Application) {
		runtime.WithInjector(Injector{runtime: runtime})
	})
}

func (s *ServiceProvider) Boot(app *kernel.Application) {
	database.Cache = cache.AppCache
}

func (s *ServiceProvider) initDB(app *kernel.Application) {
	database.WithDefaultDrivers()
	app.Configure("database", &s.config)
}

func NewProvider() *ServiceProvider {
	return &ServiceProvider{}
}
