package router

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/enorith/framework/exception"
	"github.com/enorith/framework/http/content"
	"github.com/enorith/framework/http/contract"
	"github.com/enorith/framework/kernel"
	"net/http"
	"reflect"
	"strconv"
	"strings"
)

var MethodSplitter = "@"

type CRUDHandler interface {
	Index(request contract.RequestContract) json.Marshaler
	Show(request contract.RequestContract, id int64) json.Marshaler
	Store(request contract.RequestContract) contract.ResponseContract
	Update(request contract.RequestContract, id int64) contract.ResponseContract
	Delete(request contract.RequestContract, id int64) contract.ResponseContract
}

type Handler interface {
	HandleRoute(r contract.RequestContract) contract.ResponseContract
}

//ResultHandler handle return result
type ResultHandler func(val []reflect.Value, err error) contract.ResponseContract

type GroupHandler func(r *Wrapper)

type RequestResolver interface {
	ResolveRequest(r contract.RequestContract, runtime *kernel.Application)
}

var DefaultResultHandler = func(val []reflect.Value, err error) contract.ResponseContract {
	if err != nil {
		return content.ErrResponseFromError(err, 500, nil)
	}

	if len(val) < 1 {
		return content.TextResponse("", 200)
	}

	if len(val) > 1 {
		e := val[1].Interface()
		er, isErr := e.(error) // assume second return value is an error
		if isErr && e != nil {
			return content.ErrResponseFromError(er, 500, nil)
		}
	}

	data := val[0].Interface()

	return convertResponse(data)
}

var invalidHandler RouteHandler = func(r contract.RequestContract) contract.ResponseContract {
	return content.ErrResponseFromError(fmt.Errorf("invalid route handler if [%s] %s",
		r.GetMethod(), r.GetPathBytes()), 500, nil)
}

type Wrapper struct {
	*router
	controllers     map[string]interface{}
	ResultHandler   ResultHandler
	app             *kernel.Application
	requestResolver RequestResolver
}

//BindControllers bind controllers
func (w *Wrapper) BindControllers(controllers map[string]interface{}) {
	w.controllers = controllers
}

//BindController bind single controller
func (w *Wrapper) BindController(name string, controller interface{}) {
	w.controllers[name] = controller
}

//RegisterAction register route with giving handler
// 	'handler' can be string(eg: "home@Index"，), RouteHandler
// 	or any func returns string,RequestContract,[]byte or JsonAble
//
func (w *Wrapper) RegisterAction(method int, path string, handler interface{}) *routesHolder {
	routeHandler, e := w.wrap(handler)
	if e != nil {
		routeHandler = invalidHandler
	}

	return w.Register(method, path, routeHandler)
}

func (w *Wrapper) Get(path string, handler interface{}) *routesHolder {
	return w.RegisterAction(GET, path, handler)
}

func (w *Wrapper) Post(path string, handler interface{}) *routesHolder {
	return w.RegisterAction(POST, path, handler)
}

func (w *Wrapper) Patch(path string, handler interface{}) *routesHolder {
	return w.RegisterAction(PATCH, path, handler)
}

func (w *Wrapper) Put(path string, handler interface{}) *routesHolder {
	return w.RegisterAction(PUT, path, handler)
}

func (w *Wrapper) Delete(path string, handler interface{}) *routesHolder {
	return w.RegisterAction(DELETE, path, handler)
}

func (w *Wrapper) Group(g GroupHandler, prefix string, middleware ...string) {
	tr := NewWrapper(w.app)
	g(tr)

	if strings.Index(prefix, "/") != 0 {
		prefix = "/" + prefix
	}

	for method, routes := range tr.routes {
		for _, p := range routes {
			p.path = prefix + p.path
			p.partials = resolvePartials(p.path)
			p.middleware = middleware
			w.routes[method] = append(w.routes[method], p)
		}
	}
}

//CRUD register simple crud routes
func (w *Wrapper) CRUD(path string, handler CRUDHandler, middleware ...string) {
	w.Group(func(r *Wrapper) {
		r.HandleGet("", func(r contract.RequestContract) contract.ResponseContract {
			return content.JsonResponse(handler.Index(r), 200, content.DefaultHeader())
		})
		r.HandleGet("/:id", func(r contract.RequestContract) contract.ResponseContract {
			id, _ := strconv.ParseInt(r.Param("id"), 10, 64)

			return content.JsonResponse(handler.Show(r, id), 200, content.DefaultHeader())
		})
		r.HandlePost("", func(r contract.RequestContract) contract.ResponseContract {
			return handler.Store(r)
		})
		r.HandlePut("/:id", func(r contract.RequestContract) contract.ResponseContract {
			id, _ := strconv.ParseInt(r.Param("id"), 10, 64)

			return handler.Update(r, id)
		})
		r.HandleDelete("/:id", func(r contract.RequestContract) contract.ResponseContract {
			id, _ := strconv.ParseInt(r.Param("id"), 10, 64)

			return handler.Delete(r, id)
		})
	}, path, middleware...)
}

