package container

import (
	"fmt"
	"reflect"
)

// InstanceRegister register instance for container
type InstanceRegister func(c *Container) reflect.Value

// InitializeHandler interface for conditional initializer
type InitializeHandler interface {
	Initialize(abs interface{}, last reflect.Value) reflect.Value
	When(abs interface{}) bool
}

// Initializer initializer function
type Initializer func(abs interface{}, last reflect.Value) reflect.Value

type initializerChain []Initializer

// InitializerConditionFunc initializer conditional function
type InitializerConditionFunc func(abs interface{}) bool

func (ic initializerChain) do(abs interface{}) reflect.Value {
	var va reflect.Value
	for _, v := range ic {
		va = v(abs, va)
	}

	return va
}

func conditionInitializer(requireAbs interface{}, i Initializer) Initializer {

	return func(abs interface{}, last reflect.Value) reflect.Value {
		if f, ok := requireAbs.(InitializerConditionFunc); ok {
			if f(abs) {
				return i(abs, last)
			}
		} else {
			as := TypeString(abs)
			rs := TypeString(requireAbs)

			if as == rs {
				return i(abs, last)
			}
		}

		return reflect.Value{}
	}
}

type Container struct {
	registers map[string]InstanceRegister

	singletons map[string]bool

	resolved map[string]reflect.Value

	chain initializerChain
}

func (c *Container) HandleInitialize(h InitializeHandler) *Container {
	return c.InitializeWith(conditionInitializer(InitializerConditionFunc(h.When), h.Initialize))
}

func (c *Container) InitializeWith(i Initializer) *Container {
	c.chain = append(c.chain, i)

	return c
}

func (c *Container) InitializeRequire(requireAbs interface{}, i Initializer) *Container {

	return c.InitializeWith(conditionInitializer(requireAbs, i))
}

func (c *Container) InitializeCondition(f InitializerConditionFunc, i Initializer) *Container {
	return c.InitializeWith(conditionInitializer(f, i))
}

func (c *Container) Bind(abs, instance interface{}, singleton bool) {

	if instance == nil {
		instance = abs
	}

	typ := TypeString(abs)

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
	typ := TypeString(abs)

	if v, ok := c.singletons[typ]; ok {
		return v
	}
	return false
}

func (c *Container) MethodCall(abs interface{}, method string, params ...interface{}) ([]reflect.Value, error) {
	instance := c.Instance(abs)

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

func (c *Container) Instance(abs interface{}, params ...interface{}) reflect.Value {
	var instance reflect.Value
	fallback := func() {
		va := c.chain.do(abs)
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
		str := TypeString(abs)
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

	return instance
}

func (c *Container) InstanceFor(abs interface{}, out interface{}, params ...interface{}) error {
	v := c.Instance(abs)

	o := reflect.ValueOf(out)

	if !o.IsValid() {
		return fmt.Errorf("instance for abstact [%s]", TypeString(abs))
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
		param := c.Instance(argType)
		if !param.IsValid() {
			return nil, fmt.Errorf("value not found for type %v", argType)
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
	s := TypeString(abs)
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
