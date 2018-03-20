package middleware

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/garyburd/redigo/redis"
)

type Context interface {
	GetCookie(string) string
	SetCookie(*http.Cookie)
}

// 定义session接口
type Sessioner interface {
	// 设置key value
	Set(string, interface{}) error
	// 获取 key, value  value 需要是一个指针
	Get(key string, value interface{}) error
	// 删除某个key
	Delete(string) error
	// 获取session id
	ID() string
	// 清空session
	Flush() error
}

type Provider interface {
	StartSession(opt *Options) (Sessioner, error)
	Exist(sid string) (bool, error)
	Destroy(sid string) error
	GetSessioner(sid string) (Sessioner, error)
}

type Options struct {
	// session backend 的类型， 目前只实现了`redis`一种类型
	Provider       string

	// session提供者的配置信息 对于不同的backend 可能需要不同的配置
	// redis 的配置信息如下
	// key: "address" type: "string"  redis的地址 如 `127.0.0.1:6395`
	// key: "db"  type: "int" redis使用的db 如 `0`
	ProviderConfig map[string]interface{}
	// cookie的名称
	CookieName     string
	// cookie的路径
	CookiePath     string
	// cookie的有效时间 为0 时表示永不过期
	Cookielifttime int
	// cookie生效的域名
	Domain         string

	// session id的长度
	IDLength       int
}

func NewSessioner(options *Options) func(Context) Sessioner {
	manager := NewManager(options)
	return func(ctx Context) Sessioner {
		session, err := manager.Start(ctx)
		if err != nil {
			panic(err)
		}
		return session
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
		provider = newRedisProvider(manager.opt)
	default:
		panic(fmt.Sprintf("provider %s not support", manager.opt.Provider))
	}
	return provider
}

func (manager *Manager) Start(ctx Context) (Sessioner, error) {
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
		cookie.Expires = time.Now().Add(time.Duration(366*24) * time.Hour)
	}

	ctx.SetCookie(cookie)
	return sess, nil
}

type RedisProvider struct {
	redisPool      *redis.Pool
	CookieLifeTime int
	redisKeyPrefix string
}

func newRedisProvider(opt *Options) Provider {
	provider := &RedisProvider{
		CookieLifeTime: opt.Cookielifttime,
		redisKeyPrefix: opt.ProviderConfig["key_prefix"].(string),
		redisPool: &redis.Pool{
			MaxIdle:     5,
			IdleTimeout: 240 * time.Second,
			Dial: func() (redis.Conn, error) {
				c, err := redis.Dial("tcp", opt.ProviderConfig["address"].(string))
				if err != nil {
					println(err.Error())
					return nil, err
				}

				_, err = c.Do("SELECT", opt.ProviderConfig["db"])
				if err != nil {
					println(err.Error())
					return nil, err
				}
				return c, nil
			},
			TestOnBorrow: func(c redis.Conn, t time.Time) error {
				_, err := c.Do("PING")
				if err != nil {
					println(err.Error())
				}
				return err
			},
		},
	}
	return provider

}

func (provider *RedisProvider) getRedisKey(sid string) string {
	return fmt.Sprintf("%s_%s", provider.redisKeyPrefix, sid)
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
	if exist, err := redis.Bool(conn.Do("EXISTS", provider.getRedisKey(sid))); exist || err != nil {
		return nil, false, err
	}
	// 观察sid
	conn.Send("WATCH", provider.getRedisKey(sid))
	conn.Send("MULTI")
	conn.Send("HSET", provider.getRedisKey(sid), "createtime", time.Now().Unix())
	if option.Cookielifttime > 0 {
		conn.Send("EXPIRE", provider.getRedisKey(sid), option.Cookielifttime)
	} else {
		conn.Send("EXPIRE", provider.getRedisKey(sid), 30*24*60*60)
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
		sid:           sid,
		RedisProvider: provider,
	}, nil
}

func (provider *RedisProvider) Destroy(sid string) error {
	conn := provider.redisPool.Get()
	defer conn.Close()
	conn.Send("DEL", provider.getRedisKey(sid))
	return conn.Flush()
}

func (provider *RedisProvider) Exist(sid string) (bool, error) {
	conn := provider.redisPool.Get()
	defer conn.Close()
	ok, err := redis.Bool(conn.Do("EXISTS", provider.getRedisKey(sid)))
	if provider.CookieLifeTime == 0 {
		conn.Send("EXPIRE", provider.getRedisKey(sid), 30*24*60*60)
		conn.Flush()
	}
	return ok, err
}

type RedisSessioner struct {
	*RedisProvider
	sid string
}

func (session *RedisSessioner) Rediskey() string {
	return session.getRedisKey(session.sid)
}

func (session *RedisSessioner) ID() string {
	return session.sid
}

func (session *RedisSessioner) Set(key string, value interface{}) error {
	conn := session.redisPool.Get()
	defer conn.Close()
	res, err := json.Marshal(value)
	if err != nil {
		return err
	}

	conn.Send("HSET", session.Rediskey(), key, res)
	return conn.Flush()
}

func (session *RedisSessioner) Get(key string, value interface{}) error {
	conn := session.redisPool.Get()
	defer conn.Close()
	encoded, err := redis.Bytes(conn.Do("HGET", session.Rediskey(), key))
	if err != nil {
		return err
	}

	return json.Unmarshal(encoded, value)
}

func (session *RedisSessioner) Delete(key string) error {
	conn := session.redisPool.Get()
	defer conn.Close()
	conn.Send("HDEL", session.Rediskey(), key)
	return conn.Flush()
}

func (session RedisSessioner) Flush() error {
	conn := session.redisPool.Get()
	defer conn.Close()
	conn.Send("DEL", session.Rediskey())
	return conn.Flush()
}
