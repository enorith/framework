package config

import (
	"gopkg.in/yaml.v2"
	"io/ioutil"
)

type Config interface {
	GetValue(key string) (interface{}, bool)
	GetInt(key string) (int, bool)
	GetString(key string) (string, bool)
	GetBool(key string) (bool, bool)
}

type SimpleConfig struct {
	config map[string]interface{}
}

func (c *SimpleConfig) Load(file string) (*SimpleConfig, error) {
	data, err := ioutil.ReadFile(file)

	yaml.Unmarshal(data, &c.config)
	return c, err
}

func (c *SimpleConfig) GetValue(key string) (interface{}, bool) {
	if data, ok := c.config[key]; ok {
		return data, true
	}

	return nil, false
}

func (c *SimpleConfig) GetInt(key string) (int, bool) {
	if data, ok := c.GetValue(key); ok {
		if result, o := data.(int); o {
			return result, true
		}
		return 0, false
	}
	return 0, false
}

func (c *SimpleConfig) GetString(key string) (string, bool) {
	if data, ok := c.GetValue(key); ok {
		if result, o := data.(string); o {
			return result, true
		}
		return "", false
	}

	return "", false
}

func (c *SimpleConfig) GetBool(key string) (bool, bool) {
	if data, ok := c.GetValue(key); ok {
		if result, o := data.(bool); o {
			return result, true
		}
		return false, false
	}
	return false, false
}

func Load(file string) (*SimpleConfig, error) {
	c := &SimpleConfig{}

	return c.Load(file)
}

func LoadTo(file string, out interface{}) error {
	data, err := ioutil.ReadFile(file)

	if err != nil {
		return err
	}

	return yaml.Unmarshal(data, out)
}
