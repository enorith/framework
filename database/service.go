package database

import (
	"fmt"
	"sync"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/environment"
	"github.com/enorith/framework"
	"github.com/enorith/gormdb"
	"github.com/enorith/http/contracts"
	"github.com/enorith/logging"
	"github.com/enorith/supports/reflection"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

type DriverRegister func(dsn string) gorm.Dialector

var Migrator func(tx *gorm.DB)

var (
	driverRegisters = make(map[string]DriverRegister)
	mu              sync.RWMutex
	DefaultPageSize = 15
	PageKey         = "page"
	PageSizeKey     = "per_page"
	PageParamsType  = reflection.InterfaceType[PageParams]()
	logLevelMap     = map[string]logger.LogLevel{
		"info":   logger.Info,
		"warn":   logger.Warn,
		"error":  logger.Error,
		"silent": logger.Silent,
	}
)

// GetDriverRegister: get registerd driver
func GetDriverRegister(name string) (DriverRegister, bool) {
	mu.RLock()
	defer mu.RUnlock()
	register, ok := driverRegisters[name]
	return register, ok
}

// RegisterDriver register db driver
func RegisterDriver(name string, dr DriverRegister) {
	mu.Lock()
	defer mu.Unlock()
	driverRegisters[name] = dr
}

// Service of database
type Service struct {
	config Config
}

// Register service when app starting, before http server start
// you can configure service, initialize global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("database", &s.config)
	WithDefaults()
	defaultLogger, _ := logging.DefaultManager.Channel(s.config.LogChannel)

	for name, cc := range s.config.Connections {
		config := cc
		n := name
		gormdb.DefaultManager.Register(n, func() (*gorm.DB, error) {
			dsn := config.DSN
			if n == s.config.Default {
				envDsn := environment.GetString("DB_DSN")
				if envDsn != "" {
					dsn = envDsn
					config.DSN = dsn
				}
			}
			log := defaultLogger
			if lc := config.LogChannel; lc != "" {
				log, _ = logging.DefaultManager.Channel(lc)
				if log == nil {
					log = defaultLogger
				}
			}
			logLevel := logger.Info
			if ll := config.LogLevel; ll != "" {
				var ok bool
				logLevel, ok = logLevelMap[ll]
				if !ok {
					logLevel = logger.Info
				}
			}

			register, ok := GetDriverRegister(config.Driver)
			if !ok {
				return nil, fmt.Errorf("unregistered database driver [%s]", config.Driver)
			}
			conf := &gorm.Config{}
			if defaultLogger != nil {
				conf.Logger = &Logger{
					logLevel:      logLevel,
					logger:        log,
					SlowThreshold: 300 * time.Millisecond,
					withMigration: s.config.WithMigrationLog,
				}
			}
			conf.DisableForeignKeyConstraintWhenMigrating = !s.config.WithForeignKey
			tx, e := gorm.Open(register(dsn), conf)
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

	if s.config.AuthMigrate && Migrator != nil {
		if tx, e := gormdb.DefaultManager.GetConnection(); e == nil {
			Migrator(tx)
		} else if defaultLogger != nil {
			defaultLogger.Error("[database] migration error %v")
		}
	}

	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(&gorm.DB{}, func(c container.Interface) (interface{}, error) {
			return gormdb.DefaultManager.GetConnection()
		}, false)
	})

	return nil
}

// Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {

	ioc.BindFunc(&gormdb.Paginator{}, func(c container.Interface) (interface{}, error) {
		page, _ := request.GetInt(PageKey)
		perPage, _ := request.GetInt(PageSizeKey)

		if perPage < 1 {
			perPage = DefaultPageSize
		}

		return gormdb.NewPaginator(page, perPage), nil
	}, false)

	ioc.BindFunc(PageParamsType, func(c container.Interface) (interface{}, error) {
		return RequestPageParams{request: request}, nil
	}, true)

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
