package goblet

import (
	"net/http"
	"strconv"
	"time"
)

func (c *Context) GetLoginId() (string, bool) {
	return c.GetLoginIdAs(USERCOOKIENAME)
}

func (c *Context) GetLoginIdAs(name string) (string, bool) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		return cookie.Value, true
	}
	return "", false
}

func (c *Context) AddLoginId(id interface{}) {
	switch rid := id.(type) {
	case string:
		c.addLoginAs("user", rid)
	case int64:
		c.addLoginAs("user", strconv.FormatInt(rid, 10))
	}
}

func (c *Context) addLoginAs(name string, id string) {
	expire := time.Now().AddDate(0, 0, 1)
	cookie := new(http.Cookie)
	cookie.Name = name + "Id"
	cookie.Value = id
	cookie.Expires = expire
	cookie.RawExpires = expire.Format(time.UnixDate)
	c.AddSignedCookie(cookie)
}

//Delete the login cookie saved
func (c *Context) DelLogin() {
	c.delLoginAs("user")
}

func (c *Context) delLoginAs(name string) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		cookie.MaxAge = -1
		expire := time.Now()
		cookie.Expires = expire
		cookie.RawExpires = expire.Format(time.UnixDate)
		c.AddSignedCookie(cookie)
	}
}
