package http

import (
	"bytes"
	"fmt"
	"github.com/enorith/framework/container"
	"github.com/enorith/framework/exception"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contract"
	"github.com/enorith/framework/http/errors"
	"github.com/enorith/framework/http/router"
	"github.com/enorith/framework/kernel"
	"github.com/valyala/fasthttp"
	"log"
	"net/http"
	"reflect"
	"time"
)

type handlerType int

const (
	HandlerFastHttp handlerType = iota
	HandlerNetHttp
)

//RequestMiddleware request middleware
type RequestMiddleware interface {
	Handle(r contract.RequestContract, next PipeHandler) contract.ResponseContract
}

type MiddlewareGroup map[string][]RequestMiddleware

func timeMic() int64 {
	return time.Now().UnixNano() / int64(time.Microsecond)
}

type Kernel struct {
	wrapper         *router.Wrapper
	middleware      []RequestMiddleware
	middlewareGroup map[string][]RequestMiddleware
	app             *kernel.Application
	errorHandler    errors.ErrorHandler
	tcpKeepAlive    bool
	RequestCurrency int
	OutputLog       bool
	Handler         handlerType
}

func (k *Kernel) Wrapper() *router.Wrapper {
	return k.wrapper
}

func (k *Kernel) handleFunc(f func() (request contract.RequestContract, code int)) {
	defer k.app.Terminate()
	var start int64
	if k.OutputLog {
		start = timeMic()
	}
	request, code := f()

	if k.OutputLog {
		end := timeMic()

		log.Printf("/ %s - [%s] %s '%s' (%d) <%.3fms>", request.GetClientIp(),
			request.GetMethod(), request.GetUri(), request.GetContent(), code, float64(end-start)/1000)
	}
}

func (k *Kernel) ServeHTTP(w http.ResponseWriter, r *http.Request) {

	k.handleFunc(func() (request contract.RequestContract, code int) {
		request = content.NewNetHttpRequest(r, w)
		resp := k.Handle(request)

		if resp != nil {
			if k.tcpKeepAlive {
				resp.SetHeader("Connection", "keep-alive")
			}

			headers := resp.Headers()
			if headers != nil {
				for k, v := range headers {
					w.Header().Set(k, v)
				}
			}
			if !resp.Handled() {
				// call after set headers, before write body
				w.WriteHeader(resp.StatusCode())
			}
			body := resp.Content()
			if tp, ok := resp.(*content.TemplateResponse); ok {
				temp := tp.Template()
				temp.Execute(w, tp.TemplateData())
			} else if tp, ok := resp.(*content.File); ok {
				http.ServeFile(w, r, tp.Path())
			} else if body != nil {
				w.Write(body)
			}
			code = resp.StatusCode()
		}

		return
	})
}

func (k *Kernel) FastHttpHandler(ctx *fasthttp.RequestCtx) {
	k.handleFunc(func() (request contract.RequestContract, code int) {
		request = content.NewFastHttpRequest(ctx)
		resp := k.Handle(request)

		if k.tcpKeepAlive {
			resp.SetHeader("Connection", "keep-alive")
		}

		ctx.Response.SetStatusCode(resp.StatusCode())
		if resp.Headers() != nil {
			for k, v := range resp.Headers() {
				ctx.Response.Header.Set(k, v)
			}
		}
		if tp, ok := resp.(*content.TemplateResponse); ok {
			temp := tp.Template()
			temp.Execute(ctx, tp.TemplateData())
		} else if tp, ok := resp.(*content.File); ok {
			fasthttp.ServeFile(ctx, tp.Path())
		} else {
			body := resp.Content()
			buf := bytes.NewBuffer(body)

			fmt.Fprint(ctx, buf)
		}
		code = resp.StatusCode()

		return
	})
}

func (k *Kernel) SetMiddlewareGroup(middlewareGroup map[string][]RequestMiddleware) {
	k.middlewareGroup = middlewareGroup
}

func (k *Kernel) SetMiddleware(ms []RequestMiddleware) {
	k.middleware = ms
}

func (k *Kernel) KeepAlive(b ...bool) *Kernel {
	if len(b) > 0 {
		k.tcpKeepAlive = b[0]
	} else {
		k.tcpKeepAlive = true
	}
	return k
}

func (k *Kernel) IsKeepAlive() bool {
	return k.tcpKeepAlive
}

