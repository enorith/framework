package cache

import (
	"time"

	c "github.com/enorith/cache"
)

func Has(key string) bool {
	return Default.Has(key)
}

func Get(key string, object interface{}) (c.Value, bool) {
	return Default.Get(key, object)
}

func Put(key string, data interface{}, d time.Duration) error {
	return Default.Put(key, data, d)
}

func Forever(key string, data interface{}) error {
	return Default.Forever(key, data)
}

func Remove(key string) bool {
	return Default.Remove(key)
}

func Increment(key string) bool {
	return Default.Increment(key)
}

func Decrement(key string) bool {
	return Default.Decrement(key)
}

func Add(key string, data interface{}, d time.Duration) bool {
	return Default.Add(key, data, d)
}

func Use(driver string) error {
	return Default.Use(driver)
}
