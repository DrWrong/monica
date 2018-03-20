package webserver

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"

	"gopkg.in/macaron.v1"

	"git.domob-inc.cn/domob_pad/monica.git/middleware"
)

type ServerConfig struct {
	Port       int    // 端口号
	ServerMode string // 运行模式
	URLPrefix  string // URL 前缀
}

type WebServer struct {
	*macaron.Macaron
	Config *ServerConfig
}

func New(config *ServerConfig) *WebServer {
	return &WebServer{
		Macaron: macaron.New(),
		Config:  config,
	}
}

func (s *WebServer) Run() {
	s.SetURLPrefix(s.Config.URLPrefix)
	addr := fmt.Sprintf(":%d", s.Config.Port)
	switch s.Config.ServerMode {
	case "http":
		fmt.Printf("run http server on %s\n", addr)
		err := http.ListenAndServe(addr, s)
		if err != nil {
			panic(err)
		}
	case "fastcgi":
		fmt.Printf("run fastcgi server on %s\n", addr)

		listener, err := net.Listen("tcp", addr)
		if err != nil {
			panic(err)
		}
		err = fcgi.Serve(listener, s)
		if err != nil {
			panic(err)
		}
	default:
		panic("not implement server mode")

	}

}

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
