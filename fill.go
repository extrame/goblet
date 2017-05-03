package goblet

import (
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"mime/multipart"
	"net"
	"reflect"
	"regexp"
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
	// read body
	if cx.fill_bts == nil {
		cx.fill_bts, err = ioutil.ReadAll(cx.request.Body)
	}
	if err != nil {
		return err
	}
	return json.Unmarshal(cx.fill_bts, v)
}

// an XML decoder for request body
type XmlRequestDecoder struct{}

func (d *XmlRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) (err error) {
	// read body
	if cx.fill_bts == nil {
		cx.fill_bts, err = ioutil.ReadAll(cx.request.Body)
	}
	if err != nil {
		return err
	}
	return xml.Unmarshal(cx.fill_bts, v)
}

// a form-enc decoder for request body
type FormRequestDecoder struct{}

type FileGetter func(string) (multipart.File, *multipart.FileHeader, error)

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.request.Form == nil {
		cx.request.ParseForm()
	}

	return UnmarshalForm(func(tag string) []string {
		values := (*map[string][]string)(&cx.request.Form)
		if values != nil {
			if results, ok := (*values)[tag]; !ok {
				//get the value of [Tag] from [tag](lower case), it maybe a problem TODO
				return (*values)[strings.ToLower(tag)]
			} else {
				return results
			}
		}
		return []string{}
	}, nil, v, autofill)
}

// a form-enc decoder for request body
type MultiFormRequestDecoder struct{}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	cx.request.ParseMultipartForm(32 << 20)
	values := (map[string][]string)(cx.request.Form)
	for k, v := range cx.request.MultipartForm.Value {
		values[k] = v
	}
	return UnmarshalForm(func(tag string) []string {
		return values[tag]
	}, func(tag string) (multipart.File, *multipart.FileHeader, error) {
		return cx.request.FormFile(tag)
	}, v, autofill)
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
	ct := cx.request.Header.Get("Content-Type")
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
func UnmarshalForm(value_getter func(string) []string,
	file_getter FileGetter,
	v interface{}, autofill bool) error {
	// check v is valid
	rv := reflect.ValueOf(v).Elem()
	// dereference pointer
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}

	if rv.Kind() == reflect.Struct {
		// for each struct field on v
		unmarshalStructInForm("", value_getter, file_getter, rv, 0, autofill, false, make(map[string]bool))
	} else {
		return fmt.Errorf("v must point to a struct type")
	}
	return nil
}

func unmarshalStructInForm(context string,
	values_getter func(string) []string,
	file_getter FileGetter,
	rvalue reflect.Value,
	offset int,
	autofill bool,
	inarray bool,
	parents map[string]bool) (err error) {

	if rvalue.Type().Kind() == reflect.Ptr {

		rvalue = rvalue.Elem()
	}
	rtype := rvalue.Type()

	parents[rtype.PkgPath()+"/"+rtype.Name()] = true

	success := false

	for i := 0; i < rtype.NumField() && err == nil; i++ {
		id, form_values, tag := getFormField(context, values_getter, rtype.Field(i), offset, inarray)
		increaseOffset := !(context != "" && inarray)
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
				if err = fill_struct(typ, values_getter, file_getter, val.Elem(), id, form_values, tag, used_offset, autofill, parents); err != nil {
					return err
				} else {
					break
				}
			case reflect.Struct:
				if err = fill_struct(rtype.Field(i).Type, values_getter, file_getter, rvalue.Field(i), id, form_values, tag, used_offset, autofill, parents); err != nil {
					fmt.Println(err)
					return err
				} else {
					break
				}
			case reflect.Interface:
				//ask the parent to tell me how to unmarshal it
				values := rvalue.MethodByName("UnmarshallForm").Call([]reflect.Value{reflect.ValueOf(rtype.Field(i).Name)})
				if len(values) == 2 && values[1].Interface() == nil {
					res := values[0].Interface()
					resValue := reflect.ValueOf(res)
					resType := reflect.TypeOf(res)
					if err = fill_struct(resType, values_getter, file_getter, resValue, id, form_values, tag, used_offset, autofill, parents); err != nil {
						rvalue.Field(i).Set(resValue)
						return err
					} else {
						break
					}
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
					if _, ok := parents[subRType.PkgPath()+"/"+subRType.Name()]; !ok {
						rvalueTemp := reflect.MakeSlice(rtype.Field(i).Type, 0, 0)
						subRValue := reflect.New(subRType)
						offset := 0
						for {
							err = unmarshalStructInForm(id, values_getter, file_getter, subRValue, offset, autofill, true, parents)
							if err != nil {
								break
							}
							offset++
							rvalueTemp = reflect.Append(rvalueTemp, subRValue.Elem())
						}
						rvalue.Field(i).Set(rvalueTemp)
					} else {
						err = fmt.Errorf("Too deep of type reuse %v", parents)
					}
				default:
					len_fv := len(form_values)
					rvnew := reflect.MakeSlice(rtype.Field(i).Type, len_fv, len_fv)
					for j := 0; j < len_fv; j++ {
						unmarshalField(context, rvnew.Index(j), form_values[j], autofill, tag)
					}
					rvalue.Field(i).Set(rvnew)
				}
			default:
				if len(form_values) > 0 && used_offset < len(form_values) {
					unmarshalField(context, rvalue.Field(i), form_values[used_offset], autofill, tag)
					success = true
				} else if len(tag) > 0 {
					unmarshalField(context, rvalue.Field(i), tag[0], autofill, tag)
				}
			}
		} else {
			log.Printf("cannot set value of (%s,%s) in fill", rtype.Field(i).Name, rtype.Field(i).Type.Name())
		}
	}
	if !success && err == nil {
		err = errors.New("no more element")
	}
	return
}

