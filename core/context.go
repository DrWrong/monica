package core

import (
	"encoding/json"
	"net/http"
	"strconv"
	"strings"

	"github.com/DrWrong/monica/core/inject"
)

type Handler interface{}

// the base class for the web framework
// it provide a context for each request
type Context struct {
	inject.Injector
	handlers []Handler
	index    int
	*http.Request
	Resp   http.ResponseWriter
	Kwargs map[string]string
	// a bool value which control wheather will go on processing
	stopProcess bool
}

func (c *Context) QueryInt(name string) (int, error) {
	value := c.FormValue(name)
	return strconv.Atoi(value)
}

func (c *Context) GetClientIp() string {
	return strings.Split(c.RemoteAddr, ":")[0]

}

func (c *Context) RenderJson(data interface{}) {
	c.Resp.Header().Set("Content-type", "application/json; charset=utf-8")
	res, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	c.Resp.Write(res)
	c.stopProcess = true
}

// get Cookie
// try get cookie from cookie itself
// if cookie not support for example: app rest then use query
func (c *Context) GetCookie(name string) string {
	sessionCookie, err := c.Cookie(name)
	if err != nil {
		return ""
	}
	return sessionCookie.Value
}

func (c *Context) SetCookie(cookie *http.Cookie) {
	http.SetCookie(c.Resp, cookie)
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
	}
	c.Map(c)
	return c
}
