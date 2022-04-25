package queue

import (
	"fmt"
	"reflect"
	"time"

	"github.com/enorith/queue"
	"github.com/enorith/queue/contracts"
	"github.com/enorith/queue/std"
	"github.com/enorith/supports/reflection"
)

//Dispatch payload to queue
func Dispatch(payload interface{}) error {
	return queue.DefaultDispatcher.Dispatch(payload)
}

//DispatchAfter, display payload to queue, handle after delay
func DispatchAfter(payload interface{}, delay time.Duration) error {
	return queue.DefaultDispatcher.After(delay).Dispatch(payload)
}

//DispatchOn display payload to queue connection
func DispatchOn(payload interface{}, on string) error {
	return queue.DefaultDispatcher.On(on).Dispatch(payload)
}

//DispatchOnAfter display payload to queue connection, handle after delay
func DispatchOnAfter(payload interface{}, on string, delay time.Duration) error {
	return queue.DefaultDispatcher.On(on).After(delay).Dispatch(payload)
}

//RegisterHandler register queue job handler, automaticly listen job
// fn is job handler func, first parameter suppose to be queue job payload, support parameter injection
func RegisterHandler(fn interface{}) error {
	t := reflection.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("[queue] job handler only accept a func parameter, %T giving", fn)
	}
	if t.NumIn() < 1 {
		return fmt.Errorf("[queue] job handler should contain one param of payload")
	}
	pt := t.In(0)
	pv := reflect.New(pt)
	ipv := reflect.Indirect(pv)
	var job interface{}

	if np, ok := ipv.Interface().(contracts.NamedPayload); ok {
		job = np.PayloadName()
	} else {
		job = pt
	}

	std.Listen(job, fn)
	return nil
}