func getTag(prefix string, t reflect.StructField, offset int, inarray bool) (string, []string) {
	tags := []string{""}
	tag := t.Tag.Get("goblet")
	if tag != "" {
		tags = strings.Split(tag, ",")
		tag = tags[0]
	}
	if tag == "" {
		tag = t.Name
	}

	// values := []string{}

	// if form != nil {
	// 	values = (*form)[tag]
	// }

	if prefix != "" {
		if inarray {
			tag = fmt.Sprintf(prefix+"[%d]"+"["+tag+"]", offset)
		} else {
			tag = prefix + "[" + tag + "]"
		}
	}
	return tag, tags
}

func getFormField(prefix string, values_getter func(string) []string, t reflect.StructField, offset int, inarray bool) (string, []string, []string) {

	tag, tags := getTag(prefix, t, offset, inarray)

	values := values_getter(tag)

	return tag, values, tags[1:]
}

func fill_struct(typ reflect.Type,
	values_getter func(string) []string,
	file_getter FileGetter,
	val reflect.Value, id string, form_values []string, tag []string, used_offset int, autofill bool, parents map[string]bool) error {
	if typ.PkgPath() == "time" && typ.Name() == "Time" {
		var fillby string
		var fillby_valid = regexp.MustCompile(`^\s*fillby\((.*)\)\s*$`)
		for _, v := range tag {
			matched := fillby_valid.FindStringSubmatch(v)
			if len(matched) == 2 {
				fillby = matched[1]
			}
		}
		fillby = strings.TrimSpace(fillby)
		var value string
		if len(form_values) > used_offset {
			value = form_values[used_offset]
		}

		switch fillby {
		case "now":
			val.Set(reflect.ValueOf(time.Now()))
		case "timestamp":
			if unix, err := strconv.ParseInt(value, 10, 64); err == nil {
				val.Set(reflect.ValueOf(time.Unix(unix, 0)))
			} else {
				return err
			}
		default:
			if fillby == "" {
				fillby = time.RFC3339
			}
			if value != "" {
				time, err := time.Parse(fillby, value)
				if err == nil {
					val.Set(reflect.ValueOf(time))
				} else {
					log.Println(err)
					return err
				}
			}
		}
	} else if typ.PkgPath() == "github.com/extrame/goblet" && typ.Name() == "File" {
		var file File
		if file_getter != nil {
			if f, h, err := file_getter(id); err == nil {
				file.Name = h.Filename
				file.rc = f
				val.Set(reflect.ValueOf(file))
			} else {
				log.Println(err)
				return err
			}
		} else {
			return errors.New("file fill need MultipartForm request")
		}

	} else {
		unmarshalStructInForm(id, values_getter, file_getter, val, 0, autofill, false, parents)
	}
	return nil
}

func unmarshalField(contex string, v reflect.Value, form_value string, autofill bool, tags []string) error {
	// string -> type conversion
	switch v.Kind() {
	case reflect.Int64:
		if i, err := strconv.ParseInt(form_value, 10, 64); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint64:
		if i, err := strconv.ParseUint(form_value, 10, 64); err == nil {
			v.SetUint(i)
		}
	case reflect.Int, reflect.Int32:
		if i, err := strconv.ParseInt(form_value, 10, 32); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint32:
		if i, err := strconv.ParseUint(form_value, 10, 32); err == nil {
			v.SetUint(i)
		}
	case reflect.Int16:
		if i, err := strconv.ParseInt(form_value, 10, 16); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint16:
		if i, err := strconv.ParseUint(form_value, 10, 16); err == nil {
			v.SetUint(i)
		}
	case reflect.Int8:
		if i, err := strconv.ParseInt(form_value, 10, 8); err == nil {
			v.SetInt(i)
		}
	case reflect.Uint8:
		if i, err := strconv.ParseUint(form_value, 10, 8); err == nil {
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
