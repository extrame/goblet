package redissession

import (
	"errors"
	"github.com/extrame/go-random"
	toml "github.com/extrame/go-toml-config"
	"github.com/extrame/goblet"
	"github.com/garyburd/redigo/redis"
	"log"
	"net/http"
)

//TODO

type Redis struct {
	address     *string
	pwd         *string
	db          *int64
	sessionType *string
	tokenName   *string
}

var redisPool *redis.Pool
var cookietype string
var tokenName string
var PoolMaxIdle = 10

func (r *Redis) ParseConfig(prefix string) error {
	log.Println("++++++++++++=", prefix)
	r.address = toml.String(prefix+".address", "localhost:6379")
	r.pwd = toml.String(prefix+".password", "")
	r.db = toml.Int64(prefix+".db", 0)
	r.sessionType = toml.String(prefix+".type", "cookie")
	r.tokenName = toml.String(prefix+".token_name", "token")
	return nil
}

func (s *Redis) OnNewRequest(ctx *goblet.Context) error {
	if cookietype == "cookie" {
		if _, err := ctx.SignedCookie(tokenName); err != nil {
			s.addSession(ctx)
		}
	}
	return nil
}

func (s *Redis) addSession(ctx *goblet.Context) {
	cookie := new(http.Cookie)
	cookie.Name = tokenName
	cookie.Value = gorandom.RandomAlphabetic(32)
	cookie.Path = "/"
	ctx.AddSignedCookie(cookie)
}

func (r *Redis) Init(server *goblet.Server) error {
	cookietype = *r.sessionType
	tokenName = *r.tokenName
	redisPool = redis.NewPool(func() (redis.Conn, error) {
		c, err := redis.Dial("tcp", *r.address)
		if err != nil {
			log.Println("--Redis--Connect redis fail:" + err.Error())
			return nil, err
		}
		if len(*r.pwd) > 0 {
			if _, err := c.Do("AUTH", *r.pwd); err != nil {
				c.Close()
				log.Println("--Redis--Auth redis fail:" + err.Error())
				return nil, err
			}
		}
		if _, err := c.Do("SELECT", *r.db); err != nil {
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
	hashkey := ""
	hashkey, err = getRegionId(cx)
	if err != nil {
		return err
	}

	c := redisPool.Get()
	defer c.Close()

	if _, err = c.Do("HSET", hashkey, key, item); err != nil {
		log.Println(err.Error())
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
