package database

import (
	"github.com/enorith/database/orm"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/supports/reflection"
	"reflect"
)

type Injector struct {
	runtime *kernel.Application
}

func (i Injector) Injection(abs interface{}, last reflect.Value) reflect.Value {
	var value reflect.Value
	var ts reflect.Type
	// req := i.runtime.Instance("contracts.RequestContract").Interface().(contracts.RequestContract)

	t := reflect.TypeOf(abs)
	if t.Kind() == reflect.Ptr {
		ts = t.Elem()
	} else {
		ts = t
	}

	value = reflect.New(ts)

	if t.Kind() == reflect.Struct {
		return value.Elem()
	}

	return value
}

func (i Injector) When(abs interface{}) bool {

	ts := reflection.StructType(abs)

	if ts.NumField() > 0 {
		tp := ts.Field(0).Type.String()

		s := reflect.TypeOf(&orm.Model{}).String()
		ss := reflect.TypeOf(orm.Model{}).String()
		return ts.NumField() > 0 && (tp == s || ss == tp)
	}

	return false
}
