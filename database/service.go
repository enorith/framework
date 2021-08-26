package database

import (
	"fmt"
	"reflect"
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
}

//Register service when app starting, before http server start
// you can configure service, initialize global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	var config Config
	app.Configure("database", &config)

	for name, cc := range config.Connections {
		gormdb.DefaultManager.Register(name, func() (*gorm.DB, error) {
			register, ok := GetDriverRegister(cc.Driver)
			if !ok {
				return nil, fmt.Errorf("unregistered database driver [%s]", cc.Driver)
			}
			return gorm.Open(register(cc.DSN))
		})
	}

	gormdb.DefaultManager.Using(config.Default)

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
	ioc.BindFunc(&gorm.DB{}, func(c container.Interface) (reflect.Value, error) {
		db, e := gormdb.DefaultManager.GetConnection()

		return reflect.ValueOf(db), e
	}, false)
}

func init() {
	RegisterDriver("mysql", func(dsn string) gorm.Dialector {
		return mysql.Open(dsn)
	})
}
