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
	c.AddSignedCookie(cookie)
}

//Delete the login cookie saved
func (c *Context) DelLogin(names ...string) {
	if names != nil {
		for _, v := range names {
			c.delLoginAs(v)
		}
	} else {
		c.delLoginAs("user")
	}
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
