package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/extrame/goblet"
	"github.com/extrame/jose/crypto"
	"github.com/extrame/jose/jws"
)

//New create a new LoginAsJwt plugin, secret is the secret key for jwt, idKey is the key for id in jwt
func New(secret string, opts ...JwtOption) *_LoginAsJwt {
	var jwt = &_LoginAsJwt{
		secret: secret,
	}
	for _, opt := range opts {
		err := opt(jwt)
		if err != nil {
			panic(err)
		}
	}
	return jwt
}

//JwtOption is the option for jwt
type JwtOption func(*_LoginAsJwt) error

func UseSigningMethod(method string) JwtOption {
	return func(jwt *_LoginAsJwt) error {
		m := jws.GetSigningMethod(method)
		if m == nil {
			return errors.New("NOT VALID SIGNING METHOD:" + method)
		}
		jwt.method = m
		return nil
	}
}

func UseIssuer(issuer string) JwtOption {
	return func(jwt *_LoginAsJwt) error {
		jwt.issuer = issuer
		return nil
	}
}

type _LoginAsJwt struct {
	secret string
	method crypto.SigningMethod
	issuer string
}

func (l *_LoginAsJwt) AddLoginAs(ctx *goblet.Context, name string, id string, timeduration ...time.Duration) {
	var claims jws.Claims
	claims.Set(name, id)
	j := jws.NewJWT(claims, l.method)
	j.Claims().SetIssuer(l.issuer)

	b, err := j.Serialize(l.secret)
	if err == nil {
		ctx.SetHeader("Authorization", fmt.Sprintf("Bearer %s", string(b)))
	}
}

func (l *_LoginAsJwt) GetLoginIdAs(ctx *goblet.Context, key string) (string, error) {
	auth := ctx.ReqHeader().Get("Authorization")
	if auth != "" {
		auth = strings.TrimPrefix(auth, "Bearer ")
		token, err := jws.ParseJWT([]byte(auth))
		if err == nil {
			err = token.Validate(l.secret)
			if err == nil {
				return token.Claims().Get(key).(string), nil
			}
		}
	}
	return "", errors.New("NOT VALID LOGIN INFO:" + auth)
}

func (l *_LoginAsJwt) DeleteLoginAs(ctx *goblet.Context, key string) error {
	ctx.SetHeader("Authorization", "")
	return nil
}
