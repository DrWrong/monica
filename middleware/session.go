package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"net/http"
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

// 定义session接口
type Sessioner interface {
	Set(string, interface{}) error
	Get(string) (interface{}, error)
	Delete(string) error
	ID() string
}

type Provider interface {
	Init(map[string]interface{}) error
	StartSession(opt *Options) (Sessioner, error)
	Exist(sid string) (bool, error)
	Destroy(sid string) error
	GetSessioner(sid string) (Sessioner, error)
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

func NewSessioner(options *Options) core.Handler {
	manager := NewManager(options)
	return func(ctx *core.Context) {
		session, err := manager.Start(ctx)
		if err != nil {
			panic(err)
		}
		ctx.Map(session)
	}
}

// 获取session id
func (opt *Options) sessionId() string {
	b := make([]byte, opt.IDLength/2)
	if _, err := rand.Read(b); err != nil {
		panic(err)
	}

	return hex.EncodeToString(b)

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
		provider = newRedisProvider(int(manager.opt.Cookielifttime))
	default:
		panic(fmt.Sprintf("provider %s not support", manager.opt.Provider))
	}
	provider.Init(manager.opt.ProviderConfig)
	return provider
}

func (manager *Manager) Start(ctx *core.Context) (Sessioner, error) {
	sid := ctx.GetCookie(manager.opt.CookieName)
	if len(sid) > 0 {
		if ok, err := manager.provider.Exist(sid); err != nil {
			return nil, err
		} else if ok {
			return manager.provider.GetSessioner(sid)
		}
	}
	sess, err := manager.provider.StartSession(manager.opt)
	if err != nil {
		return nil, err
	}
	cookie := &http.Cookie{
		Name:   manager.opt.CookieName,
		Value:  sess.ID(),
		Path:   manager.opt.CookiePath,
		Domain: manager.opt.Domain,
	}
	if manager.opt.Cookielifttime > 0 {
		cookie.Expires = time.Now().Add(time.Duration(manager.opt.Cookielifttime) * time.Second)
	} else {
		cookie.Expires = time.Now().Add(time.Duration(366 * 24) * time.Hour)
	}

	ctx.SetCookie(cookie)
	return sess, nil
}

type RedisProvider struct {
	redisPool      *redis.Pool
	CookieLifeTime int
}

func newRedisProvider(cookieLifetime int) Provider {
	return &RedisProvider{CookieLifeTime: cookieLifetime}
}

// 初始化redis pool
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

func (provider *RedisProvider) StartSession(option *Options) (Sessioner, error) {
	conn := provider.redisPool.Get()
	defer conn.Close()
	for {
		sid := option.sessionId()
		sess, ok, err := provider.startSession(sid, conn, option)
		if err != nil {
			return nil, err
		}

		if ok {
			return sess, nil
		}
	}
}

// check - lock - check
func (provider *RedisProvider) startSession(sid string, conn redis.Conn, option *Options) (Sessioner, bool, error) {
	// 检查随机生成的session是否被占用了
	if exist, err := redis.Bool(conn.Do("EXISTS", getRedisKey(sid))); exist || err != nil {
		return nil, false, err
	}
	// 观察sid
	conn.Send("WATCH", getRedisKey(sid))
	conn.Send("MULTI")
	conn.Send("HSET", getRedisKey(sid), "createtime", time.Now().Unix())
	if option.Cookielifttime > 0 {
		conn.Send("EXPIRE", getRedisKey(sid), option.Cookielifttime)
	} else {
		conn.Send("EXPIRE", getRedisKey(sid), 30*24*60*60)
	}

	replay, err := conn.Do("EXEC")
	if err != nil {
		return nil, false, err
	}
	if replay == nil {
		return nil, false, nil
	}
	sessioner, err := provider.GetSessioner(sid)
	return sessioner, true, err
}

func (provider *RedisProvider) GetSessioner(sid string) (Sessioner, error) {
	return &RedisSessioner{
		redisPool: provider.redisPool,
		sid:       sid,
	}, nil
}

func (provider *RedisProvider) Destroy(sid string) error {
	conn := provider.redisPool.Get()
	defer conn.Close()
	conn.Send("DEL", getRedisKey(sid))
	return conn.Flush()
}

func (provider *RedisProvider) Exist(sid string) (bool, error) {
	conn := provider.redisPool.Get()
	defer conn.Close()
	ok, err := redis.Bool(conn.Do("EXISTS", getRedisKey(sid)))
	if provider.CookieLifeTime == 0 {
		conn.Send("EXPIRE", getRedisKey(sid), 30*24*60*60)
		conn.Flush()
	}
	return ok, err
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
func (session *RedisSessioner) Get(key string) (interface{}, error) {
	conn := session.redisPool.Get()
	defer conn.Close()
	return conn.Do("HGET", getRedisKey(session.sid), key)
}

func (session *RedisSessioner) Delete(key string) error {
	conn := session.redisPool.Get()
	defer conn.Close()
	conn.Send("HDEL", getRedisKey(session.sid), key)
	return conn.Flush()
}
