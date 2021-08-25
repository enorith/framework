package framework

import (
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/reflection"
)

type Service interface {
	//Register service when app starting, before http server start
	// you can configure service, initialize global vars etc.
	// running at main goroutine
	Register(app *App) error

	//Lifetime container callback
	// usually register request lifetime instance to IoC-Container (per-request unique)
	// this function will run before every request
	Lifetime(ioc container.Interface, request contracts.RequestContract)
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
func (cs *ConfigService) Lifetime(container container.Interface, request contracts.RequestContract) {
	cs.mu.RLock()
	defer cs.mu.RUnlock()
	for k, v := range cs.configs {
		container.Singleton("config."+k, reflection.StructValue(v))
		container.Singleton(reflection.StructType(v), reflection.StructValue(v))
	}
}