func (k *Kernel) SetErrorHandler(handler errors.ErrorHandler) {
	k.errorHandler = handler
}

func (k *Kernel) Handle(r contract.RequestContract) (resp contract.ResponseContract) {
	defer func() {
		if x := recover(); x != nil {
			resp = k.errorHandler.HandleError(x, r)
		}
	}()

	resp = k.SendRequestToRouter(r)

	if t, ok := resp.(*content.ErrorResponse); ok {
		resp = k.errorHandler.HandleError(t.E(), r)
	}

	if t, ok := resp.(exception.Exception); ok {
		resp = k.errorHandler.HandleError(t, r)
	}

	return resp
}

func (k *Kernel) SendRequestToRouter(r contract.RequestContract) contract.ResponseContract {
	pipe := new(Pipeline)
	pipe.Send(r)
	for _, m := range k.middleware {
		pipe.ThroughMiddleware(m)
	}
	p := k.wrapper.Match(r)
	if !p.IsValid() {
		return content.NotFoundResponse("not found")
	}
	if mid := p.Middleware(); mid != nil {
		for _, v := range mid {
			if ms, exists := k.middlewareGroup[v]; exists {
				for _, md := range ms {
					pipe.ThroughMiddleware(md)
				}
			}
		}
	}

	return pipe.Then(func(r contract.RequestContract) contract.ResponseContract {
		//resp := k.wrapper.Dispatch(r)
		return p.Handler()(r)
	})
}

func NewKernel(app *kernel.Application) *Kernel {
	k := new(Kernel)
	k.wrapper = router.NewWrapper(app)
	k.wrapper.ResolveRequest(KernelRequestResolver{})
	k.errorHandler = &errors.StandardErrorHandler{
		App: app,
	}
	k.app = app
	k.RequestCurrency = fasthttp.DefaultConcurrency
	return k
}

type KernelRequestResolver struct {
}

func (rr KernelRequestResolver) ResolveRequest(r contract.RequestContract, runtime *kernel.Application) {
	runtime.RegisterSingleton(r)
	runtime.Singleton("contract.RequestContract", r)

	runtime.BindFunc(&content.Request{}, func(c *container.Container) reflect.Value {

		return reflect.ValueOf(&content.Request{RequestContract: r})
	}, false)

	runtime.InitializeCondition(func(abs interface{}) bool {
		// dependency is sub struct of content.Request
		ts, _ := ofStruct(typeOf(abs))
		if ts.NumField() > 0 {
			tp := ts.Field(0).Type.String()

			s := reflect.TypeOf(&content.Request{}).String()
			return ts.NumField() > 0 && tp == s
		}

		return false
	}, func(abs interface{}, last reflect.Value) reflect.Value {
		// dependency injection sub struct of content.Request
		ts, _ := ofStruct(typeOf(abs))
		value := reflect.New(ts)
		ft := ts.Field(0).Type
		request := runtime.Instance(&content.Request{}).Interface().(*content.Request)
		instanceReq := runtime.Instance(ft)
		value.Elem().Field(0).Set(instanceReq)
		for i := 1; i < ts.NumField(); i++ {
			f := ts.Field(i)
			input := f.Tag.Get("input")
			if input != "" {
				parseInput(input, f.Type, value.Elem().Field(i), request)
			} else {
				param := f.Tag.Get("param")
				if param != "" {
					parseParam(input, f.Type, value.Elem().Field(i), request)
				}
			}
		}

		return value
	})
}
func parseParam(input string, fieldType reflect.Type, field reflect.Value, request *content.Request) {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(request.Param(input))
	case reflect.Int:
	case reflect.Int32:
	case reflect.Int64:
		in, _ := request.ParamInt64(input)
		field.SetInt(in)
	case reflect.Uint:
	case reflect.Uint32:
	case reflect.Uint64:
		in, _ := request.ParamUint64(input)
		field.SetUint(in)
	}
}

func parseInput(input string, fieldType reflect.Type, field reflect.Value, request *content.Request) {
	switch fieldType.Kind() {
	case reflect.String:
		field.SetString(request.GetString(input))
	case reflect.Int:
	case reflect.Int32:
	case reflect.Int64:
		in, _ := request.GetInt64(input)
		field.SetInt(in)
	case reflect.Uint:
	case reflect.Uint32:
	case reflect.Uint64:
		in, _ := request.GetUint64(input)
		field.SetUint(in)
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
