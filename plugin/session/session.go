package session

import (
	"github.com/extrame/go-random"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
	"log"
	"net/http"
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
	case "redis":
		s.store = &redisStore{}
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

func GetInts(ctx *goblet.Context, key string) ([]int, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getInts(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
}

func GetInt64(ctx *goblet.Context, key string) (int64, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getInt64(c.Value, key); ok {
			return item, true
		}
	}
	return 0, false
}

func GetInt64Map(ctx *goblet.Context, key string) (map[string]int64, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getInt64Map(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
}

func GetIntMap(ctx *goblet.Context, key string) (map[string]int, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getIntMap(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
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

func GetFloat64(ctx *goblet.Context, key string) (float64, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getFloat64(c.Value, key); ok {
			return item, true
		}
	}
	return 0.0, false
}

func GetString(ctx *goblet.Context, key string) (string, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getString(c.Value, key); ok {
			return item, true
		}
	}
	return "", false
}

func GetStrings(ctx *goblet.Context, key string) ([]string, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getStrings(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
}

func GetStringMap(ctx *goblet.Context, key string) (map[string]string, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getStringMap(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
}

func GetBool(ctx *goblet.Context, key string) (bool, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getBool(c.Value, key); ok {
			return item, true
		}
	}
	return false, false
}

func GetBytes(ctx *goblet.Context, key string) ([]byte, bool) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.SignedCookie(sessionName); err == nil {
		if item, ok := s.store.getBytes(c.Value, key); ok {
			return item, true
		}
	}
	return nil, false
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

func RemoveItem(ctx *goblet.Context, key string) {
	s := ctx.Server.GetPlugin("session").(*Session)
	if c, err := ctx.GetCookie(sessionName); err != nil {
		log.Fatal("session plugin is not inited currect!")
	} else {
		s.store.removeItem(c.Value, key)
	}
}
