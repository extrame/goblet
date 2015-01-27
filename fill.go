package goblet

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"reflect"
	"strconv"
	"strings"
	"time"
)

// types that impliment RequestDecoder can unmarshal
// the request body into an apropriate type/struct
type RequestDecoder interface {
	Unmarshal(cx *Context, v interface{}, autofill bool) error
}

// a JSON decoder for request body (just a wrapper to json.Unmarshal)
type JsonRequestDecoder struct{}

func (d *JsonRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) (err error) {
	var request []byte
	// read body
	request, err = ioutil.ReadAll(cx.Request.Body)
	if err != nil {
		return err
	}
	return json.Unmarshal(request, v)
}

// an XML decoder for request body
type XmlRequestDecoder struct{}

func (d *XmlRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	// read body
	data, err := ioutil.ReadAll(cx.Request.Body)
	if err != nil {
		return err
	}
	return xml.Unmarshal(data, v)
}

// a form-enc decoder for request body
type FormRequestDecoder struct{}

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.Request.Form == nil {
		cx.Request.ParseForm()
	}
	var values *map[string][]string
	values = (*map[string][]string)(&cx.Request.Form)
	return UnmarshalForm(values, v, autofill)
}

// a form-enc decoder for request body
type MultiFormRequestDecoder struct{}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	cx.Request.ParseMultipartForm(32 << 20)
	values := (map[string][]string)(cx.Request.Form)
	for k, v := range cx.Request.MultipartForm.Value {
		values[k] = v
	}
	return UnmarshalForm(&values, v, autofill)
}

// map of Content-Type -> RequestDecoders
var decoders map[string]RequestDecoder = map[string]RequestDecoder{
	"application/json":                  new(JsonRequestDecoder),
	"application/xml":                   new(XmlRequestDecoder),
	"text/xml":                          new(XmlRequestDecoder),
	"application/x-www-form-urlencoded": new(FormRequestDecoder),
	"text/plain":                        new(FormRequestDecoder),
	"multipart/form-data":               new(MultiFormRequestDecoder),
}

// goweb.Context Helper function to fill a variable with the contents
// of the request body. The body will be decoded based
// on the content-type and an apropriate RequestDecoder
// automatically selected
// If you want to use md5 function for the specified field, please add
// md5 tag for it. AND the md5 tag must be the last one of the tags, so
// if you have no other tag, please add ',' before md5
func (cx *Context) Fill(v interface{}, fills ...bool) error {
	// get content type
	ct := cx.Request.Header.Get("Content-Type")
	// default to urlencoded
	if ct == "" {
		ct = "application/x-www-form-urlencoded"
	}
	autofill := true
	if len(fills) > 0 {
		autofill = fills[0]
	}
	return cx.FillAs(v, autofill, ct)
}

func (cx *Context) FillAs(v interface{}, autofill bool, ct string) error {
	// ignore charset (after ';')
	ct = strings.Split(ct, ";")[0]
	// get request decoder
	decoder, ok := decoders[ct]
	if ok != true {
		return fmt.Errorf("Cannot decode request for %s data", ct)
	}
	// decode
	err := decoder.Unmarshal(cx, v, autofill)
	if err != nil {
		fmt.Println(err)
		return err
	}
	// all clear
	return nil
}

// Fill a struct `v` from the values in `goblet`
func UnmarshalForm(form *map[string][]string, v interface{}, autofill bool) error {
	// check v is valid
	rv := reflect.ValueOf(v).Elem()
	// dereference pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {

		// for each struct field on v
		unmarshalStructInForm("", form, rv, 0, autofill, false)
	} else if rv.Kind() == reflect.Map && !rv.IsNil() {
		// for each form value add it to the map
		for k, v := range *form {
			if len(v) > 0 {
				rv.SetMapIndex(reflect.ValueOf(k), reflect.ValueOf(v[0]))
			}
		}
	} else {
		return fmt.Errorf("v must point to a struct or a non-nil map type")
	}
	return nil
}

