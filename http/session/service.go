package session

import (
	"path/filepath"
	"sync"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/http/contracts"
	"github.com/enorith/session"
	"github.com/enorith/session/handlers"
)

var (
	sessionHandlers = make(map[string]session.Handler, 0)
	mu              sync.RWMutex
)

func RegisterHandler(name string, handler session.Handler) {
	mu.Lock()
	defer mu.Unlock()
	sessionHandlers[name] = handler
}

func GetHandler(name string) (session.Handler, bool) {
	mu.RLock()
	defer mu.RUnlock()
	handler, ok := sessionHandlers[name]

	return handler, ok
}

type Config struct {
	Default     string `yaml:"default" env:"SESSION_HANDLER" default:"file"`
	Dir         string `yaml:"dir" default:"sessions"`
	MaxLifeTime int64  `yaml:"max_life_time"`
	CacheStore  string `yaml:"cache_store"`
}

type Service struct {
	config      Config
	storagePath string
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("session", &s.config)

	return nil
}
func (s *Service) registerFileHandler() {
	RegisterHandler("file", handlers.NewFileSessionHandler(filepath.Join(s.storagePath, s.config.Dir)))
}

func (s *Service) registerCacheHandler() error {
	// if s.config.CacheStore == "" {
	// 	return errors.New("[session] cache store is not configured")
	// }

	// if cache.Default == nil {
	// 	return errors.New("[session] cache service is not configured")
	// }

	// RegisterHandler("cache", handlers.NewCacheHandler())
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request handling
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {

}

func NewService(storagePath string) *Service {
	return &Service{storagePath: storagePath}
}