func (w *Wrapper) ResolveRequest(rs RequestResolver) {
	w.requestResolver = rs
}

func (w *Wrapper) parseController(s string) (c string, m string) {
	partials := strings.SplitN(s, MethodSplitter, 2)
	ctrl := partials[0]
	var method string
	if len(partials) > 1 {
		method = partials[1]
	} else {
		method = "Index"
	}

	return ctrl, method
}

func (w *Wrapper) wrap(handler interface{}) (RouteHandler, error) {
	if t, ok := handler.(Handler); ok {
		return t.HandleRoute, nil
	}

	if t, ok := handler.(RouteHandler); ok { // router handler
		return t, nil
	}

	if h, isHandler := handler.(http.Handler); isHandler { // raw handler
		return NewRouteHandlerFromHttp(h), nil
	} else if t, ok := handler.(string); ok { // string controller handler
		name, method := w.parseController(t)
		controller, exists := w.controllers[name]
		if !exists {
			panic(fmt.Sprintf("panic: router: controller [%s] not registered", name))
		}
		return func(req contract.RequestContract) contract.ResponseContract {
			runtime := w.getRuntimeApp(req)
			val, err := runtime.MethodCall(controller, method)
			return w.handleResult(val, err)
		}, nil
	} else if reflect.TypeOf(handler).Kind() == reflect.Func { // function
		return func(req contract.RequestContract) contract.ResponseContract {
			runtime := w.getRuntimeApp(req)
			val, err := runtime.Invoke(handler)
			return w.handleResult(val, err)
		}, nil
	}
	panic(fmt.Sprintf("panic: router handler expect string or func, %s giving", reflect.TypeOf(handler).Kind()))
}

func (w *Wrapper) handleResult(val []reflect.Value, err error) contract.ResponseContract {

	if w.ResultHandler == nil {
		return DefaultResultHandler(val, err)
	}

	return w.ResultHandler(val, err)
}

func NewRouteHandlerFromHttp(h http.Handler) RouteHandler {
	return func(req contract.RequestContract) contract.ResponseContract {
		if request, ok := req.(*content.NetHttpRequest); ok {
			return NetHttpHandlerFromHttp(request, h)
		} else if request, ok := req.(*content.FastHttpRequest); ok {
			return FastHttpHandlerFromHttp(request, h)
		}

		return content.ErrResponseFromError(errors.New("invalid handler giving"), 500, nil)
	}
}

func (w *Wrapper) getRuntimeApp(req contract.RequestContract) *kernel.Application {
	runtime := w.app.NewRuntime()
	w.requestResolver.ResolveRequest(req, runtime)

	for _, v := range w.app.RuntimeRegisters() {
		runtime.Bind(v.Abs(), v.Instance(), v.Singleton())
	}

	runtime.RegisterSingleton(w.app)
	runtime.Singleton(&kernel.Application{}, runtime)

	return runtime
}

func convertResponse(data interface{}) contract.ResponseContract {

	if t, ok := data.(error); ok { // return error
		return content.ErrResponseFromError(t, 500, nil)
	} else if t, ok := data.(string); ok { // return string
		return content.TextResponse(t, 200)
	} else if t, ok := data.([]byte); ok { // return []byte
		return content.NewResponse(t, map[string]string{}, 200)
	} else if t, ok := data.(*content.ErrorResponse); ok { // return ErrorResponse
		return t
	} else if t, ok := data.(contract.ResponseContract); ok { // return Response
		return t
	} else if t, ok := data.(json.Marshaler); ok { // return json or error
		j, err := t.MarshalJSON()
		if err != nil {
			return content.ErrResponse(exception.NewExceptionFromError(err, 500), 500, nil)
		}
		return content.NewResponse(j, content.JsonHeader(), 200)
	} else if t, ok := data.(fmt.Stringer); ok { // return string
		return content.TextResponse(t.String(), 200)
	} else {
		// fallback to json
		return content.JsonResponse(data, 200, nil)
	}
}

func NewWrapper(app *kernel.Application) *Wrapper {
	r := &router{
		routes: func() map[string][]*paramRoute {
			rs := map[string][]*paramRoute{}
			for _, v := range methodMap {
				rs[v] = []*paramRoute{}
			}

			return rs
		}(),
	}

	return &Wrapper{r, nil, nil, app, defaultRequestResolver{}}
}

type defaultRequestResolver struct {
}

func (d defaultRequestResolver) ResolveRequest(r contract.RequestContract, runtime *kernel.Application) {
	runtime.RegisterSingleton(r)
	runtime.Singleton("contract.RequestContract", r)
}
