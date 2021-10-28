package console

import (
	"bufio"
	"bytes"
	"fmt"
	"log"
	"net"
	"os"
	"path/filepath"

	"github.com/enorith/framework"
)

type Config struct {
}

type Service struct {
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s *Service) Register(app *framework.App) error {
	wd, _ := os.Getwd()
	file := filepath.Join(wd, "enorith.sock")
	app.Daemon(func(exit chan struct{}) {
		lis, err := net.Listen("unix", file)
		if err != nil {
			log.Println("[console] socket listening error ", err)
			return
		}
		log.Printf("[console] socket listening %s", file)
		go func() {
			for {
				conn, err := lis.Accept()
				reader := bufio.NewReader(conn)

				if err != nil {
					log.Println("[console] socket accept error ", err)
					return
				}
				go func() {
					defer conn.Close()
					for {
						data, e := reader.ReadBytes('\n')
						if e == nil {
							log.Printf("[console] socket accept %s", data)
							conn.Write([]byte(fmt.Sprintf("accepted [%s]\n", bytes.TrimSpace(data))))
						}
					}
				}()

			}
		}()

		<-exit
		log.Printf("[console] socket closing")
		lis.Close()
	})
	return nil
}

func NewService() *Service {
	return &Service{}
}
