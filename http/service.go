package http

import (
	"github.com/enorith/framework/container"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/kernel"
	"reflect"
)

type Service struct {
}

func (s *Service) Register(app *kernel.Application) {
	app.BindFunc(&content.Request{}, func(c *container.Container) reflect.Value {

		return app.Instance(&content.Request{})
	}, false)

	app.InitializeCondition(func(abs interface{}) bool {
		var t reflect.Type
		if typ, ok := abs.(reflect.Type); ok {
			t = typ
		} else {
			t = reflect.TypeOf(abs)
		}

		ts, _ := ofStruct(t)
		if ts.NumField() > 0 {
			tp := ts.Field(0).Type.String()

			s := reflect.TypeOf(&content.Request{}).String()
			return ts.NumField() > 0 && tp == s
		}

		return false
	}, func(abs interface{}, last reflect.Value) reflect.Value {
		var t reflect.Type
		if typ, ok := abs.(reflect.Type); ok {
			t = typ
		} else {
			t = reflect.TypeOf(abs)
		}

		ts, _ := ofStruct(t)
		va := reflect.New(ts)
		ty := ts.Field(0).Type.String()

		req := app.Instance(&content.Request{}).Interface().(*content.Request)

		instanceReq := app.Instance(ty)
		va.Elem().Field(0).Set(instanceReq)

		for i := 1; i < ts.NumField(); i++ {
			f := ts.Field(i)
			input := f.Tag.Get("input")
			if input != "" {
				switch f.Type.Kind() {
				case reflect.String:
					va.Elem().Field(i).SetString(req.GetString(input))
				}

			}
		}

		return va
	})
}

func (s *Service) Boot(app *kernel.Application) {

}
