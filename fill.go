package goblet

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"mime/multipart"
	"reflect"
	"strings"

	"github.com/extrame/unmarshall"
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
	return xml.Unmarshal(cx.fill_bts, v)
}

// a form-enc decoder for request body
type FormRequestDecoder struct{}

type FileGetter func(string) (multipart.File, *multipart.FileHeader, error)

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.request.Form == nil {
		cx.request.ParseForm()
	}

	var unmarshaller = unmarshall.Unmarshaller{
		Values: func() map[string][]string {
			return cx.request.Form
		},
		ValueGetter: func(tag string) []string {
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
		},
		TagConcatter: concatPrefix,
		AutoFill:     autofill,
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
	cx.request.ParseMultipartForm(32 << 20)
	values := (map[string][]string)(cx.request.Form)
	for k, v := range cx.request.MultipartForm.Value {
		values[k] = v
	}

	var unmarshaller = unmarshall.Unmarshaller{
		Values: func() map[string][]string {
			return values
		},
		ValueGetter: func(tag string) []string {
			return values[tag]
		},
		TagConcatter: concatPrefix,
		FillForSpecifiledType: map[string]func(id string) (reflect.Value, error){
			"github.com/extrame/goblet.File": func(id string) (reflect.Value, error) {
				var file File
				var err error
				var f multipart.File
				var h *multipart.FileHeader
				if f, h, err = cx.request.FormFile(id); err == nil {
					file.Name = h.Filename
					file.rc = f
				}
				return reflect.ValueOf(file), err
			},
		},
		AutoFill: autofill,
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

func concatPrefix(prefix, tag string) string {
	return prefix + "[" + tag + "]"
}
