package database

import (
	"reflect"

	"github.com/enorith/database/orm"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/supports/reflection"
)

var (
	typeModel reflect.Type
)

type Injector struct {
	runtime *kernel.Application
}

func (i Injector) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {
	var value reflect.Value
	var ts reflect.Type
	// req := i.runtime.Instance("contracts.RequestContract").Interface().(contracts.RequestContract)

	if last.IsValid() {
		value = last
	} else {
		value = reflect.New(ts)
	}

	t := reflection.TypeOf(abs)
	ts = reflection.StructType(t)

	value = reflect.New(ts)
	value.CanSet()

	if t.Kind() == reflect.Struct {
		return value.Elem(), nil
	}

	return value, nil
}

func (i Injector) When(abs interface{}) bool {

	ts := reflection.StructType(abs)

	if ts.Kind() == reflect.Struct {
		for i := 0; i < ts.NumField(); i++ {
			if ts.Field(i).Anonymous {
				t := reflection.StructType(ts.Field(i).Type)
				if t == typeModel {
					return true
				}
			}
		}
	}

	return false
}

func init() {
	typeModel = reflection.StructType(orm.Model{})
}