func unmarshalStructInForm(context string, form *map[string][]string, rvalue reflect.Value, offset int, autofill bool, inarray bool) (err error) {

	if rvalue.Type().Kind() == reflect.Ptr {

		rvalue = rvalue.Elem()
	}
	rtype := rvalue.Type()

	success := false

	for i := 0; i < rtype.NumField(); i++ {
		id, form_values, tag, increaseOffset := getFormField(context, form, rtype.Field(i), offset, inarray)
		var used_offset = 0
		if increaseOffset {
			used_offset = offset
		}
		if rvalue.Field(i).CanSet() {
			switch rtype.Field(i).Type.Kind() {
			case reflect.Ptr: //TODO if the ptr point to a basic data, it will crash
				val := rvalue.Field(i)
				typ := rtype.Field(i).Type.Elem()
				if val.IsNil() {
					val.Set(reflect.New(typ))
				}
				if err := fill_struct(typ, form, rvalue.Field(i), id, form_values, tag, used_offset, autofill); err != nil {
					fmt.Println(err)
					return err
				} else {
					break
				}
			case reflect.Struct:
				if err := fill_struct(rtype.Field(i).Type, form, rvalue.Field(i), id, form_values, tag, used_offset, autofill); err != nil {
					fmt.Println(err)
					return err
				} else {
					break
				}
			case reflect.Slice:
				fType := rtype.Field(i).Type
				subRType := rtype.Field(i).Type.Elem()
				if fType.PkgPath() == "net" && fType.Name() == "IP" && len(form_values) > 0 && used_offset < len(form_values) {
					rvalue.Field(i).Set(reflect.ValueOf(net.ParseIP(form_values[used_offset])))
					continue
				}
				switch subRType.Kind() {
				case reflect.Struct:
					rvalueTemp := reflect.MakeSlice(rtype.Field(i).Type, 0, 0)
					subRValue := reflect.New(subRType)
					offset := 0
					for {
						err = unmarshalStructInForm(id, form, subRValue, offset, autofill, true)
						if err != nil {
							fmt.Println(err)
							break
						}

						offset++
						rvalueTemp = reflect.Append(rvalueTemp, subRValue.Elem())
					}
					rvalue.Field(i).Set(rvalueTemp)
				default:
					len_fv := len(form_values)
					rvnew := reflect.MakeSlice(rtype.Field(i).Type, len_fv, len_fv)
					for j := 0; j < len_fv; j++ {
						unmarshalField(context, form, rvnew.Index(j), form_values[i], autofill, tag)
					}
					rvalue.Field(i).Set(rvnew)
				}
			default:
				if len(form_values) > 0 && used_offset < len(form_values) {
					unmarshalField(context, form, rvalue.Field(i), form_values[used_offset], autofill, tag)
					success = true
				}
			}
		}
	}
	if !success {
		return errors.New("no more element")
	}
	return nil
}

func getFormField(prefix string, form *map[string][]string, t reflect.StructField, offset int, inarray bool) (string, []string, []string, bool) {
	tags := strings.Split(t.Tag.Get("goblet"), ",")
	tag := tags[0]
	var values = (*form)[tag]
	var increaseOffset = true
	if len(tags) == 0 || tags[0] == "" {
		tag = t.Name
		values = (*form)[tag]
	}
	if prefix != "" {
		if inarray {
			increaseOffset = false
			tag = fmt.Sprintf(prefix+"[%d]"+"["+tag+"]", offset)
		} else {
			increaseOffset = true
			tag = prefix + "[" + tag + "]"
		}
		values = (*form)[tag]
	}
	return tag, values, tags[1:], increaseOffset
}

func fill_struct(typ reflect.Type, form *map[string][]string, val reflect.Value, id string, form_values []string, tag []string, used_offset int, autofill bool) error {
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		if len(tag) > 0 && tag[0] == "fillby(now)" && autofill {
			val.Set(reflect.ValueOf(time.Now()))
		} else if len(form_values) > 0 {
			time, err := time.Parse(time.RFC3339, form_values[used_offset])
			if err == nil {
				val.Set(reflect.ValueOf(time))
			} else {
				fmt.Println(err)
				return err
			}
		}
	} else {
		unmarshalStructInForm(id, form, val, 0, autofill, false)
	}
	return nil
}

func unmarshalField(context string, form *map[string][]string, v reflect.Value, form_value string, autofill bool, tags []string) error {
	// string -> type conversion
	switch v.Kind() {
	case reflect.Int64:
		fallthrough
	case reflect.Int:
		fallthrough
	case reflect.Int8:
		fallthrough
	case reflect.Int16:
		// convert to Int
		// convert to Int64
		if i, err := strconv.ParseInt(form_value, 10, 16); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint16:
		// convert to Int
		// convert to Int64
		if i, err := strconv.ParseUint(form_value, 10, 16); err == nil {
			v.SetUint(i)
		}
	case reflect.String:
		// copy string
		if len(tags) > 0 && tags[len(tags)-1] == "md5" {
			h := md5.New()
			h.Write([]byte(form_value))
			v.SetString(hex.EncodeToString(h.Sum(nil)))
		} else {
			v.SetString(form_value)
		}
	case reflect.Float64:
		if f, err := strconv.ParseFloat(form_value, 64); err == nil {
			v.SetFloat(f)
		}
	case reflect.Float32:
		if f, err := strconv.ParseFloat(form_value, 32); err == nil {
			v.SetFloat(f)
		}
	case reflect.Bool:
		// the following strings convert to true
		// 1,true,on,yes
		fv := form_value
		if fv == "1" || fv == "true" || fv == "on" || fv == "yes" {
			v.SetBool(true)
		}
	default:
		fmt.Println("unknown type", v.Kind())
	}
	return nil
}
