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
	cookie, err := c.Server.loginSaver.GetLoginIdAs(c, name)
	if cookie != nil && err == nil {
		return cookie.Name, true
	}
	return "", false
}

func (c *Context) GetLoginInfo() (*LoginContext, bool) {
	return c.GetLoginInfoAs(USERCOOKIENAME)
}

func (c *Context) GetLoginInfoAs(name string) (*LoginContext, bool) {
	cookie, err := c.Server.loginSaver.GetLoginIdAs(c, name)
	if cookie != nil && err == nil {
		return cookie, true
	}
	return nil, false
}

func (c *Context) AddLoginIdAs(id interface{}, name string, setter ...LoginInfoSetter) {
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

	lctx := &LoginContext{
		Name: name,
		Id:   userid,
	}

	for _, s := range setter {
		s(lctx)
	}

	if lctx.Deadline == nil {
		deadline := time.Now().AddDate(0, 0, 1)
		lctx.Deadline = &deadline
	}

	c.Server.loginSaver.AddLoginAs(c, lctx)

}

func (c *Context) AddLoginId(id interface{}, setter ...LoginInfoSetter) {
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
	c.AddLoginIdAs(userid, "user", setter...)

}

// Delete the login cookie saved
func (c *Context) DelLogin() error {
	return c.DelLoginAs("user")
}

// Delete the login cookie as specified name
func (c *Context) DelLoginAs(name string) error {
	return c.Server.loginSaver.DeleteLoginAs(c, name)
}
