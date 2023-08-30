package framework

import (
	"log"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/logging"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	LabelKey = "label"
)

type LoggingConfig struct {
	Default  string `yaml:"default" env:"LOGGING_CHANNEL" default:"default"`
	Channels map[string]LogChannelConfig
}

type LogChannelConfig struct {
	Outputs    []string `yaml:"outputs"`
	Errouts    []string `yaml:"errouts"`
	Encoding   string   `yaml:"encoding" default:"json"`
	TimeFormat string   `yaml:"time_format" default:"2006-01-02T15:04:05.999"`
	Label      string   `yaml:"label"`
}

type LoggingService struct {
	baseDir string
}

// Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *LoggingService) Register(app *App) error {
	var conf LoggingConfig
	app.Configure("logging", &conf)

	logging.WithDefaults(logging.Config{
		BaseDir:  s.baseDir,
		Fallback: logging.NewStdLogger(),
	})

	logging.DefaultManager.Using(conf.Default)

	for ch, cc := range conf.Channels {
		cr := cc
		logging.DefaultManager.Resolve(ch, func(conf zap.Config) (*zap.Logger, error) {
			conf.OutputPaths = cr.Outputs
			conf.ErrorOutputPaths = cr.Errouts
			if cr.TimeFormat == "" {
				cr.TimeFormat = "2006-01-02T15:04:05.999"
			}
			if cr.Encoding == "" {
				cr.Encoding = "json"
			}
			conf.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout(cr.TimeFormat)
			conf.EncoderConfig.StacktraceKey = "trace"
			conf.Encoding = cr.Encoding
			fields := make(map[string]interface{})
			if cr.Label != "" {
				fields[LabelKey] = cr.Label
			}

			conf.InitialFields = fields

			return conf.Build()
		})
	}

	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(&zap.Logger{}, func(c container.Interface) (interface{}, error) {
			return logging.DefaultManager.Channel()
		}, true)

		ioc.BindFunc(&logging.Manager{}, func(c container.Interface) (interface{}, error) {
			return logging.DefaultManager, nil
		}, true)
	})
	app.Daemon(func(exit chan struct{}) {
		log.Println("[logging] service cleanup started")
		logging.DefaultManager.Cleanup()
		for {
			select {
			case <-exit:
				logging.DefaultManager.Sync()
				log.Println("[logging] service exit")
				return
			case <-time.After(5 * time.Minute):
				logging.DefaultManager.Cleanup()
			}
		}
	})

	return nil
}

func NewLoggingService(baseDir string) *LoggingService {
	return &LoggingService{
		baseDir: baseDir,
	}
}
