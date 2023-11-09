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

// New create a new LoginAsJwt plugin, secret is the secret key for jwt, idKey is the key for id in jwt
func JWT() *_JwtLoginPlugin {
	return &_JwtLoginPlugin{}
}

type _JwtLoginPlugin struct {
	Secret   string
	secret   []byte
	method   crypto.SigningMethod
	Issuer   string
	Alg      string
	Duration string
	duration time.Duration
}

func (j *_JwtLoginPlugin) AddCfgAndInit(server *goblet.Server) error {

	server.AddConfig("jwt", j)

	m := jws.GetSigningMethod(j.Alg)
	if m == nil {
		return errors.New("NOT VALID SIGNING METHOD:" + j.Alg)
	}

	duration, err := time.ParseDuration(j.Duration)
	if err != nil {
		j.duration = 24 * time.Hour
	} else {
		j.duration = duration
	}

	j.method = m
	j.secret = []byte(j.Secret)

	return nil
}

func (l *_JwtLoginPlugin) AddLoginAs(ctx *goblet.Context, lctx *goblet.LoginContext) {
	var claims = make(jws.Claims)
	claims.Set(lctx.Name+"Id", lctx.Id)
	j := jws.NewJWT(claims, l.method)
	j.Claims().SetIssuer(l.Issuer)
	j.Claims().SetExpiration(*lctx.Deadline)

	if lctx.Attrs != nil {
		for k, v := range lctx.Attrs {
			j.Claims().Set(k, v)
		}
	}

	b, err := j.Serialize(l.secret)
	if err == nil {
		ctx.SetHeader("Authorization", fmt.Sprintf("Bearer %s", string(b)))
	}
}

func (l *_JwtLoginPlugin) GetLoginIdAs(ctx *goblet.Context, key string) (*goblet.LoginContext, error) {
	auth := ctx.ReqHeader().Get("Authorization")
	if auth != "" {
		auth = strings.TrimPrefix(auth, "Bearer ")
		token, err := jws.ParseJWT([]byte(auth))
		if err == nil {
			err = token.Validate(l.secret)
			if err == nil {
				var id = token.Claims().Get(key + "Id")
				if id == nil {
					return nil, errors.New("NOT EXISTED LOGIN INFO: " + auth)
				}
				var result = &goblet.LoginContext{
					Name: key,
					Id:   id.(string),
				}

				if result.Attrs == nil {
					result.Attrs = make(map[string]interface{})
				}

				for k, v := range token.Claims() {
					if k != key+"Id" && k != "exp" && k != "nbf" && k != "iat" {
						result.Attrs[k] = v
					}
				}

				return result, nil
			}
		}
	}
	return nil, errors.New("NOT VALID LOGIN INFO: " + auth)
}

func (l *_JwtLoginPlugin) DeleteLoginAs(ctx *goblet.Context, key string) error {
	ctx.SetHeader("Authorization", "")
	return nil
}
