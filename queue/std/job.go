package std

import (
	"fmt"
	"reflect"

	"github.com/enorith/supports/reflection"
	"github.com/vmihailenco/msgpack/v5"
)

type Job struct {
	payloadType string
	payload     []byte
}

func (j Job) Invoke(fn interface{}) error {
	t := reflection.TypeOf(fn)
	if t.Kind() != reflect.Func {
		return fmt.Errorf("job invoke only accept a func parameter, %T giving", fn)
	}
	ti := t.In(0)
	iv := reflect.New(ti)

	msgpack.Unmarshal(j.payload, reflect.Indirect(iv).Interface())

	fv := reflect.ValueOf(fn)

	fv.Call([]reflect.Value{iv})

	return nil
}

func ToJob(payload interface{}) (j Job, e error) {
	j.payloadType = reflection.TypeString(payload)
	j.payload, e = msgpack.Marshal(payload)

	return
}
