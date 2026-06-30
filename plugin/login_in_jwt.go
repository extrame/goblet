package plugin

import (
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/extrame/goblet/v2"
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
	Headers  []string
	duration time.Duration
	headers  map[string]string
}

func (j *_JwtLoginPlugin) AddCfgAndInit(server *goblet.Server) error {

	server.AddConfig("jwt", j)

	//get headers from config
	if j.Headers == nil {
		j.Headers = []string{"authorization"}
	}
	j.headers = make(map[string]string)
	for _, h := range j.Headers {
		//if h is xxx:yyy, means xxx is key and yyy is actual header name
		key, actual, ok := strings.Cut(h, ":")
		if ok {
			j.headers[strings.ToLower(key)] = actual
		} else {
			j.headers[strings.ToLower(h)] = h
		}
	}

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

func (l *_JwtLoginPlugin) AddLoginAs(ctx *goblet.Context, lctx *goblet.LoginContext) string {
	token := l.GetToken(lctx)
	if l.headers["authorization"] != "" {
		ctx.SetHeader(l.headers["authorization"], token)
	}
	//Token-Expire 字段，值为 Unix 时间戳（毫秒）
	if l.headers["token-expire"] != "" {
		ctx.SetHeader(l.headers["token-expire"], fmt.Sprintf("%d", lctx.Deadline.Unix()*1000))
	}
	return token
}

func (l *_JwtLoginPlugin) GetToken(lctx *goblet.LoginContext) string {
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
	if err != nil {
		return ""
	}
	token := fmt.Sprintf("Bearer %s", string(b))
	return token
}

func (l *_JwtLoginPlugin) GetLoginIdAs(ctx *goblet.Context, key string) (*goblet.LoginContext, error) {
	auth := ctx.Request.Header.Get("Authorization")
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
				exp, ok := token.Claims().Expiration()
				if ok {
					result.SetDeadline(exp)
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
