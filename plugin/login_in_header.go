package plugin

import (
	"errors"
	"fmt"
	"strings"

	"github.com/extrame/goblet"
)

var LoginInHead = new(_loginInHead)

type _loginInHead struct {
}

func (l *_loginInHead) AddLoginAs(ctx *goblet.Context, lctx *goblet.LoginContext) string {
	var hashValue = ctx.Server.Hash(lctx.Id)
	var token = fmt.Sprintf("Basic %s:%s:%s", lctx.Name, lctx.Id, hashValue)
	ctx.SetHeader("Authorization", token)
	return token
}
func (l *_loginInHead) GetLoginIdAs(ctx *goblet.Context, key string) (*goblet.LoginContext, error) {
	auth := ctx.ReqHeader().Get("Authorization")
	if auth != "" && strings.HasPrefix(auth, "Basic ") {
		auth = strings.TrimPrefix(auth, "Basic ")
		parts := strings.Split(auth, ":")
		if len(parts) == 3 {
			if parts[0] == key && parts[2] == ctx.Server.Hash(parts[1]) {
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
