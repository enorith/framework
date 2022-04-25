package http

import (
	"fmt"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/enorith/framework/http/rules"
	"github.com/enorith/gormdb"
	h "github.com/enorith/http"
	"github.com/enorith/http/content"
	"github.com/enorith/http/contracts"
	"github.com/enorith/http/errors"
	"github.com/enorith/http/router"
	"github.com/enorith/http/validation"
	"github.com/enorith/http/validation/rule"
	"github.com/enorith/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	Port                int      `yaml:"port" env:"HTTP_PORT" default:"8000"`
	AccessLog           bool     `yaml:"access_log" env:"HTTP_ACCESS_LOG" default:"false"`
	ErrorLogChannel     string   `yaml:"error_log_channel" default:"default"`
	AccessLogChannel    string   `yaml:"access_log_channel"`
	TrustedProxies      []string `yaml:"trusted_proxies"`
	TrustedProxyHeaders []string `yaml:"trusted_proxy_headers"`
}

type HttpBoundle interface {
	//Lifetime container callback
	// usually register request lifetime instance to IoC-Container (per-request unique)
	// this function will run before every request handling
	Lifetime(ioc container.Interface, request contracts.RequestContract)
}

type RoutesRegister interface {
	RegisterRoutes(rw *router.Wrapper)
}

type Service struct {
	rg     func(rw *router.Wrapper)
	config Config
}

//WithRoutes, with http routes
func (s *Service) WithRoutes(rg func(rw *router.Wrapper)) *Service {
	s.rg = rg
	return s
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {

	app.Configure("http", &s.config)
	content.SetTrustedProxies(s.config.TrustedProxies...)
	content.SetTrustedProxyHeaderSets(s.config.TrustedProxyHeaders...)
	validation.Register("unique", func(attribute string, r contracts.InputSource, args ...string) (rule.Rule, error) {
		var field string
		if len(args) == 0 {
			return nil, fmt.Errorf("validation: table arg required (usage 'unique:table')")
		}

		if len(args) > 1 {
			field = args[1]
		} else {
			field = attribute
		}
		tx, e := gormdb.DefaultManager.GetConnection()
		if e != nil {
			return nil, fmt.Errorf("validation: unique rule initialize error, %s", e.Error())
		}

		return rules.NewUnique(tx, args[0], field), nil
	})
	config := app.GetConfig()
	services := app.Services()
	kernel := h.NewKernel(func(request contracts.RequestContract) container.Interface {
		ioc := app.Container().Clone()
		for _, s := range services {
			if hs, ok := s.(HttpBoundle); ok {
				hs.Lifetime(ioc, request)
			}
		}
		return ioc
	}, config.Debug)
	server := NewServer(kernel)

	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(&h.Kernel{}, func(c container.Interface) (interface{}, error) {
			return kernel, nil
		}, true)
	})

	app.Daemon(func(done chan struct{}) {
		server.Serve(fmt.Sprintf(":%d", s.config.Port), func(rw *router.Wrapper, k *h.Kernel) {
			k.OutputLog = s.config.AccessLog
			for _, s := range services {
				if rr, ok := s.(RoutesRegister); ok {
					rr.RegisterRoutes(rw)
				}
			}
			if s.rg != nil {
				s.rg(rw)
			}
		}, done)
	})

	app.Resolving(func(lm *logging.Manager) {
		logger, e := lm.Channel(s.config.ErrorLogChannel)
		if e == nil {
			kernel.SetErrorHandler(&errors.StandardErrorHandler{
				Debug: config.Debug,
				Callback: func(ed errors.ErrorData) {
					if ed.Fatal {
						logger.WithOptions(zap.WithCaller(false),
							zap.AddStacktrace(loggerTraceEnabler{})).Error(ed.Message, zap.Any("trace", ed.Traces),
							zap.String("caller", fmt.Sprintf("%s:%d", ed.File, ed.Line)))
					}
				},
			})
		}
		if s.config.AccessLogChannel != "" {
			accessLogger, e := lm.Channel(s.config.AccessLogChannel)
			if e == nil {
				h.RequestLogger = func(request contracts.RequestContract, statusCode int, start time.Time) {

					duration := time.Since(start)
					accessLogger.WithOptions(zap.WithCaller(false)).Info("",
						zap.String("remote", request.RemoteAddr()),
						zap.String("latency", fmt.Sprintf("%.6fms", float64(duration)/float64(time.Millisecond))),
						zap.String("path", string(request.GetUri())),
						zap.Int("status_code", statusCode))
				}
			}
		}
	})

	return nil
}

func NewService() *Service {
	return &Service{}
}

type loggerTraceEnabler struct {
}

func (loggerTraceEnabler) Enabled(zapcore.Level) bool {
	return false
}
