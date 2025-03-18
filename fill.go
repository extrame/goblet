package goblet

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io/ioutil"
	"log/slog"
	"mime/multipart"
	"reflect"
	"strings"

	"github.com/go-playground/form/v4"
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
		slog.Error("Failed to unmarshal XML request",
			"request", string(cx.fill_bts),
			"error", err)
	}
	return err
}

// FormRequestDecoder handles application/x-www-form-urlencoded data
type FormRequestDecoder struct {
	decoder *form.Decoder
}

func NewFormRequestDecoder() *FormRequestDecoder {
	decoder := form.NewDecoder()

	// 注册自定义类型转换器
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		if len(vals) > 0 {
			return vals[0], nil
		}
		return "", nil
	}, "")

	return &FormRequestDecoder{decoder: decoder}
}

type FileGetter func(string) (multipart.File, *multipart.FileHeader, error)

func (d *FormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	// 解析表单数据
	if err := cx.request.ParseForm(); err != nil {
		return fmt.Errorf("parsing form: %w", err)
	}

	// 使用 form 包解码基本字段
	if err := d.decoder.Decode(v, cx.request.Form); err != nil {
		return fmt.Errorf("decoding form: %w", err)
	}

	// 处理自定义类型
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	// 遍历结构体字段，处理自定义类型
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		typeStr := field.Type().String()
		if filler, ok := cx.Server.filler[typeStr]; ok {
			tag := typ.Field(i).Tag.Get("goblet")
			if tag == "" {
				continue
			}

			values := cx.request.Form[tag]
			if len(values) > 0 {
				if obj, err := filler(values[0]); err == nil {
					field.Set(reflect.ValueOf(obj))
				}
			}
		}
	}

	return nil
}

// MultiFormRequestDecoder handles multipart/form-data
type MultiFormRequestDecoder struct {
	decoder *form.Decoder
}

func NewMultiFormRequestDecoder() *MultiFormRequestDecoder {
	decoder := form.NewDecoder()
	decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
		if len(vals) > 0 {
			return vals[0], nil
		}
		return "", nil
	}, "")

	return &MultiFormRequestDecoder{decoder: decoder}
}

func (d *MultiFormRequestDecoder) Unmarshal(cx *Context, v interface{}, autofill bool) error {
	// 解析多部分表单数据，设置32MB的最大内存
	if err := cx.request.ParseMultipartForm(32 << 20); err != nil {
		return fmt.Errorf("parsing multipart form: %w", err)
	}

	if cx.request.MultipartForm == nil {
		return fmt.Errorf("multipart form is empty")
	}

	// 使用 form 包解码基本字段
	if err := d.decoder.Decode(v, cx.request.MultipartForm.Value); err != nil {
		return fmt.Errorf("decoding multipart form: %w", err)
	}

	// 处理文件和自定义类型
	val := reflect.ValueOf(v)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}

	if val.Kind() != reflect.Struct {
		return nil
	}

	// 遍历结构体字段
	typ := val.Type()
	for i := 0; i < val.NumField(); i++ {
		field := val.Field(i)
		if !field.CanSet() {
			continue
		}

		// 获取字段类型和标签
		typeStr := field.Type().String()
		structField := typ.Field(i)
		tag := structField.Tag.Get("form")
		fieldName := structField.Name
		if tag != "" {
			fieldName = tag
		}
		lowerFieldName := strings.ToLower(fieldName)

		// 处理 goblet.File 类型
		if typeStr == "github.com/extrame/goblet.File" {
			// 处理单个文件
			if field.Kind() != reflect.Slice {
				// 尝试使用原始字段名
				file, header, err := cx.request.FormFile(fieldName)
				if err != nil {
					// 如果找不到，尝试使用全小写字段名
					file, header, err = cx.request.FormFile(lowerFieldName)
					if err != nil {
						slog.Debug("No file found for field",
							"original_field", fieldName,
							"lowercase_field", lowerFieldName,
							"error", err)
						continue
					}
				}

				gobletFile := File{
					Name:   header.Filename,
					Header: header.Header,
					Size:   header.Size,
					rc:     file,
				}
				field.Set(reflect.ValueOf(gobletFile))
				continue
			}

			// 处理文件数组
			var files []*multipart.FileHeader
			// 先尝试原始字段名
			if f := cx.request.MultipartForm.File[fieldName]; len(f) > 0 {
				files = f
			} else {
				// 如果找不到，尝试全小写字段名
				if f := cx.request.MultipartForm.File[lowerFieldName]; len(f) > 0 {
					files = f
				}
			}

			if len(files) > 0 {
				// 创建文件切片
				sliceType := reflect.SliceOf(field.Type().Elem())
				fileSlice := reflect.MakeSlice(sliceType, len(files), len(files))

				// 处理每个文件
				for j, fileHeader := range files {
					file, err := fileHeader.Open()
					if err != nil {
						slog.Error("Failed to open file",
							"filename", fileHeader.Filename,
							"error", err)
						continue
					}

					gobletFile := File{
						Name:   fileHeader.Filename,
						Header: fileHeader.Header,
						Size:   fileHeader.Size,
						rc:     file,
					}
					fileSlice.Index(j).Set(reflect.ValueOf(gobletFile))
				}
				field.Set(fileSlice)
			}
			continue
		}

		// 处理自定义类型
		if filler, ok := cx.Server.multiFiller[typeStr]; ok {
			// 先尝试使用原始字段名
			if obj, err := filler(cx, fieldName); err == nil {
				field.Set(reflect.ValueOf(obj))
			} else {
				// 如果失败，尝试使用全小写字段名
				if obj, err := filler(cx, lowerFieldName); err == nil {
					field.Set(reflect.ValueOf(obj))
				} else {
					slog.Debug("Failed to fill custom type",
						"type", typeStr,
						"original_field", fieldName,
						"lowercase_field", lowerFieldName,
						"error", err)
				}
			}
		}
	}

	return nil
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
