package session

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"path/filepath"
	"sync"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/cache"
	"github.com/enorith/http/contracts"
	"github.com/enorith/session"
	"github.com/enorith/session/handlers"
	"github.com/enorith/supports/reflection"
	"github.com/enorith/supports/str"
)

type HandlerRegister func() (session.Handler, error)

var (
	sessionHandlerRegisters = make(map[string]HandlerRegister, 0)
	sessionHandlers         = make(map[string]session.Handler, 0)
	mu                      sync.RWMutex
)
var (
	handlerType = reflection.InterfaceType[session.Handler]()
	Manager     *session.Manager
)

type SessionID struct {
	ID string
}

func RegisterHandler(name string, register HandlerRegister) {
	mu.Lock()
	defer mu.Unlock()
	sessionHandlerRegisters[name] = register
}

func GetHandler(name string) (session.Handler, error) {
	mu.RLock()
	if handler, ok := sessionHandlers[name]; ok {
		mu.RUnlock()
		return handler, nil
	}
	register, ok := sessionHandlerRegisters[name]
	mu.RUnlock()
	if !ok {
		return nil, fmt.Errorf("[session] handler (%s) is not registered", name)
	}

	handler, err := register()
	if err != nil {
		return nil, err
	}
	sessionHandlers[name] = handler
	return handler, nil
}

type Config struct {
	Default     string `yaml:"default" env:"SESSION_HANDLER" default:"file"`
	Dir         string `yaml:"dir" default:"sessions"`
	MaxLifeTime int    `yaml:"max_life_time" default:"600"`
	CacheStore  string `yaml:"cache_store" default:"session"`
	CookieName  string `yaml:"cookie_name" default:"enorith-session" env:"COOKIE_NAME"`
	Domain      string `yaml:"domain" default:""`
	Path        string `yaml:"path" default:"/"`
	Secure      bool   `yaml:"secure" default:"false"`
	HttpOnly    bool   `yaml:"http_only" default:"true"`
	SameSite    string `yaml:"same_site" default:"lax"`
	Encrypted   bool   `yaml:"encrypted" default:"false"`
}

type Service struct {
	config      Config
	storagePath string
	sameSite    http.SameSite
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	app.Configure("session", &s.config)
	if s.config.SameSite != "" {
		switch s.config.SameSite {
		case "default":
			s.sameSite = http.SameSiteDefaultMode
		case "lax":
			s.sameSite = http.SameSiteLaxMode
		case "strict":
			s.sameSite = http.SameSiteStrictMode
		case "none":
			s.sameSite = http.SameSiteNoneMode
		default:
			s.sameSite = http.SameSiteDefaultMode
		}
	}

	s.registerFileHandler()
	err := s.registerCacheHandler()
	if err != nil {
		return err
	}
	handler, err := GetHandler(s.config.Default)
	if err != nil {
		return err
	}
	Manager = session.NewManager(handler)
	app.Daemon(func(exit chan struct{}) {
		log.Println("[session] gc started")
		for {
			select {
			case <-exit:
				return
			case <-time.After(time.Second * 3):
				Manager.GC(time.Duration(s.config.MaxLifeTime) * time.Second)
			}
		}
	})

	return nil
}
func (s *Service) registerFileHandler() {
	RegisterHandler("file", func() (session.Handler, error) {
		return handlers.NewFileSessionHandler(filepath.Join(s.storagePath, s.config.Dir)), nil
	})
}

func (s *Service) registerCacheHandler() error {
	RegisterHandler("cache", func() (session.Handler, error) {
		if s.config.CacheStore == "" {
			return nil, errors.New("[session] cache store is not configured")
		}

		if cache.Default == nil {
			return nil, errors.New("[session] cache service is not configured")
		}
		repo, ok := cache.Default.GetRepository(s.config.CacheStore)
		if !ok {
			return nil, fmt.Errorf("[session] cache store (%s) is not found", s.config.CacheStore)
		}
		return handlers.NewCacheHandler(repo, time.Duration(s.config.MaxLifeTime)*time.Second), nil
	})
	return nil
}

//Lifetime container callback
// usually register request lifetime instance to IoC-Container (per-request unique)
// this function will run before every request handling
func (s *Service) Lifetime(ioc container.Interface, request contracts.RequestContract) {
	var id string
	if rc, ok := request.(contracts.WithRequestCookies); ok {
		cb := rc.CookieByte(s.config.CookieName)
		if len(cb) > 0 {
			id = string(cb)
		} else {
			id = str.RandString(32)
		}
	} else {
		id = str.RandString(32)
	}

	ioc.Bind(SessionID{}, SessionID{ID: id}, true)

	ioc.BindFunc(&session.Session{}, func(c container.Interface) (interface{}, error) {
		return Manager.Get(id), nil
	}, true)

	ioc.BindFunc("middleware.session", func(c container.Interface) (interface{}, error) {
		return Middleware{
			manager:     Manager,
			id:          id,
			cookieName:  s.config.CookieName,
			maxLifeTime: int(s.config.MaxLifeTime),
			domain:      s.config.Domain,
			path:        s.config.Path,
			secure:      s.config.Secure,
			httpOnly:    s.config.HttpOnly,
			sameSite:    s.sameSite,
		}, nil
	}, true)

}

func NewService(storagePath string) *Service {
	return &Service{storagePath: storagePath, sameSite: http.SameSiteDefaultMode}
}
