package goblet

import (
	"net/http"
	"reflect"
	"time"

	"github.com/sirupsen/logrus"
)

type LoginContext struct {
	Name     string
	Id       string
	Deadline *time.Time
	Attrs    map[string]interface{}
}

func (l *LoginContext) HasAttr(key string, content interface{}) bool {
	if l.Attrs == nil {
		return false
	}
	if v, ok := l.Attrs[key]; ok {
		//use reflect to test v is slice
		if compareStringAndNumber(v, content) {
			return true
		}
		rv := reflect.ValueOf(v)
		if rv.Kind() == reflect.Slice {
			for i := 0; i < rv.Len(); i++ {
				if compareStringAndNumber(rv.Index(i).Interface(), content) {
					return true
				}
			}
		} else if rv.Kind() == reflect.Map {
			if rv.MapIndex(reflect.ValueOf(content)).IsValid() {
				return true
			}
		}
	}
	return false
}

func compareStringAndNumber(v interface{}, content interface{}) bool {
	if v == content {
		return true
	}
	var rc = reflect.ValueOf(content)
	return rc.Convert(reflect.TypeOf(v)).Interface() == v
}

type LoginInfoSetter func(*LoginContext)

func WithDuration(t time.Duration) LoginInfoSetter {
	return func(ctx *LoginContext) {
		t := time.Now().Add(t)
		ctx.Deadline = &t
	}
}

func WithAttribute(key string, value interface{}) LoginInfoSetter {
	return func(ctx *LoginContext) {
		if ctx.Attrs == nil {
			ctx.Attrs = make(map[string]interface{})
		}
		ctx.Attrs[key] = value
	}
}

type LoginInfoStorer interface {
	AddLoginAs(ctx *Context, lctx *LoginContext) string
	GetLoginIdAs(ctx *Context, key string) (*LoginContext, error)
	DeleteLoginAs(ctx *Context, key string) error
}

type CookieLoginInfoStorer struct{}

func (c *CookieLoginInfoStorer) AddLoginAs(ctx *Context, lctx *LoginContext) string {

	cookie := new(http.Cookie)
	cookie.Name = lctx.Name + "Id"
	cookie.Value = lctx.Id
	cookie.Expires = *lctx.Deadline
	cookie.Path = "/"
	cookie.RawExpires = lctx.Deadline.Format(time.UnixDate)
	if lctx.Attrs != nil {
		logrus.Error("Cookie Login Info Storer not support attrs, please consider use session or jwt")
	}
	ctx.AddSignedCookie(cookie)
	return cookie.Value
}

func (c *CookieLoginInfoStorer) GetLoginIdAs(ctx *Context, key string) (*LoginContext, error) {
	cookie, err := ctx.SignedCookie(key + "Id")
	if err == nil {
		return &LoginContext{
			Name: key,
			Id:   cookie.Value,
		}, nil
	}
	return nil, err
}

func (c *CookieLoginInfoStorer) DeleteLoginAs(ctx *Context, key string) error {
	cookie, err := ctx.SignedCookie(key + "Id")
	if cookie != nil && err == nil {
		cookie.MaxAge = -1
		expire := time.Now()
		cookie.Expires = expire
		cookie.Path = "/"
		cookie.RawExpires = expire.Format(time.UnixDate)
		_, err = ctx.AddSignedCookie(cookie)
	}
	return err
}
