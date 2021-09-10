package database

import (
	"fmt"
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/gormdb"
	"github.com/enorith/http/contracts"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

type DriverRegister func(dsn string) gorm.Dialector

var Migrator func(tx *gorm.DB)

var (
	driverRegisters = make(map[string]DriverRegister)
	mu              sync.RWMutex
	DefaultPageSize = 15
)

//GetDriverRegister: get registerd driver
func GetDriverRegister(name string) (DriverRegister, bool) {
	mu.RLock()
	defer mu.RUnlock()
	register, ok := driverRegisters[name]
	return register, ok
}

//RegisterDriver register db driver
func RegisterDriver(name string, dr DriverRegister) {
	mu.Lock()
	defer mu.Unlock()
	driverRegisters[name] = dr
}

//Service of database
type Service struct {
	config Config
}

//Register service when app starting, before http server start
// you can configure service, initialize global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("database", &s.config)
	WithDefaults()

	for name, cc := range s.config.Connections {
		gormdb.DefaultManager.Register(name, func() (*gorm.DB, error) {
			register, ok := GetDriverRegister(cc.Driver)
			if !ok {
				return nil, fmt.Errorf("unregistered database driver [%s]", cc.Driver)
			}
			tx, e := gorm.Open(register(cc.DSN))
			if e != nil {
				return nil, e
			}
			db, e := tx.DB()
			if e != nil {
				return nil, e
			}
			db.SetMaxIdleConns(MaxIdelConns)
			db.SetMaxOpenConns(MaxOpenConns)
			db.SetConnMaxIdleTime(MaxIdleTime)
			db.SetConnMaxLifetime(MaxLifeTime)
			return tx, e
		})
	}

	gormdb.DefaultManager.Using(s.config.Default)

	if Migrator != nil {
		if tx, e := gormdb.DefaultManager.GetConnection(); e == nil {
			Migrator(tx)
		} else {
			return e
		}
	}

	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	ioc.BindFunc(&gorm.DB{}, func(c container.Interface) (interface{}, error) {

		return gormdb.DefaultManager.GetConnection()
	}, false)

	ioc.BindFunc(&gormdb.Paginator{}, func(c container.Interface) (interface{}, error) {
		page, _ := request.GetInt("page")
		perPage, _ := request.GetInt("per_page")

		if perPage < 1 {
			perPage = DefaultPageSize
		}

		return gormdb.NewPaginator(page, perPage), nil
	}, false)

	ioc.WithInjector(Injector{r: request, c: s.config, ioc: ioc})
}

func NewService() *Service {
	return &Service{}
}

func WithDefaults() {
	RegisterDriver("mysql", func(dsn string) gorm.Dialector {
		return mysql.Open(dsn)
	})
}
