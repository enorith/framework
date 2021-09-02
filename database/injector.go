package database

import (
	"fmt"
	"reflect"
	"strings"

	"github.com/enorith/container"
	"github.com/enorith/gormdb"
	"github.com/enorith/http/contracts"
	"github.com/enorith/supports/reflection"
)

var ModelType = reflect.TypeOf(Model{})
var DestSliceType = reflect.TypeOf([]interface{}{})

type Injector struct {
	r   contracts.RequestContract
	c   Config
	ioc container.Interface
}

func (ij Injector) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {

	index := reflection.SubStructOf(abs, ModelType)
	modelV := reflect.Indirect(last).Field(index)
	modelE := reflect.Indirect(modelV)
	if ij.c.ImplicitInjection && len(ij.r.Params()) > 0 {
		model := parseModelName(abs)
		id, e := ij.r.ParamInt64(strings.ToLower(model))
		if e == nil {
			tx, e := gormdb.DefaultManager.GetConnection()

			if e != nil {
				return last, fmt.Errorf("[database] implicit injection model [%s] error %s", model, e.Error())
			}
			e = tx.First(last.Interface(), id).Error

			if e != nil {
				return last, fmt.Errorf("[database] implicit injection model [%s] error %s", model, e.Error())
			}
		}
	}

	modelI, e := ij.ioc.Instance(modelE.Type())
	modelE.Set(modelI)
	dv := modelE.FieldByName("Dest")
	dsv := modelE.FieldByName("DestSlice")
	dv.Set(last)
	sliceVal := reflect.New(reflect.SliceOf(last.Type()))
	dsv.Set(sliceVal)

	return last, e
}

func (ij Injector) When(abs interface{}) bool {
	return reflection.SubStructOf(abs, ModelType) > -1
}

func parseModelName(abs interface{}) string {
	str := reflection.TypeString(abs)

	parts := strings.Split(str, ".")

	return parts[len(parts)-1]
}
