package framework

import (
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/supports/reflection"
)

type Service interface {
	//Register service when app starting, before http server start
	// you can configure service, prepare global vars etc.
	// running at main goroutine
	Register(app *App) error
}

//ConfigService of application
type ConfigService struct {
	configs map[string]interface{}
	mu      sync.RWMutex
}

//Add config instance to service
func (cs *ConfigService) Add(name string, config interface{}) *ConfigService {
	cs.mu.Lock()
	defer cs.mu.Unlock()
	cs.configs[name] = config
	return cs
}

//Register config service when app starting
//
func (cs *ConfigService) Register(app *App) error {
	app.Bind(func(ioc container.Interface) {
		cs.mu.RLock()
		defer cs.mu.RUnlock()
		for k, v := range cs.configs {
			sv := reflection.StructValue(v)
			st := reflection.StructType(v)
			resolver := func(c container.Interface) (interface{}, error) {
				return sv, nil
			}

			ioc.BindFunc(st, resolver, true)
			ioc.BindFunc("config."+k, resolver, true)
		}
	})

	return nil
}
