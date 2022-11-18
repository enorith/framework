package crond

import (
	"context"
	"log"
	"time"

	"github.com/enorith/container"
	"github.com/enorith/framework"
	"github.com/go-co-op/gocron"
)

var Scheduler *gocron.Scheduler

type CronRegister func(sch *gocron.Scheduler, ctx context.Context)

var register CronRegister

func RegisterCron(rg CronRegister) {
	register = rg
}

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
		ctx, cancel := context.WithCancel(context.Background())
		if register != nil {
			register(Scheduler, ctx)
		}
		Scheduler.StartAsync()
		log.Println("[cron] started")
		<-exit
		log.Println("[cron] stopping")
		cancel()
		Scheduler.Stop()
		log.Println("[cron] stopped")
	})

	return nil
}
