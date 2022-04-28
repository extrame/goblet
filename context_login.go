package goblet

import (
	"fmt"
	"strconv"
	"time"
)

func (c *Context) GetLoginId() (string, bool) {
	return c.GetLoginIdAs(USERCOOKIENAME)
}

func (c *Context) GetLoginIdAs(name string) (string, bool) {
	cookie, err := c.Server.loginSaver.GetLoginIdAs(c, name+"Id")
	if cookie != "" && err == nil {
		return cookie, true
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
		c.Server.loginSaver.AddLoginAs(c, name, userid)
	} else {
		c.Server.loginSaver.AddLoginAs(c, name, userid, timeduration[0])
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
	default:
		userid = fmt.Sprintf("%s", id)
	}
	if timeduration == nil {
		c.Server.loginSaver.AddLoginAs(c, "user", userid)
	} else {
		c.Server.loginSaver.AddLoginAs(c, "user", userid, timeduration[0])
	}

}

//Delete the login cookie saved
func (c *Context) DelLogin() {
	c.DelLoginAs("user")
}

//Delete the login cookie as specified name
func (c *Context) DelLoginAs(name string) {
	cookie, err := c.SignedCookie(name + "Id")
	if cookie != nil && err == nil {
		cookie.MaxAge = -1
		expire := time.Now()
		cookie.Expires = expire
		cookie.RawExpires = expire.Format(time.UnixDate)
		c.AddSignedCookie(cookie)
	}
}
