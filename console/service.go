package console

import (
	"bufio"
	"bytes"
	"fmt"
	"net"
	"os"
	"path/filepath"

	"github.com/enorith/framework"
	"github.com/enorith/logging"
)

type Config struct {
	Socket string `yaml:"socket"`
}

type Service struct {
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	var config Config
	app.Configure("console", &config)
	if config.Socket == "" {
		wd, _ := os.Getwd()
		config.Socket = filepath.Join(wd, "enorith.sock")
	}

	app.Daemon(func(exit chan struct{}) {
		lis, err := net.Listen("unix", config.Socket)
		if err != nil {
			logging.Infof("[console] socket listening error %v", err)
			return
		}
		logging.Infof("[console] socket listening %s", config.Socket)
		go func() {
			for {
				conn, err := lis.Accept()
				reader := bufio.NewReader(conn)

				if err != nil {
					logging.Infof("[console] socket accept error %v", err)
					return
				}
				go func() {
					defer conn.Close()
					for {

						data, e := reader.ReadBytes('\n')
						if e == nil {
							logging.Infof("[console] socket accept %s", data)
							conn.Write([]byte(fmt.Sprintf("accepted [%s]\n", bytes.TrimSpace(data))))
						}
					}
				}()

			}
		}()

		<-exit
		logging.Info("[console] socket closing")
		lis.Close()
	})
	return nil
}

func NewService() *Service {
	return &Service{}
}
