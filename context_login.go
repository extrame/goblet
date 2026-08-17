package goblet

import (
	"fmt"
	"time"
)

func (c *Context) GetLoginId() (string, bool) {
	return c.GetLoginIdAs(USERCOOKIENAME)
}

func (c *Context) GetLoginIdAs(name string) (string, bool) {
	cookie, err := c.Server.loginSaver.GetLoginIdAs(c, name)
	if cookie != nil && err == nil {
		return cookie.Id, true
	}
	return "", false
}

func (c *Context) GetLoginInfo() (*LoginContext, bool) {
	return c.GetLoginInfoAs(USERCOOKIENAME)
}

// EncryptLoginContext encrypt the login context to a token
func (c *Context) EncryptLoginContext(lctx *LoginContext) string {
	return c.Server.loginSaver.GetToken(lctx)
}

func (c *Context) GetLoginInfoAs(name string) (*LoginContext, bool) {
	cookie, err := c.Server.loginSaver.GetLoginIdAs(c, name)
	if cookie != nil && err == nil {
		return cookie, true
	}
	return nil, false
}

func (c *Context) AddLoginIdAs(id interface{}, name string, setter ...LoginInfoSetter) string {
	if name == "" {
		name = "user"
	}
	var userid string
	switch rid := id.(type) {
	case string:
		userid = rid
	case int, int32, int64, uint, uint32, uint64:
		userid = fmt.Sprintf("%d", rid)
	default:
		userid = fmt.Sprintf("%s", id)
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

	return c.Server.loginSaver.AddLoginAs(c, lctx)

}

func (c *Context) AddLoginId(id interface{}, setter ...LoginInfoSetter) string {
	return c.AddLoginIdAs(id, "user", setter...)

}

// Delete the login cookie saved
func (c *Context) DelLogin() error {
	return c.DelLoginAs("user")
}

// Delete the login cookie as specified name
func (c *Context) DelLoginAs(name string) error {
	return c.Server.loginSaver.DeleteLoginAs(c, name)
}
