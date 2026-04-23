package plugin

import (
	"crypto/sha1"
	"errors"
	"fmt"
	"strings"

	"github.com/extrame/goblet/v2"
)

// New create a new LoginAsHeader plugin
func LoginInHead() *_loginInHead {
	return &_loginInHead{}
}

type _loginInHead struct {
	Secret string
}

func (j *_loginInHead) AddCfgAndInit(server *goblet.Server) error {
	j.Secret = server.Config.Basic.HashSecret
	return nil
}

// Hash 获得一个字符串的加密版本
func (s *_loginInHead) Hash(str string) string {
	hash := sha1.New()
	hash.Write([]byte(str))
	hash.Write([]byte(s.Secret))
	return fmt.Sprintf("%x", hash.Sum(nil))
}

func (l *_loginInHead) AddLoginAs(ctx *goblet.Context, lctx *goblet.LoginContext) string {
	token := l.GetToken(lctx)
	ctx.SetHeader("Authorization", token)
	return token
}

func (l *_loginInHead) GetToken(lctx *goblet.LoginContext) string {
	var hashValue = l.Hash(lctx.Id)
	var token = fmt.Sprintf("Basic %s:%s:%s", lctx.Name, lctx.Id, hashValue)
	return token
}

func (l *_loginInHead) GetLoginIdAs(ctx *goblet.Context, key string) (*goblet.LoginContext, error) {
	auth := ctx.Request.Header.Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Basic ") {
		auth = strings.TrimPrefix(auth, "Basic ")
		parts := strings.Split(auth, ":")
		if len(parts) == 3 {
			if parts[0] == key && parts[2] == l.Hash(parts[1]) {
				return &goblet.LoginContext{
					Name: key,
					Id:   parts[1],
				}, nil
			}
		}
	}
	return nil, errors.New("NOT VALID LOGIN INFO:" + auth)
}

func (l *_loginInHead) DeleteLoginAs(ctx *goblet.Context, key string) error {
	ctx.SetHeader("Authorization", "")
	return nil
}
