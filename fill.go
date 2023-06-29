package goblet

import (
	"encoding/json"
	"encoding/xml"
	"errors"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"net/url"
	"reflect"
	"strings"

	"github.com/extrame/unmarshall"
	"github.com/sirupsen/logrus"
)

type FormFillFn func(content string) (interface{}, error)
type MultiFormFillFn func(ctx *Context, id string) (interface{}, error)

func (s *Server) AddFillForTypeInForm(typ string, fn FormFillFn) {
	s.filler[typ] = fn
}

func (s *Server) AddFillForTypeInMultiForm(typ string, fn MultiFormFillFn) {
	s.multiFiller[typ] = fn
}

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
	if err = xml.Unmarshal(cx.fill_bts, v); err != nil {
		logrus.Errorf("[Fill Error]Request:%s,Err:%s\n", string(cx.fill_bts), err.Error())
	}
	return err
}

// a form-enc decoder for request body
type FormRequestDecoder struct{}

type FileGetter func(string) (multipart.File, *multipart.FileHeader, error)

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.request.Form == nil {
		cx.request.ParseForm()
	}

	var maxlength = 0
	for k, _ := range cx.request.Form {
		if len(k) > maxlength {
			maxlength = len(k)
		}
	}

	var unmarshaller = unmarshall.Unmarshaller{
		Values: func() map[string][]string {
			return cx.request.Form
		},
		ValuesGetter: func(prefix string) url.Values {
			values := (*map[string][]string)(&cx.request.Form)
			var sub = make(url.Values)
			if values != nil {
				for k, v := range *values {
					if strings.HasPrefix(k, prefix+"[") {
						sub[k] = v
					}
				}
			}
			return sub
		},
		ValueGetter: func(tag string) []string {
			values := (*map[string][]string)(&cx.request.Form)
			if values != nil {
				var lower = strings.ToLower(tag)
				if results, ok := (*values)[tag]; ok {
					return results
				}
				if results, ok := (*values)[lower]; ok {
					return results
				}
				if results, ok := (*values)[tag+"[]"]; ok {
					return results
				}
				if results, ok := (*values)[lower+"[]"]; ok {
					return results
				}
			}
			return []string{}
		},
		Tag:          "goblet",
		MaxLength:    maxlength,
		TagConcatter: concatPrefix,
		BaseName: func(path string, prefix string) string {
			return strings.Split(strings.TrimPrefix(path, prefix+"["), "]")[0]
		},
		AutoFill: autofill,
	}

	for typ, fn := range cx.Server.filler {
		if unmarshaller.FillForSpecifiledType == nil {
			unmarshaller.FillForSpecifiledType = make(map[string]func(id string) (reflect.Value, error))
		}
		unmarshaller.FillForSpecifiledType[typ] = func(content string) (reflect.Value, error) {
			obj, err := fn(content)
			return reflect.ValueOf(obj), err
		}
	}

	return unmarshaller.Unmarshall(v)
}

// a form-enc decoder for request body
type MultiFormRequestDecoder struct{}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	err := cx.request.ParseMultipartForm(32 << 20)
	if err != nil {
		return err
	}
	values := (map[string][]string)(cx.request.Form)
	if cx.request.MultipartForm == nil {
		return errors.New("MultipartForm is empty")
	}
	var maxlength = 0

	for k, v := range cx.request.MultipartForm.Value {
		values[k] = v
		if len(k) > maxlength {
			maxlength = len(k)
		}
	}

	for k, _ := range cx.request.MultipartForm.File {
		if len(k) > maxlength {
			maxlength = len(k)
		}
	}

	var unmarshaller = unmarshall.Unmarshaller{
		Values: func() map[string][]string {
			return values
		},
		MaxLength: maxlength,
		ValueGetter: func(tag string) []string {
			return values[tag]
		},
		ValuesGetter: func(prefix string) url.Values {
			var sub = make(url.Values)
			if values != nil {
				for k, v := range values {
					if strings.HasPrefix(k, prefix+"[") {
						sub[k] = v
					}
				}
			}
			return sub
		},
		TagConcatter: concatPrefix,
		BaseName: func(path string, prefix string) string {
			return strings.Split(strings.TrimPrefix(path, prefix+"["), "]")[0]
		},
		FillForSpecifiledType: map[string]func(id string) (reflect.Value, error){
			"github.com/extrame/goblet.File": func(id string) (reflect.Value, error) {
				var file File
				var err error
				var f multipart.File
				var h *multipart.FileHeader
				if f, h, err = cx.request.FormFile(id); err == nil {
					file.Name = h.Filename
					file.Header = h.Header
					file.rc = f
					return reflect.ValueOf(file), err
				} else {
					return reflect.ValueOf(nil), err
				}
			},
		},
		AutoFill: autofill,
		Tag:      "goblet",
	}

	for typ, fn := range cx.Server.multiFiller {
		unmarshaller.FillForSpecifiledType[typ] = func(id string) (reflect.Value, error) {
			obj, err := fn(cx, id)
			return reflect.ValueOf(obj), err
		}
	}

	return unmarshaller.Unmarshall(v)
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
	//if method is GET, only form in url is supported
	if cx.request.Method == "GET" {
		ct = "application/x-www-form-urlencoded"
	}
	// default to urlencoded
	if ct == "" {
		if cx.Server.Basic.DefaultType != "" {
			ct = cx.Server.Basic.DefaultType
		} else {
			ct = "application/x-www-form-urlencoded"
		}
	} else if strings.HasPrefix(ct, "text/plain") && cx.Server.Basic.DefaultType != "" {
		ct = cx.Server.Basic.DefaultType
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

func concatPrefix(prefix, tag string) string {
	return prefix + "[" + tag + "]"
}
