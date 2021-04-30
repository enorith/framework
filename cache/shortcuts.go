package cache

import (
	"time"

	. "github.com/enorith/cache"
)

func Has(key string) bool {
	return AppCache.Has(key)
}

func Get(key string, object interface{}) (Value, bool) {
	return AppCache.Get(key, object)
}

func Put(key string, data interface{}, d time.Duration) error {
	return AppCache.Put(key, data, d)
}

func Forever(key string, data interface{}) error {
	return AppCache.Forever(key, data)
}

func Remove(key string) bool {
	return AppCache.Remove(key)
}

func Increment(key string) bool {
	return AppCache.Increment(key)
}

func Decrement(key string) bool {
	return AppCache.Decrement(key)
}

func Add(key string, data interface{}, d time.Duration) bool {
	return AppCache.Add(key, data, d)
}

func Use(driver string) error {
	return AppCache.Use(driver)
}
