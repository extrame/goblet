package goblet

import (
	"net/http"
	"time"
)

type LoginInfoStorer interface {
	AddLoginAs(ctx *Context, name string, id string, timeduration ...time.Duration)
	GetLoginIdAs(ctx *Context, key string) (string, error)
}

type CookieLoginInfoStorer struct{}

func (c *CookieLoginInfoStorer) AddLoginAs(ctx *Context, name string, id string, timeduration ...time.Duration) {
	expire := time.Now().AddDate(0, 0, 1)
	if timeduration != nil {
		expire = time.Now().Add(timeduration[0])
	}
	cookie := new(http.Cookie)
	cookie.Name = name + "Id"
	cookie.Value = id
	cookie.Expires = expire
	cookie.Path = "/"
	cookie.RawExpires = expire.Format(time.UnixDate)
	ctx.AddSignedCookie(cookie)
}

func (c *CookieLoginInfoStorer) GetLoginIdAs(ctx *Context, key string) (string, error) {
	cookie, err := ctx.SignedCookie(key + "Id")
	if err == nil {
		return cookie.Value, nil
	}
	return "", err
}
