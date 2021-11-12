package redissession

import (
	"errors"
	"log"
	"net/http"

	gorandom "github.com/extrame/go-random"
	"github.com/extrame/goblet"
	"github.com/garyburd/redigo/redis"
)

//TODO

type RedisSession struct {
	Address     string `goblet:"address,localhost:6379"`
	Pwd         string `goblet:"pwd"`
	Db          int64  `goblet:"db"`
	SessionType string `goblet:"type,cookie"`
	TokenName   string `goblet:"token_name,token"`
}

var redisPool *redis.Pool
var cookietype string
var tokenName string
var PoolMaxIdle = 10

func (r *RedisSession) AddCfgAndInit(server *goblet.Server) error {
	err := server.AddConfig("redissession", r)
	if err == nil {
		return r.Init(server)
	}
	return err
}

func (s *RedisSession) OnNewRequest(ctx *goblet.Context) error {
	if cookietype == "cookie" {
		if _, err := ctx.SignedCookie(tokenName); err != nil {
			s.addSession(ctx)
		}
	}
	return nil
}

func (s *RedisSession) addSession(ctx *goblet.Context) {
	cookie := new(http.Cookie)
	cookie.Name = tokenName
	cookie.Value = gorandom.RandomAlphabetic(32)
	cookie.Path = "/"
	ctx.AddSignedCookie(cookie)
}

func addSessionWithValue(ctx *goblet.Context, value string) {
	cookie := new(http.Cookie)
	cookie.Name = tokenName
	cookie.Value = value
	cookie.Path = "/"
	ctx.AddSignedCookie(cookie)
}

func (r *RedisSession) Init(server *goblet.Server) error {
	cookietype = r.SessionType
	tokenName = r.TokenName
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", r.Address)
		if err != nil {
			log.Println("--Redis--Connect redis fail:" + err.Error())
			return nil, err
		}
		if len(r.Pwd) > 0 {
			if _, err := c.Do("AUTH", r.Pwd); err != nil {
				c.Close()
				log.Println("--Redis--Auth redis fail:" + err.Error())
				return nil, err
			}
		}
		if _, err := c.Do("SELECT", r.Db); err != nil {
			c.Close()
			log.Println("--Redis--Select redis db fail:" + err.Error())
			return nil, err
		}
		return c, nil
	}, PoolMaxIdle)
	return nil
}

func getRegionId(ctx *goblet.Context) (result string, err error) {
	switch cookietype {
	case "cookie":
		tmp := new(http.Cookie)
		tmp, err = ctx.GetCookie(tokenName)
		result = tmp.Value
	case "form":
		result = ctx.FormValue(tokenName)
	case "header":
		result = ctx.ReqHeader().Get(tokenName)
	default:
		err = errors.New("Type must be cookie or form.Current is " + cookietype)
	}
	if result == "" {
		err = errors.New("HaskKey is empty.")
	}
	return result, err
}

func Store(cx *goblet.Context, key string, item interface{}) (err error) {
	hashkey, _ := getRegionId(cx)

	flag := false
	if hashkey == "" {
		flag = true
		hashkey = gorandom.RandomAlphabetic(32)
	}

	c := redisPool.Get()
	defer c.Close()

	if _, err = c.Do("HSET", hashkey, key, item); err != nil {
		log.Println(err.Error())
	}

	if flag {
		switch cookietype {
		case "cookie":
			addSessionWithValue(cx, hashkey)
		case "form":
			cx.AddRespond(tokenName, hashkey)
		case "header":
			cx.SetHeader(tokenName, hashkey)
		}
	}
	return err
}

func Exists(cx *goblet.Context, key string) (bool, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return false, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, key))
	if count == 0 {
		return false, err
	}
	return true, err
}

func Get(cx *goblet.Context, Key string) (interface{}, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		res, _ := redis.Values(c.Do("HGET", hashkey, Key))

		return res, err
	}
}

func GetBool(cx *goblet.Context, Key string) (bool, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return false, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return false, err
	} else {
		n, _ := redis.Bool(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetBytes(cx *goblet.Context, Key string) ([]byte, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.Bytes(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetFloat64(cx *goblet.Context, Key string) (float64, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return 0.0, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return 0, err
	} else {
		n, _ := redis.Float64(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetInt(cx *goblet.Context, Key string) (int, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return 0, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return 0, err
	} else {
		n, _ := redis.Int(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetInt64(cx *goblet.Context, Key string) (int64, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return 0, err
	}

	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return 0, err
	} else {
		n, _ := redis.Int64(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetIntMap(cx *goblet.Context, Key string) (map[string]int, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.IntMap(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetInt64Map(cx *goblet.Context, Key string) (map[string]int64, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.Int64Map(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetInts(cx *goblet.Context, Key string) ([]int, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.Ints(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetString(cx *goblet.Context, Key string) (string, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return "", err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return "", err
	} else {
		n, _ := redis.String(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetStrings(cx *goblet.Context, Key string) ([]string, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.Strings(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetStringMap(cx *goblet.Context, Key string) (map[string]string, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return nil, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return nil, err
	} else {
		n, _ := redis.StringMap(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func GetUint64(cx *goblet.Context, Key string) (uint64, error) {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return 0, err
	}
	c := redisPool.Get()
	defer c.Close()
	count, _ := redis.Int(c.Do("HEXISTS", hashkey, Key))
	if count == 0 {
		return 0, err
	} else {
		n, _ := redis.Uint64(c.Do("HGET", hashkey, Key))
		return n, err
	}
}

func RemoveItem(cx *goblet.Context, Key string) error {
	hashkey, err := getRegionId(cx)
	if err != nil {
		return err
	}
	c := redisPool.Get()
	defer c.Close()
	c.Do("HDEL", hashkey, Key)
	return err
}
