package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica/log"
	"github.com/garyburd/redigo/redis"
)

var (
	sessionLogger *log.MonicaLogger
)

func init() {
	sessionLogger = log.GetLogger("/monica/middleware/session")
}

type Sessioner interface {
	Set(string, interface{}) error
	Get(string) interface{}
	Delete(string) error
	ID() string
}

type Options struct {
	Provider       string
	ProviderConfig map[string]interface{}
	CookieName     string
	CookiePath     string
	Cookielifttime int64
	Domain         string
	IDLength       int
}

func (opt *Options) sessionId() string {
	b := make([]byte, opt.IDLength/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)

}

type Provider interface {
	Init(map[string]interface{}) error
	StartSession(opt *Options) Sessioner
	Exist(sid string) bool
	Destroy(sid string) error
	GetSessioner(sid string) Sessioner
}

type Manager struct {
	provider Provider
	opt      *Options
}

func NewManager(opt *Options) *Manager {
	manager := &Manager{
		opt: opt,
	}
	manager.provider = manager.getProvider()
	return manager
}

func (manager *Manager) getProvider() Provider {
	var provider Provider
	switch manager.opt.Provider {
	case "redis":
		provider = newRedisProvider()
	default:
		panic(fmt.Sprintf("provider %s not support", manager.opt.Provider))
	}
	provider.Init(manager.opt.ProviderConfig)
	return provider
}

func (manager *Manager) Start(ctx *core.Context) (Sessioner, error) {
	sid := ctx.GetCookie(manager.opt.CookieName)
	if len(sid) > 0 && manager.provider.Exist(sid) {
		return manager.provider.GetSessioner(sid), nil
	}
	sess := manager.provider.StartSession(manager.opt)
	cookie := &http.Cookie{
		Name:   manager.opt.CookieName,
		Value:  sess.ID(),
		Path:   manager.opt.CookiePath,
		Domain: manager.opt.Domain,
	}
	if manager.opt.Cookielifttime > 0 {
		cookie.Expires = time.Now().Add(time.Duration(manager.opt.Cookielifttime) * time.Second)
	}

	ctx.SetCookie(cookie)
	return sess, nil
}

type RedisProvider struct {
	redisPool *redis.Pool
	sync.Mutex
}

func newRedisProvider() Provider {
	return &RedisProvider{}
}

func (provider *RedisProvider) Init(config map[string]interface{}) error {
	provider.redisPool = &redis.Pool{
		MaxIdle:     5,
		IdleTimeout: 240 * time.Second,
		Dial: func() (redis.Conn, error) {
			c, err := redis.Dial("tcp", config["address"].(string))
			if err != nil {
				sessionLogger.Error(err.Error())
				return nil, err
			}

			_, err = c.Do("SELECT", config["db"])
			if err != nil {
				sessionLogger.Error(err.Error())
				return nil, err
			}
			return c, nil
		},
		TestOnBorrow: func(c redis.Conn, t time.Time) error {
			_, err := c.Do("PING")
			if err != nil {
				sessionLogger.Error(err.Error())
			}
			return err
		},
	}
	return nil
}

func getRedisKey(sid string) string {
	return fmt.Sprintf("MONICA_SESSION_%s", sid)
}

func (provider *RedisProvider) StartSession(option *Options) Sessioner {
	conn := provider.redisPool.Get()
	defer conn.Close()
	for {
		sid := option.sessionId()
		sess, ok := provider.startSession(sid, conn, option)
		if ok {
			return sess
		}
	}
}

// check - lock - check
func (provider *RedisProvider) startSession(sid string, conn redis.Conn, option *Options) (Sessioner, bool) {
	if exist, _ := redis.Bool(conn.Do("EXISTS", getRedisKey(sid))); exist {
		return nil, false
	}
	provider.Lock()
	defer provider.Unlock()
	if exist, _ := redis.Bool(conn.Do("EXISTS", getRedisKey(sid))); exist {
		return nil, false
	}
	conn.Send("MULTI")
	conn.Send("HSET", getRedisKey(sid), "createtime", time.Now().Unix())
	if option.Cookielifttime > 0 {
		conn.Send("EXPIRE", getRedisKey(sid), option.Cookielifttime)
	}
	conn.Send("EXEC")
	conn.Flush()
	return provider.GetSessioner(sid), true
}

func (provider *RedisProvider) GetSessioner(sid string) Sessioner {
	return &RedisSessioner{
		redisPool: provider.redisPool,
		sid:       sid,
	}
}

func (provider *RedisProvider) Destroy(sid string) error {
	conn := provider.redisPool.Get()
	defer conn.Close()
	conn.Send("DEL", getRedisKey(sid))
	return conn.Flush()
}

func (provider *RedisProvider) Exist(sid string) bool {
	conn := provider.redisPool.Get()
	defer conn.Close()
	ok, _ := redis.Bool(conn.Do("EXISTS", getRedisKey(sid)))
	return ok
}

type RedisSessioner struct {
	redisPool *redis.Pool
	sid       string
}

func (session *RedisSessioner) ID() string {
	return session.sid
}

func (session *RedisSessioner) Set(key string, value interface{}) error {
	conn := session.redisPool.Get()
	defer conn.Close()

	conn.Send("HSET", getRedisKey(session.sid), key, value)
	return conn.Flush()
}

// TODO: interface convert
func (session *RedisSessioner) Get(key string) interface{} {
	conn := session.redisPool.Get()
	defer conn.Close()
	res, _ := conn.Do("HGET", getRedisKey(session.sid), key)
	return res
}

func (session *RedisSessioner) Delete(key string) error {
	conn := session.redisPool.Get()
	defer conn.Close()
	conn.Send("HDEL", getRedisKey(session.sid), key)
	return conn.Flush()
}
