package container

import "reflect"

func TypeString(abstract interface{}) (typeString string) {

	if t, ok := abstract.(string); ok {
		typeString = t
	} else if t, ok := abstract.(reflect.Type); ok {
		typeString = t.String()
	} else {
		typeString = reflect.TypeOf(abstract).String()
	}

	return
}
