package goblet

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"net/url"
	"reflect"
	"strings"

	"log/slog"

	"github.com/creasty/defaults"
	"github.com/gorilla/schema"
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
		cx.fill_bts, err = ioutil.ReadAll(cx.Request.Body)
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
		cx.fill_bts, err = ioutil.ReadAll(cx.Request.Body)
	}
	if err != nil {
		return err
	}
	if err = xml.Unmarshal(cx.fill_bts, v); err != nil {
		slog.Error("[Fill Error]", "request", string(cx.fill_bts), "error", err)
	}
	return err
}

// a form-enc decoder for request body
type FormRequestDecoder struct {
	decoder *schema.Decoder
}

func NewFormRequestDecoder() *FormRequestDecoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("query") // 使用form标签
	decoder.IgnoreUnknownKeys(true)
	return &FormRequestDecoder{decoder: decoder}
}

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if cx.Request.Form == nil {
		if err := cx.Request.ParseForm(); err != nil {
			return err
		}
	}

	// 处理特殊类型
	if len(cx.Server.filler) > 0 {
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Ptr {
			rv = rv.Elem()
		}
		if rv.Kind() == reflect.Struct {
			for i := 0; i < rv.NumField(); i++ {
				field := rv.Type().Field(i)
				if typ, ok := field.Tag.Lookup("query"); ok {
					if fn, exists := cx.Server.filler[typ]; exists {
						values := cx.Request.Form[field.Name]
						if len(values) > 0 {
							obj, err := fn(values[0])
							if err != nil {
								return err
							}
							rv.Field(i).Set(reflect.ValueOf(obj))
						}
					}
				}
			}
		}
	}

	return d.decoder.Decode(v, cx.Request.Form)
}

// a form-enc decoder for request body
type MultiFormRequestDecoder struct {
	decoder *schema.Decoder
	tag     string
}

func NewMultiFormRequestDecoder() *MultiFormRequestDecoder {
	decoder := schema.NewDecoder()
	decoder.SetAliasTag("form")
	decoder.IgnoreUnknownKeys(true)
	return &MultiFormRequestDecoder{decoder: decoder, tag: "form"}
}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	if err := cx.Request.ParseMultipartForm(32 << 20); err != nil {
		return err
	}

	// 合并表单值
	values := make(url.Values)
	for k, v := range cx.Request.MultipartForm.Value {
		values[k] = v
	}
	for k, v := range cx.Request.Form {
		values[k] = v
	}

	// 处理文件上传
	rv := reflect.ValueOf(v)
	if rv.Kind() == reflect.Ptr {
		rv = rv.Elem()
	}
	if rv.Kind() == reflect.Struct {
		for i := 0; i < rv.NumField(); i++ {
			field := rv.Type().Field(i)
			if field.Type.PkgPath() == "github.com/extrame/goblet/v2" && field.Type.Name() == "File" {
				slog.Info("parse file arguments", "field", field)
				tag, ok := field.Tag.Lookup(d.tag)
				if !ok {
					tag = field.Name
				}
				if file, _, err := cx.Request.FormFile(tag); err == nil {
					fileObj := File{
						rc:     file,
						Name:   field.Name,
						Size:   0, // 需要从header获取
						Header: make(map[string][]string),
					}
					rv.Field(i).Set(reflect.ValueOf(fileObj))
				}
			}
		}
	}

	// 处理特殊类型
	if len(cx.Server.multiFiller) > 0 {
		if rv.Kind() == reflect.Struct {
			for i := 0; i < rv.NumField(); i++ {
				field := rv.Type().Field(i)
				if typ, ok := field.Tag.Lookup("form"); ok {
					if fn, exists := cx.Server.multiFiller[typ]; exists {
						obj, err := fn(cx, field.Name)
						if err != nil {
							return err
						}
						rv.Field(i).Set(reflect.ValueOf(obj))
					}
				}
			}
		}
	}

	return d.decoder.Decode(v, values)
}

// map of Content-Type -> RequestDecoders
var decoders map[string]RequestDecoder = map[string]RequestDecoder{
	"application/json":                  new(JsonRequestDecoder),
	"application/xml":                   new(XmlRequestDecoder),
	"text/xml":                          new(XmlRequestDecoder),
	"application/x-www-form-urlencoded": NewFormRequestDecoder(),
	"text/plain":                        NewFormRequestDecoder(),
	"multipart/form-data":               NewMultiFormRequestDecoder(),
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
	//if method is GET, only form in url is supported
	if cx.Request.Method == "GET" {
		ct = "application/x-www-form-urlencoded"
	}
	// default to urlencoded
	if ct == "" {
		if cx.Server.Config.Basic.DefaultType != "" {
			ct = cx.Server.Config.Basic.DefaultType
		} else {
			ct = "application/x-www-form-urlencoded"
		}
	} else if strings.HasPrefix(ct, "text/plain") && cx.Server.Config.Basic.DefaultType != "" {
		ct = cx.Server.Config.Basic.DefaultType
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

	// 设置默认值
	if autofill {
		defaults.Set(v)
	}

	// 无论任何 Content Type，都应该首先从 query string 中 fill 数据
	if cx.Request.URL.RawQuery != "" {
		queryDecoder := schema.NewDecoder()
		queryDecoder.SetAliasTag("query")
		queryDecoder.IgnoreUnknownKeys(true)
		if err := queryDecoder.Decode(v, cx.Request.URL.Query()); err != nil {
			slog.Error("Failed to decode query string", "error", err)
			return err
		}
	}

	// decode
	err := decoder.Unmarshal(cx, v, autofill)
	if err != nil {
		slog.Error("Failed to decode request data", "error", err)
		return err
	}
	// all clear
	return nil
}
