package framework

import (
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/http/contracts"
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

	return nil
}

//Lifetime container callback
// register config instance to request IoC-Container
func (cs *ConfigService) Lifetime(ioc container.Interface, request contracts.RequestContract) {
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
}
