package http

import (
	"errors"
	"fmt"
	"github.com/buger/jsonparser"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contracts"
	"github.com/enorith/framework/http/validation"
	"github.com/enorith/framework/kernel"
	"github.com/enorith/supports/byt"
	"reflect"
	"sync"
)

var cs cacheStruct

type input interface {
	Get(key string) []byte
	Param(key string) string
	File(key string) (contracts.UploadFile, error)
}

type jsonInputHandler func(j jsonInput)

type jsonInput []byte

func (j jsonInput) Get(key string) []byte {
	value, _, _, _ := jsonparser.Get(j, key)

	return value
}

func (j jsonInput) Param(key string) string {
	return ""
}

func (j jsonInput) File(key string) (contracts.UploadFile, error) {
	return nil, errors.New("jsonInput does not implement func File(key string) (content.UploadFile, error)")
}

func (j jsonInput) Each(h jsonInputHandler) error {
	_, e := jsonparser.ArrayEach(j, func(value []byte, dataType jsonparser.ValueType, offset int, err error) {
		h(value)
	})

	return e
}

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

type RequestInjector struct {
	runtime   *kernel.Application
	request   contracts.RequestContract
	validator *validation.Validator
}

func (r RequestInjector) SetRuntime(runtime *kernel.Application) {
	r.runtime = runtime
}

func (r RequestInjector) Injection(abs interface{}, last reflect.Value) reflect.Value {
	var value reflect.Value
	// dependency injection sub struct of content.Request
	defer func() {
		if x := recover(); x != nil {
			value = reflect.Value{}
		}
	}()

	t := typeOf(abs)
	ts, e := ofStruct(t)

	if e != nil {
		return reflect.Value{}
	}

	value = reflect.New(ts)

	req := ts.Field(0).Type
	instanceReq := r.runtime.Instance(req)

	value.Elem().Field(0).Set(instanceReq)

	r.parseStruct(ts, value, r.request, 1)

	if t.Kind() == reflect.Struct {
		return value.Elem()
	}

	return value
}

func (r RequestInjector) When(abs interface{}) bool {
	ok, e := cs.get(abs)
	if e {
		return ok
	}

	// dependency is sub struct of content.Request
	is := r.isRequest(abs)
	cs.set(abs, is)

	return is
}

func (r RequestInjector) isRequest(abs interface{}) bool {
	ts, _ := ofStruct(typeOf(abs))
	if ts.NumField() > 0 {
		tp := ts.Field(0).Type.String()

		s := reflect.TypeOf(&content.Request{}).String()
		ss := reflect.TypeOf(content.Request{}).String()
		return ts.NumField() > 0 && (tp == s || ss == tp)
	}
	return false
}

func init() {
	cs = cacheStruct{cache: map[interface{}]bool{}, mu: sync.RWMutex{}}
}

func (r RequestInjector) parseStruct(structType reflect.Type, newValue reflect.Value, request input, offset int) reflect.Value {

	for i := offset; i < structType.NumField(); i++ {
		f := structType.Field(i)
		fieldValue := newValue.Elem().Field(i)
		if validated, ok := newValue.Interface().(validation.WithValidation); ok {
			validated.Rules()
		}
		if f.Type.Kind() == reflect.Struct && f.Anonymous {
			value := r.parseStruct(f.Type, reflect.New(f.Type), request, 0)
			fieldValue.Set(value.Elem())
		} else {
			if input := f.Tag.Get("input"); input != "" {
				r.parseField(f.Type, fieldValue, request.Get(input))
			} else if param := f.Tag.Get("param"); param != "" {
				r.parseField(f.Type, fieldValue, []byte(request.Param(param)))
			} else if file := f.Tag.Get("file"); file != "" {
				if f.Type.String() == "contracts.UploadFile" {
					uploadFile, e := request.File(file)
					if e == nil {
						fieldValue.Set(reflect.ValueOf(uploadFile))
					}
				}
			}
		}
	}

	return newValue
}

func (r RequestInjector) parseField(fieldType reflect.Type, field reflect.Value, data []byte) {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(byt.ToString(data))
	case reflect.Bool:
		in, _ := byt.ToBool(data)
		field.SetBool(in)
	case reflect.Int:
		fallthrough
	case reflect.Int32:
		fallthrough
	case reflect.Int64:
		in, _ := byt.ToInt64(data)
		field.SetInt(in)
	case reflect.Uint:
		fallthrough
	case reflect.Uint32:
		fallthrough
	case reflect.Uint64:
		in, _ := byt.ToUint64(data)
		field.SetUint(in)
	case reflect.Struct:
		in := r.parseStruct(fieldType, reflect.New(fieldType), jsonInput(data), 0)
		field.Set(in.Elem())
	case reflect.Ptr:
		in := r.parseStruct(fieldType, reflect.New(fieldType), jsonInput(data), 0)
		field.Set(in)
	case reflect.Slice:
		it := fieldType.Elem()
		var ivs []reflect.Value
		jsonInput(data).Each(func(j jsonInput) {
			iv := reflect.New(it).Elem()
			r.parseField(it, iv, j)
			ivs = append(ivs, iv)
		})
		l := len(ivs)
		slice := reflect.MakeSlice(fieldType, l, l)
		for index, v := range ivs {
			slice.Index(index).Set(v)
		}

		field.Set(slice)
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
