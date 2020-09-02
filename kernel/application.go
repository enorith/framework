package kernel

import (
	"github.com/enorith/framework/container"
	"github.com/enorith/framework/kernel/config"
	"github.com/enorith/supports/carbon"
	"path/filepath"
	"time"
)

const Version = "0.0.1"

var (
	Timezone    string
	TimezoneSet bool
)

type Cb func()

type ServiceProvider interface {
	// Register your service provider when app starting
	//
	Register(app *Application)
	// Boot your service provider when app started
	//
	Boot(app *Application)
}

type RuntimeRegisters struct {
	abs       interface{}
	instance  interface{}
	singleton bool
}

func (r *RuntimeRegisters) Singleton() bool {
	return r.singleton
}

func (r *RuntimeRegisters) Instance() interface{} {
	return r.instance
}

func (r *RuntimeRegisters) Abs() interface{} {
	return r.abs
}

type Application struct {
	container.Container
	providers  []ServiceProvider
	debug      bool
	env        string
	basePath   string
	runtime    []*RuntimeRegisters
	terminates []Cb
	defers     []Cb
}

func (a *Application) RuntimeRegisters() []*RuntimeRegisters {
	return a.runtime
}

func (a *Application) Env() string {
	return a.env
}

func (a *Application) Debug() bool {
	return a.debug
}

func (a *Application) RegisterController(name string, controller interface{}) {
	a.Bind("controller."+name, controller, false)
}

// BindRuntime bind request lift time object to container
// object can be injection to route handler
func (a *Application) BindRuntime(abs interface{}, instance interface{}, singleton bool) {
	a.Bind(abs, instance, singleton)
	a.runtime = append(a.runtime, &RuntimeRegisters{
		abs,
		instance,
		singleton,
	})
}

func (a *Application) BindRuntimeFunc(abs interface{}, register container.InstanceRegister, singleton bool) {
	a.BindRuntime(abs, register, singleton)
}

// TerminateWith run your function at end of request life time
func (a *Application) TerminateWith(f Cb) {
	a.terminates = append(a.terminates, f)
}

func (a *Application) Terminate() {
	for _, v := range a.terminates {
		v()
	}
}

/// Defer run your function before server down
func (a *Application) Defer(cb Cb) {
	a.defers = append(a.defers, cb)
}

func (a *Application) RunDefers() {
	for _, v := range a.defers {
		v()
	}
}

func (a *Application) Bootstrap() {
	a.registerAndBootProviders()
	a.setUpTimeLocation()
}

func (a *Application) setUpTimeLocation() {
	if TimezoneSet {
		return
	}

	loc, err := time.LoadLocation(Timezone)

	if err == nil {
		carbon.Timezone = loc
		TimezoneSet = true
	}
}
func (a *Application) registerAndBootProviders() {
	for _, v := range a.providers {
		v.Register(a)
	}

	for _, v := range a.providers {
		v.Boot(a)
	}
}

func (a *Application) RegisterServiceProvider(provider ServiceProvider) {
	a.providers = append(a.providers, provider)
}

func (a *Application) GetConfigPath() string {
	return filepath.Join(a.basePath, "config")
}

func (a *Application) GetBasePath() string {
	return a.basePath
}

func (a *Application) Configure(name string, to interface{}) error {
	path := filepath.Join(a.GetConfigPath(), name+".yml")
	return config.LoadTo(path, to)
}

func (a *Application) NewRuntime() *Application {
	return NewApp(a.Env(), a.Debug(), a.GetBasePath())
}

func NewApp(env string, debug bool, basePath string) *Application {
	app := &Application{}
	app.Init()
	app.providers = []ServiceProvider{}
	app.env = env
	app.debug = debug
	app.runtime = []*RuntimeRegisters{}
	app.terminates = []Cb{}
	app.basePath = basePath
	return app
}
