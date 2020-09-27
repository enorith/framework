package container

import (
	"fmt"
	"github.com/enorith/supports/reflection"
	"reflect"
)

// InstanceRegister register instance for container
type InstanceRegister func(c *Container) reflect.Value

// Injector interface for conditional initializer
type Injector interface {
	Injection(abs interface{}, last reflect.Value) (reflect.Value, error)
	When(abs interface{}) bool
}

// InjectionFunc injection function
type InjectionFunc func(abs interface{}, last reflect.Value) (reflect.Value, error)

type injectionChain []InjectionFunc

// ConditionInjectionFunc  conditional injection function
type ConditionInjectionFunc func(abs interface{}) bool

func (ic injectionChain) do(abs interface{}) (va reflect.Value, e error) {
	ts := reflection.StructType(abs)
	va = reflect.New(ts)

	for _, v := range ic {
		va, e = v(abs, va)
		if e != nil {
			return
		}
	}

	return
}

func conditionInjectionFunc(requireAbs interface{}, i InjectionFunc) InjectionFunc {

	return func(abs interface{}, last reflect.Value) (reflect.Value, error) {
		if f, ok := requireAbs.(ConditionInjectionFunc); ok {
			if f(abs) {
				return i(abs, last)
			}
		} else {
			as := reflection.StructType(abs)
			rs := reflection.StructType(requireAbs)

			if as == rs {
				return i(abs, last)
			}
		}

		return last, nil
	}
}

type Container struct {
	registers map[string]InstanceRegister

	singletons map[string]bool

	resolved map[string]reflect.Value

	injectionChain injectionChain
}

func (c *Container) WithInjector(h Injector) *Container {
	return c.InjectionWith(conditionInjectionFunc(ConditionInjectionFunc(h.When), h.Injection))
}

func (c *Container) InjectionWith(i InjectionFunc) *Container {
	c.injectionChain = append(c.injectionChain, i)
	return c
}

func (c *Container) InjectionRequire(requireAbs interface{}, i InjectionFunc) *Container {
	return c.InjectionWith(conditionInjectionFunc(requireAbs, i))
}

func (c *Container) InjectionCondition(f ConditionInjectionFunc, i InjectionFunc) *Container {
	return c.InjectionWith(conditionInjectionFunc(f, i))
}

func (c *Container) Bind(abs, instance interface{}, singleton bool) {

	if instance == nil {
		instance = abs
	}

	typ := reflection.TypeString(abs)

	c.registers[typ] = c.getResolver(instance)
	c.singletons[typ] = singleton
}

func (c *Container) BindFunc(abs interface{}, register InstanceRegister, singleton bool) {
	c.Bind(abs, register, singleton)
}

func (c *Container) Register(instance interface{}, singleton bool) {
	c.Bind(instance, nil, singleton)
}

func (c *Container) RegisterSingleton(instance interface{}) {
	c.Bind(instance, nil, true)
}

func (c *Container) Singleton(abs interface{}, instance interface{}) {
	c.Bind(abs, instance, true)
}

func (c *Container) IsSingleton(abs interface{}) bool {
	typ := reflection.TypeString(abs)

	if v, ok := c.singletons[typ]; ok {
		return v
	}
	return false
}

func (c *Container) MethodCall(abs interface{}, method string, params ...interface{}) ([]reflect.Value, error) {
	instance, e := c.Instance(abs)
	if e != nil {
		return nil, e
	}

	if !instance.IsValid() {
		return nil, fmt.Errorf("invalid method for type %v method [%s]", reflect.TypeOf(abs), method)
	}

	m := instance.MethodByName(method)

	return c.Invoke(m, params)
}

func (c *Container) getResolver(instance interface{}) InstanceRegister {
	var r func(c *Container) reflect.Value

	if t, ok := instance.(reflect.Type); ok {
		r = func(c *Container) reflect.Value {
			return reflect.New(t).Elem()
		}
	} else if t, ok := instance.(InstanceRegister); ok {
		r = t
	} else {
		r = func(c *Container) reflect.Value {
			return reflect.ValueOf(instance)
		}
	}

	return r
}

func (c *Container) Instance(abs interface{}, params ...interface{}) (reflect.Value, error) {
	var instance reflect.Value
	var e error
	fallback := func() {
		var va reflect.Value
		va, e = c.injectionChain.do(abs)

		if va.IsValid() {
			instance = va
		} else {
			instance = reflect.Value{}
		}
	}

	if t, ok := abs.(string); ok {
		if r := c.getResolve(t); r.IsValid() {
			instance = r
		} else {
			fallback()
		}
	} else if t, ok := abs.(reflect.Type); ok {
		str := t.String()
		if r := c.getResolve(str); r.IsValid() {
			instance = r
		} else {
			fallback()
		}

	} else if c.Bound(abs) {
		str := reflection.TypeString(abs)
		if r := c.getResolve(str); r.IsValid() {
			instance = r
		}
	} else {
		fallback()
	}

	for k, p := range params {
		if tp, ok := p.(reflect.Value); ok {
			instance.Field(k).Set(tp)
		} else {
			instance.Elem().Field(k).Set(reflect.ValueOf(p))
		}
	}

	return instance, e
}

func (c *Container) InstanceFor(abs interface{}, out interface{}, params ...interface{}) error {
	v, e := c.Instance(abs)
	if e != nil {
		return e
	}

	o := reflect.ValueOf(out)

	if !o.IsValid() {
		return fmt.Errorf("instance for abstact [%s]", reflection.TypeString(abs))
	}

	if o.Kind() == reflect.Ptr {
		o = o.Elem()
	}

	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}

	o.Set(v)

	return nil
}

func (c *Container) Invoke(f interface{}, params ...interface{}) ([]reflect.Value, error) {
	var t reflect.Type
	var fun reflect.Value

	if typ, ok := f.(reflect.Value); ok {
		t = typ.Type()
		fun = typ
	} else {
		t = reflect.TypeOf(f)
		fun = reflect.ValueOf(f)
	}

	var in = make([]reflect.Value, t.NumIn())
	for i := 0; i < t.NumIn(); i++ {
		argType := t.In(i)
		param, e := c.Instance(argType)
		if e != nil {
			return nil, e
		}

		if !param.IsValid() {
			return nil, fmt.Errorf("inject %v failed, parameter [%d] of type %v is invalid", t.String(), i, argType)
		}
		in[i] = param
	}

	return fun.Call(in), nil
}

func (c *Container) GetRegisters() map[string]InstanceRegister {
	return c.registers
}

func (c *Container) getResolve(abs string) reflect.Value {
	if resolved, ok := c.resolved[abs]; ok {
		return resolved
	}

	if resolver, o := c.registers[abs]; o {
		instance := resolver(c)

		if _, r := c.resolved[abs]; r && c.IsSingleton(abs) {
			c.resolved[abs] = instance
		}

		return instance
	}
	return reflect.Value{}
}

func (c *Container) Bound(abs interface{}) bool {
	s := reflection.TypeString(abs)
	_, o := c.registers[s]

	return o
}

func (c *Container) Init() {
	if c.registers == nil {
		c.registers = map[string]InstanceRegister{}
	}
	if c.singletons == nil {
		c.singletons = map[string]bool{}
	}
	if c.resolved == nil {
		c.resolved = map[string]reflect.Value{}
	}
}

func New() *Container {
	c := &Container{}
	c.Init()

	return c
}
