package http

import (
	"fmt"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/kernel"
	"reflect"
	"sync"
)

var cs cacheStruct

type parseType int

const (
	parseTypeInput parseType = iota
	parseTypeParam
)

type cacheStruct struct {
	cache map[interface{}]bool
	mu    sync.RWMutex
}

func (c *cacheStruct) get(abs interface{}) (ok bool, exist bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	ok, exist = c.cache[abs]

	return
}

func (c *cacheStruct) set(abs interface{}, b bool) {
	c.mu.Lock()
	c.cache[abs] = b
	c.mu.Unlock()
}

type RequestInitializer struct {
	runtime *kernel.Application
}

func (r RequestInitializer) SetRuntime(runtime *kernel.Application) {
	r.runtime = runtime
}

func (r RequestInitializer) Initialize(abs interface{}, last reflect.Value) reflect.Value {
	// dependency injection sub struct of content.Request
	ts, _ := ofStruct(typeOf(abs))
	value := reflect.New(ts)

	req := ts.Field(0).Type
	request := r.runtime.Instance(&content.Request{}).Interface().(*content.Request)
	instanceReq := r.runtime.Instance(req)

	value.Elem().Field(0).Set(instanceReq)

	parseStruct(ts, value, request, 1)

	return value
}

func (r RequestInitializer) When(abs interface{}) bool {
	ok, e := cs.get(abs)
	if e {
		return ok
	}

	// dependency is sub struct of content.Request
	is := r.isRequest(abs)
	cs.set(abs, is)

	return is
}

func (r RequestInitializer) isRequest(abs interface{}) bool {
	ts, _ := ofStruct(typeOf(abs))
	if ts.NumField() > 0 {
		tp := ts.Field(0).Type.String()

		s := reflect.TypeOf(&content.Request{}).String()
		return ts.NumField() > 0 && tp == s
	}
	return false
}

func init() {
	cs = cacheStruct{cache: map[interface{}]bool{}, mu: sync.RWMutex{}}
}

func parseStruct(structType reflect.Type, newValue reflect.Value, request *content.Request, offset int) reflect.Value {

	for i := offset; i < structType.NumField(); i++ {
		f := structType.Field(i)
		fieldValue := newValue.Elem().Field(i)
		if f.Type.Kind() == reflect.Struct {
			value := parseStruct(f.Type, reflect.New(f.Type), request, 0)
			fieldValue.Set(value.Elem())
		} else {
			if input := f.Tag.Get("input"); input != "" {
				parseField(input, f.Type, fieldValue, request, parseTypeInput)
			} else if param := f.Tag.Get("param"); param != "" {
				parseField(param, f.Type, fieldValue, request, parseTypeParam)
			}
		}
	}

	return newValue
}

func parseField(input string, fieldType reflect.Type, field reflect.Value, request *content.Request, parseType parseType) {
	switch fieldType.Kind() {
	case reflect.String:
		if parseType == parseTypeParam {
			field.SetString(request.Param(input))
		} else {
			field.SetString(request.GetString(input))
		}
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		if parseType == parseTypeParam {
			in, _ := request.ParamInt64(input)
			field.SetInt(in)
		} else {
			in, _ := request.GetInt64(input)
			field.SetInt(in)
		}
	case reflect.Uint:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		if parseType == parseTypeParam {
			in, _ := request.ParamUint64(input)
			field.SetUint(in)
		} else {
			in, _ := request.GetUint64(input)
			field.SetUint(in)
		}
	}
}

func ofStruct(t reflect.Type) (reflect.Type, error) {
	if t.Kind() == reflect.Ptr {
		return t.Elem(), nil
	} else if t.Kind() == reflect.Struct {
		return t, nil
	}

	return nil, fmt.Errorf("invalid type to struct %v", t)
}

func typeOf(abs interface{}) (t reflect.Type) {
	if typ, ok := abs.(reflect.Type); ok {
		t = typ
	} else {
		t = reflect.TypeOf(abs)
	}
	return
}
