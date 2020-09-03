package database

import (
	"github.com/enorith/database"
	"github.com/enorith/database/orm"
	"github.com/enorith/database/rithythm"
	"github.com/enorith/framework/cache"
	"github.com/enorith/framework/container"
	"github.com/enorith/framework/kernel"
	"reflect"
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
		for !database.Conns.Empty() {
			database.Conns.ShiftAndClose()
		}
	})
	app.BindRuntimeFunc(&database.QueryBuilder{}, func(c *container.Container) reflect.Value {
		return reflect.ValueOf(NewDefaultBuilder())
	}, false)
	app.BindRuntimeFunc(&orm.Builder{}, func(c *container.Container) reflect.Value {
		return reflect.ValueOf(&orm.Builder{QueryBuilder: NewDefaultBuilder()})
	}, false)
}

func (s *ServiceProvider) Boot(app *kernel.Application) {
	database.Cache = cache.AppCache
}

func (s *ServiceProvider) initDB(app *kernel.Application) {
	database.WithDefaultDrivers()
	app.Configure("database", &s.config)
	DB = database.NewConnection(s.config.Default, s.config)
	rithythm.Config(s.config)
}

func NewProvider() *ServiceProvider {
	return &ServiceProvider{}
}

func NewDefaultBuilder() *database.QueryBuilder {
	return database.NewBuilder(DB.Clone())
}
