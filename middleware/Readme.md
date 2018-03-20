# monica middleware package 

在web开发中有一个middleware的概念， 指所有的请求都要先通过一系列middleware 然后在落到最后的handler函数上。

这个包中包含了一系列通用的middleware

## session middleware 

用于提供session层， 目前session的provider只有redis一种

使用示例

+ 初始化session

```golang 

# 以macaron为例
monica.WebServer.Use(webserver.MacaronSessioner(&middleware.Options{
	Provider:       "redis",
	ProviderConfig: providerConfig,
	CookieName:     "spring_promotion_sessionid",
	IDLength:       32,
}))


type SessionContext struct {
	*macaron.Context
}

func (c *SessionContext) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
}

func MacaronSessioner(options *middleware.Options) macaron.Handler {
	sessionFunc := middleware.NewSessioner(options)

	return func(c *macaron.Context) {
		sessioner := sessionFunc(&SessionContext{c})
		c.Map(sessioner)
	}
}


```

+ 使用

sessioner 有如下接口

```golang
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

```

+ session 配置

```golang 

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


```
