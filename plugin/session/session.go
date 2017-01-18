package session

import (
	"net/http"

	"log"

	"github.com/extrame/go-random"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
)

const sessionName = "goblet-session-id"

type Session struct {
	store sessionStore
}

func (s *Session) OnNewRequest(ctx *goblet.Context) error {
	if _, err := ctx.SignedCookie(sessionName); err != nil {
		s.addSession(ctx)
	}
	return nil
}

func (s *Session) Init() (err error) {
	return s.store.init()
}

func (s *Session) ParseConfig(prefix string) (err error) {
	store := toml.String(prefix+".store", "local")
	switch *store {
	case "local":
		s.store = &localStore{}
	}
	s.store.parseConfig(prefix)
	return
}

func (s *Session) addSession(ctx *goblet.Context) {
	cookie := new(http.Cookie)
	cookie.Name = sessionName
	cookie.Value = gorandom.RandomAlphabetic(32)
	cookie.Path = "/"
	ctx.AddSignedCookie(cookie)
}

func GetInt(ctx *goblet.Context, key string) (int, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getInt(c.Value, key); ok {
			return item, true
		}
	}
	return 0, false
}

func Get(ctx *goblet.Context, key string) (interface{}, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.get(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
}

func Store(ctx *goblet.Context, key string, val interface{}) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.GetCookie(sessionName); err != nil {
		log.Fatal("session plugin is not inited currect!")
	} else {
		s.store.storeForUser(c.Value, key, val)
	}
}
