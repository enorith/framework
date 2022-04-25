package crond

import (
	"log"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/go-co-op/gocron"
)

var Scheduler *gocron.Scheduler

type Service struct {
}

//Register service when app starting, before http server start
// you can configure service, prepare global vars etc.
// running at main goroutine
func (s Service) Register(app *framework.App) error {
	Scheduler = gocron.NewScheduler(time.Local)
	app.Bind(func(ioc container.Interface) {
		ioc.BindFunc(&gocron.Scheduler{}, func(c container.Interface) (interface{}, error) {
			return Scheduler, nil
		}, true)
	})

	app.Daemon(func(exit chan struct{}) {
		Scheduler.StartAsync()
		log.Println("[cron] started")
		<-exit
		Scheduler.Stop()
		log.Println("[cron] stopped")
	})

	return nil
}
