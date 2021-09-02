package contracts

import (
	"os"
	"time"
)

type Connection interface {
	Consume(concurrency int, done chan os.Signal) error
	Stop() error
	Dispatch(payload interface{}, delay ...time.Duration) error
}
