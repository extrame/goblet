package goblet

import (
	"reflect"
	"regexp"
	"strings"
)

var hideMatcher = regexp.MustCompile("^hide(\\((([^,]+|,)*)\\))*$")

func autoHide(data interface{}, ctx *Context) interface{} {

	if data == nil {
		return nil
	}

	origin := reflect.ValueOf(data)

	autoHideElem(origin, ctx.ReqURL().Path)

	return origin.Interface()
}

func autoHideElem(origin reflect.Value, path string) {
	var val reflect.Value
	if origin.Kind() == reflect.Ptr {
		val = origin.Elem()
	} else {
		val = origin
	}
	if val.Kind() == reflect.Struct {
		valtype := val.Type()
		for i := 0; i < valtype.NumField(); i++ {
			t := valtype.Field(i)
			v := val.Field(i)
			tags := strings.Split(string(t.Tag.Get("goblet")), ",")
			typeName := strings.ToLower(t.Name)
			if (typeName == "pwd" || typeName == "password") && v.CanSet() {
				v.SetString("hidden")
				continue
			}
			for _, tag := range tags {
				found := hideMatcher.FindStringSubmatch(tag)
				if len(found) > 0 {
					if found[2] != "" {
						for _, f := range strings.Split(found[2], ",") {
							if f == path {
								if v.CanSet() {
									v.SetString("hidden")
								}
							}
						}
					} else {
						if v.CanSet() {
							v.SetString("hidden")
						}
					}

				}
			}
		}
	} else if val.Kind() == reflect.Slice {
		for index := 0; index < val.Len(); index++ {
			autoHideElem(val.Index(index), path)
		}
	}
}
