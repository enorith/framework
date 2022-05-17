package session

import (
	"reflect"

	"github.com/enorith/supports/reflection"
)

var (
	slType = reflection.InterfaceType[SessionLoader]()
)

type SessionLoader interface {
	SessionKey() string
}

type SessionInjector struct {
	id string
}

func (si *SessionInjector) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {
	inter := last.Interface()
	var e error
	if sl, ok := inter.(SessionLoader); ok {
		e = Manager.Get(si.id).Get(sl.SessionKey(), inter)
	}

	return last, e
}

func (si *SessionInjector) When(abs interface{}) bool {
	if t, o := abs.(reflect.Type); o {
		return t.Implements(slType)
	}

	_, ok := abs.(SessionLoader)

	return ok
}

func NewSessionInjector(id string) *SessionInjector {
	return &SessionInjector{id: id}
}
