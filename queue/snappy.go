package queue

import (
	"time"

	"github.com/enorith/queue"
)

//Dispatch payload to queue
func Dispatch(payload interface{}) error {
	return queue.DefaultDispatcher.Dispatch(payload)
}

//DispatchAfter, display payload to queue, handle after delay
func DispatchAfter(payload interface{}, delay time.Duration) error {
	return queue.DefaultDispatcher.After(delay).Dispatch(payload)
}

//DispatchOn queue connection
func DispatchOn(payload interface{}, on string) error {
	return queue.DefaultDispatcher.On(on).Dispatch(payload)
}
