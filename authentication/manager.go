package authentication

import (
	"fmt"
	"sync"

	"github.com/enorith/authenticate"
	"github.com/enorith/http/contracts"
)

//AuthManager default manager
var AuthManager = &Manager{
	drivers:   make(map[string]DriverRegister),
	providers: make(map[string]AuthProvider),
	mu:        sync.RWMutex{},
}

//DriverRegister register of auth guard
type DriverRegister func(up AuthProvider, r contracts.RequestContract) (authenticate.Guard, error)

//GuardConfig config of auth guard
type GuardConfig struct {
	Driver   string `yaml:"driver"`
	Provider string `yaml:"provider"`
}

//Config of authentication
type Config struct {
	Default string                 `yaml:"default"`
	Guards  map[string]GuardConfig `yaml:"guards"`
}

//Manager of authentication
type Manager struct {
	drivers   map[string]DriverRegister
	providers map[string]AuthProvider
	mu        sync.RWMutex
}

//GetDriverRegister get auth guard driver register
func (m *Manager) GetDriverRegister(name string) (DriverRegister, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	dr, ok := m.drivers[name]

	return dr, ok
}

//GetProvider get auth provider
func (m *Manager) GetProvider(name string) (AuthProvider, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	p, ok := m.providers[name]

	return p, ok
}

//WithProvider register auth provider
func (m *Manager) WithProvider(name string, p AuthProvider) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.providers[name] = p
}

//RegisterDriver register auth guard driver
func (m *Manager) RegisterDriver(name string, dr DriverRegister) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.drivers[name] = dr
}

//GetGuard get auth guard
func (m *Manager) GetGuard(driver, provider string, r contracts.RequestContract) (authenticate.Guard, error) {

	dr, ok := m.GetDriverRegister(driver)
	if !ok {
		return nil, fmt.Errorf("unregisterd auth driver [%s]", driver)
	}
	p, ok := m.GetProvider(provider)
	if !ok {
		return nil, fmt.Errorf("unregisterd auth provider [%s]", driver)
	}

	return dr(p, r)
}
