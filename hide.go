package goblet

import (
	"reflect"
	"strings"
)

func autoHide(data interface{}) interface{} {

	val := reflect.ValueOf(data)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	valtype := val.Type()

	if val.Kind() == reflect.Struct {
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)
			v := val.Field(i)
			tags := strings.Split(string(t.Tag), ",")
			typeName := strings.ToLower(t.Name)
			if typeName == "pwd" || typeName == "password" {
				v.SetString("hidden")
				continue
			}
			for _, tag := range tags {
				if tag == "hide" {
					v.SetString("hidden")
				}
			}
		}
	}
	return val.Interface()
}
