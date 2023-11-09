package goblet

import (
	"net/http"
	"time"

	"github.com/sirupsen/logrus"
)

type LoginContext struct {
	Name     string
	Id       string
	Deadline *time.Time
	Attrs    map[string]interface{}
}

type LoginInfoSetter func(*LoginContext)

func WithDuration(t time.Duration) LoginInfoSetter {
	return func(ctx *LoginContext) {
		t := time.Now().Add(t)
		ctx.Deadline = &t
	}
}

type LoginInfoStorer interface {
	AddLoginAs(ctx *Context, lctx *LoginContext)
	GetLoginIdAs(ctx *Context, key string) (*LoginContext, error)
	DeleteLoginAs(ctx *Context, key string) error
}

type CookieLoginInfoStorer struct{}

func (c *CookieLoginInfoStorer) AddLoginAs(ctx *Context, lctx *LoginContext) {

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
