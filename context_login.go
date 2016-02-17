package goblet

import (
	"github.com/valyala/fasthttp"
	"strconv"
	"time"
)

func (c *Context) GetLoginId() (string, bool) {
	return c.GetLoginIdAs(USERCOOKIENAME)
}

func (c *Context) GetLoginIdAs(name string) (string, bool) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		return string(cookie.Value()), true
	}
	return "", false
}

func (c *Context) AddLoginIdAs(id interface{}, name string, timeduration ...time.Duration) {
	if name == "" {
		name = "user"
	}
	var userid string
	switch rid := id.(type) {
	case string:
		userid = rid
	case int:
		userid = strconv.FormatInt(int64(rid), 10)
	case int32:
		userid = strconv.FormatInt(int64(rid), 10)
	case int64:
		userid = strconv.FormatInt(rid, 10)
	}
	if timeduration == nil {
		c.addLoginAs(name, userid)
	} else {
		c.addLoginAs(name, userid, timeduration[0])
	}

}

func (c *Context) AddLoginId(id interface{}, timeduration ...time.Duration) {
	var userid string
	switch rid := id.(type) {
	case string:
		userid = rid
	case int:
		userid = strconv.FormatInt(int64(rid), 10)
	case int32:
		userid = strconv.FormatInt(int64(rid), 10)
	case int64:
		userid = strconv.FormatInt(rid, 10)
	}
	if timeduration == nil {
		c.addLoginAs("user", userid)
	} else {
		c.addLoginAs("user", userid, timeduration[0])
	}

}

func (c *Context) addLoginAs(name string, id string, timeduration ...time.Duration) {
	cookie := new(fasthttp.Cookie)
	cookie.SetKey(name + "Id")
	cookie.SetValue(id)
	if timeduration != nil {
		expire := time.Now().Add(timeduration[0])
		cookie.SetExpire(expire)
	}
	cookie.SetPath("/")
	// cookie.RawExpires = expire.Format(time.UnixDate)
	c.AddSignedCookie(cookie)
}

//Delete the login cookie saved
func (c *Context) DelLogin() {
	c.DelLoginAs("user")
}

//Delete the login cookie as specified name
func (c *Context) DelLoginAs(name string) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		// cookie.MaxAge = -1
		cookie.SetKey(name + "Id")
		cookie.SetExpire(fasthttp.CookieExpireDelete)
		// cookie.RawExpires = expire.Format(time.UnixDate)
		c.AddSignedCookie(cookie)
	}
}
