package contracts

import (
	"time"
)

type Connection interface {
	Consume(concurrency int, done chan struct{}) error
	Stop() error
	Dispatch(payload interface{}, delay ...time.Duration) error
}
