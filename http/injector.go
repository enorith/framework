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
	"github.com/enorith/supports/reflection"
	"reflect"
	"sync"
)

var cs cacheStruct

var (
	typeRequest,
	typeParamInt64,
	typeParamString,
	typeParamInt,
	typeParamUnit reflect.Type
)

type input interface {
	Get(key string) []byte
	Param(key string) string
	ParamBytes(key string) []byte
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

func (j jsonInput) ParamBytes(key string) []byte {
	return nil
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

//RequestInjector inject request object, with validation
type RequestInjector struct {
	runtime    *kernel.Application
	request    contracts.RequestContract
	validator  *validation.Validator
	paramIndex int
}

func (r RequestInjector) Injection(abs interface{}, last reflect.Value) (reflect.Value, error) {
	var value reflect.Value
	var e error
	defer func() {
		if x := recover(); x != nil {
			value = reflect.Value{}
			if err, ok := x.(error); ok {
				e = err
			}
		}
	}()
	t := typeOf(abs)
	ts := reflection.StructType(abs)

	//
	if last.IsValid() {
		value = last
	} else {
		value = reflect.New(ts)
	}

	//value = last
	if r.isRequest(abs) {
		// dependency injection sub struct of content.Request

		for i := 0; i < ts.NumField(); i++ {
			tf := ts.Field(i).Type
			if reflection.StructType(tf) == typeRequest {
				instanceReq, err := r.runtime.Instance(tf)
				if err != nil {
					return value, err
				}
				value.Elem().Field(i).Set(instanceReq)
				break
			}
		}

		r.parseStruct(ts, value, r.request, 1)

		if t.Kind() == reflect.Struct {
			return value.Elem(), nil
		}

		return value, nil
	} else if r.isParam(abs) {
		// parameter injection
		params := r.request.ParamsSlice()
		paramsLength := len(params)
		if paramsLength > r.paramIndex {
			param := params[r.paramIndex]
			if ts == typeParamInt64 || ts == typeParamInt {
				val, err := byt.ToInt64(param)
				if err != nil {

					return value, err
				}
				value.Elem().SetInt(val)
			} else if ts == typeParamUnit {

				val, err := byt.ToUint64(param)
				if err != nil {

					return value, err
				}
				value.Elem().SetUint(val)
			} else if ts == typeParamString {
				val := byt.ToString(param)

				value.Elem().SetString(val)
			}

			r.paramIndex++
		}

		if t.Kind() == reflect.Ptr {
			return value, nil
		}

		return value.Elem(), nil
	}

	return value, e
}

func (r RequestInjector) When(abs interface{}) bool {
	ok, e := cs.get(abs)
	if e {
		return ok
	}

	// dependency is sub struct of content.Request
	is := r.isParam(abs) || r.isRequest(abs)
	cs.set(abs, is)

	return is
}

func (r RequestInjector) isParam(abs interface{}) bool {
	ts := reflection.StructType(abs)

	return ts == typeParamInt || ts == typeParamString || ts == typeParamInt64 || ts == typeParamUnit
}

func (r RequestInjector) isRequest(abs interface{}) bool {
	ts := reflection.StructType(abs)

	if ts.Kind() == reflect.Struct {
		for i := 0; i < ts.NumField(); i++ {
			if ts.Field(i).Anonymous {
				t := reflection.StructType(ts.Field(i).Type)
				if t == typeRequest {
					return true
				}
			}
		}
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
				r.parseField(f.Type, fieldValue, request.ParamBytes(param))
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
	v := field.Interface()
	if _, ok := v.([]byte); ok {
		field.SetBytes(data)
		return
	}

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
	case reflect.Float32:
		fallthrough
	case reflect.Float64:
		in, _ := byt.ToFloat64(data)
		field.SetFloat(in)
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

	return
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

func init() {
	typeParamInt64 = reflection.StructType(content.ParamInt64(42))
	typeParamString = reflection.StructType(content.Param("42"))
	typeParamUnit = reflection.StructType(content.ParamUint64(42))
	typeParamInt = reflection.StructType(content.ParamInt(42))
	typeRequest = reflection.StructType(content.Request{})
}
