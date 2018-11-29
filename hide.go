package goblet

import (
	"reflect"
	"strings"
)

func autoHide(data interface{}) interface{} {

	if data == nil {
		return nil
	}

	origin := reflect.ValueOf(data)
	var val reflect.Value
	if origin.Kind() == reflect.Ptr {
		val = origin.Elem()
	} else {
		val = origin
	}

	valtype := val.Type()

	if val.Kind() == reflect.Struct {
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)
			v := val.Field(i)
			tags := strings.Split(string(t.Tag), ",")
			typeName := strings.ToLower(t.Name)
			if (typeName == "pwd" || typeName == "password") && v.CanSet() {
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
	if origin.Kind() == reflect.Ptr {
		return val.Addr().Interface()
	}
	return val.Interface()
}
