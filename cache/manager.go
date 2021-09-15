package cache

import (
	"fmt"
	"sync"

	"github.com/enorith/cache"
	"gopkg.in/yaml.v3"
)

type DriverResolver func(conf yaml.Node) (cache.Repository, error)

var (
	driverResolvers = make(map[string]DriverResolver)
	mu              sync.RWMutex
)

func Resolve(name string, dr DriverResolver) {
	mu.Lock()
	defer mu.Unlock()
	driverResolvers[name] = dr
}

func ResolveDriver(name string, config yaml.Node) (cache.Repository, error) {
	mu.RLock()
	defer mu.RUnlock()
	if dr, ok := driverResolvers[name]; ok {
		return dr(config)
	}

	return nil, fmt.Errorf("[cache] unregister resolver %s", name)
}
