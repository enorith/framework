package container_test

import (
	"fmt"
	"github.com/enorith/supports/reflection"
	"reflect"
	"testing"

	"github.com/enorith/framework/container"
)

type foo struct {
	name string
}

func TestContainer_Bind(t *testing.T) {
	bt := bindTable()
	c := container.New()
	for _, v := range bt {
		t.Run(v.name, func(t *testing.T) {
			c.Bind(v.abs, v.instance, v.singleton)

			if !c.Bound(v.abs) {
				t.Fatalf("bind failed of %v, instance %v", v.abs, v.instance)
			}
			if c.IsSingleton(v.abs) != v.singleton {
				t.Fatalf("bind singleton failed of %v, instance %v", v.abs, v.instance)
			}
		})
	}
}

func TestContainer_Instance(t *testing.T) {
	bt := bindTable()
	c := container.New()
	for _, v := range bt {
		t.Run(v.name, func(t *testing.T) {
			c.Bind(v.abs, v.instance, v.singleton)
			obj := c.Instance(v.abs)

			if !obj.IsValid() {
				t.Fatalf("instance of %v is invalid", v.abs)
			}
			if i, ok := obj.Interface().(*foo); ok {
				if i.name != v.name {
					t.Fatalf("instance of %v is invalid, object name %s != %s", v.abs, i.name, v.name)
				}
			} else {
				t.Fatalf("instance of %v is invalid, got object %v", v.abs, obj)
			}
		})
	}
}

func TestContainer_Invoke(t *testing.T) {
	c := container.New()
	t.Run("invoke func", func(t *testing.T) {
		outs, err := c.Invoke(funcBar)

		if err != nil {
			t.Fatalf("invoke func fail %s", err)
		}
		if b, ok := outs[0].Interface().(bool); ok {
			if !b {
				t.Fatalf("invoke func fail got %v", b)
			}
		} else {
			t.Fatalf("invoke func fail got %v", outs)
		}
	})

	t.Run("invoke func injection", func(t *testing.T) {
		c.BindFunc(&foo{}, func(c *container.Container) reflect.Value {
			return reflect.ValueOf(&foo{name: "test foo"})
		}, false)

		outs, err := c.Invoke(funcBarInjection)

		if err != nil {
			t.Fatalf("invoke func injection fail %s", err)
		}

		if b, ok := outs[0].Interface().(string); ok {
			if b != "test foo" {
				t.Fatalf("invoke func injection fail got %v", b)
			}
		} else {
			t.Fatalf("invoke func injection fail got %v", outs)
		}
	})
}

func TestTypeString(t *testing.T) {
	tt := []struct {
		name string
		abs  interface{}
		str  string
	}{
		{"struct type", foo{}, "container_test.foo"},
		{"ptr type", &foo{}, "*container_test.foo"},
		{"struct type type", reflect.TypeOf(foo{}), "container_test.foo"},
		{"ptr type type", reflect.TypeOf(&foo{}), "*container_test.foo"},
	}

	for _, v := range tt {
		t.Run(v.name, func(t *testing.T) {
			str := reflection.TypeString(v.abs)

			if str != v.str {
				t.Fatalf("type if %v expect string [%s], got [%s]", v.abs, v.str, str)
			}
		})
	}
}

func TestContainer_InstanceFor(t *testing.T) {
	c := container.New()

	c.BindFunc("foo", func(c *container.Container) reflect.Value {

		return reflect.ValueOf(&foo{"test name"})
	}, false)

	var f foo

	c.InstanceFor("foo", &f)
	if f.name != "test name" {
		t.Fatal("instance failed")
	}
	t.Log(f)
}

type InitializeHandler struct {
}

func (i InitializeHandler) Injection(abs interface{}, last reflect.Value) reflect.Value {

	return reflect.ValueOf(foo{"test foo"})
}

func (i InitializeHandler) When(abs interface{}) bool {

	str := reflection.TypeString(abs)

	fmt.Println(str)

	return str == "container_test.foo"
}

func TestContainer_HandleInitialize(t *testing.T) {
	c := container.New()

	c.HandleInitialize(InitializeHandler{})

	i := c.Instance("container_test.foo")
	t.Log(i)
	if !i.IsValid() {
		t.Fatal("instance failed")
	}
}

func funcBar() bool {

	return true
}

func funcBarInjection(f *foo) string {
	return f.name
}

func bindTable() []struct {
	name      string
	abs       interface{}
	instance  interface{}
	singleton bool
} {
	typ := reflect.TypeOf(&foo{})

	return []struct {
		name      string
		abs       interface{}
		instance  interface{}
		singleton bool
	}{
		{"string abs", "foo_s", &foo{name: "string abs"}, false},
		{"string abs singleton", "foo_ss", &foo{name: "string abs singleton"}, true},
		{"string abs func", "foo_s_f", container.InstanceRegister(func(c *container.Container) reflect.Value {
			return reflect.ValueOf(&foo{name: "string abs func"})
		}), false},
		{"type abs", typ, &foo{name: "type abs"}, false},
		{"object abs", &foo{}, &foo{name: "object abs"}, false},
	}
}
