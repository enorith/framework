package container

import (
	"reflect"
)

type FieldTag map[string]reflect.StructTag

var (
	CacheTags map[string]FieldTag
)

func GetTags(i interface{}) FieldTag {
	typ := reflect.TypeOf(i)

	if v, ok := CacheTags[typ.String()]; ok {
		return v
	}

	ft := make(FieldTag)
	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)
		tag := field.Tag
		if !field.Anonymous {
			ft[field.Name] = tag
		}
	}

	return cacheTag(typ, ft)
}

func cacheTag(typ reflect.Type, ft FieldTag) FieldTag {
	CacheTags[typ.String()] = ft

	return ft
}

func init() {
	CacheTags = make(map[string]FieldTag)
}
