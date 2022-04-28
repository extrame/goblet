package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/extrame/goblet"
)

var LoginInHead = new(_loginInHead)

type _loginInHead struct {
}

func (l *_loginInHead) AddLoginAs(ctx *goblet.Context, name string, id string, timeduration ...time.Duration) {
	var hashValue = ctx.Server.Hash(id)
	ctx.SetHeader("Authorization", fmt.Sprintf("Basic %s %s %s", name, id, hashValue))
}
func (l *_loginInHead) GetLoginIdAs(ctx *goblet.Context, key string) (string, error) {
	auth := ctx.ReqHeader().Get("Authorization")
	if auth != "" {
		parts := strings.Split(auth, " ")
		if len(parts) == 4 {
			if parts[0] == "Basic" && parts[1] == key && parts[4] == ctx.Server.Hash(parts[3]) {
				return parts[3], nil
			}
		}
	}
	return "", errors.New("NOT VALID LOGIN INFO:" + auth)
}
