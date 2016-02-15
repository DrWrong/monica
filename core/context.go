package core

import (
	"github.com/DrWrong/monica/core/inject"
	"net/http"
)

type Handler interface{}

// the base class for the web framework
// it provide a context for each request
type Context struct {
	inject.Injector
	handlers []Handler
	index    int
	*http.Request
	Resp          http.ResponseWriter
	CookieSupport bool
	Kwargs        map[string]string
	// Data is used for a json response which
	Data map[string]interface{}
	// a bool value which control wheather will go on processing
	stopProcess bool
}

// get Cookie
// try get cookie from cookie itself
// if cookie not support for example: app rest then use query
func (c *Context) GetCookie(name string) string {
	if c.CookieSupport {
		sessionCookie, err := c.Cookie(name)
		if err != nil {
			return ""
		}
		return sessionCookie.Value
	}
	return c.FormValue(name)
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	if c.CookieSupport {
		http.SetCookie(c.Resp, cookie)
	} else {
		c.Data[cookie.Name] = cookie.Value
	}
}

// stop process
func (c *Context) Stop() {
	c.stopProcess = true
}

// yield another handler
func (c *Context) Next() {
	c.index += 1
	c.run()
}

func (c *Context) run() {
	for c.index < len(c.handlers) {
		if c.stopProcess {
			break
		}

		_, err := c.Invoke(c.handlers[c.index])

		if err != nil {
			panic(err)
		}
		c.index += 1
	}
}

func NewContext(resp http.ResponseWriter, req *http.Request) *Context {
	c := &Context{
		Injector:      inject.New(),
		Request:       req,
		Resp:          resp,
		CookieSupport: true,
	}
	c.Map(c)
	return c
}
