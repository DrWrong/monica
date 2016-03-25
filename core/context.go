package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"strconv"
	"strings"
	"sync"

	"github.com/DrWrong/monica/core/inject"
	"github.com/DrWrong/monica/form"
	"github.com/DrWrong/monica/log"
	"github.com/astaxie/beego/validation"
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
	Kwargs        map[string]string
	parseFormOnce sync.Once
	// a bool value which control wheather will go on processing
	stopProcess bool
	logger      *log.MonicaLogger
}

func (c *Context) parseForm() {
	c.ParseForm()
	for key, value := range c.Kwargs {
		c.Form.Set(key, value)
	}
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

	c.logger.Debugf("processing request %+v ok: response: %s", c.Request, res)
	if err != nil {
		panic(err)
	}
	c.Resp.Write(res)
	c.stopProcess = true
}

func (c *Context) Render(data []byte) {
	c.Resp.Write(data)
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

func (c *Context) Bind(formStruct interface{}) (bool, error) {
	form.Bind(c.Form, formStruct)
	valid := validation.Validation{}
	b, err := valid.Valid(formStruct)
	if err != nil {
		return false, err
	}
	if b {
		return true, nil
	}
	res := ""
	for _, err := range valid.Errors {
		res += fmt.Sprintf("%s: %s;", err.Key, err.Message)
	}
	return false, errors.New(res)
}

func (c *Context) run() {
	c.parseFormOnce.Do(c.parseForm)
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
		Injector: inject.New(),
		Request:  req,
		Resp:     resp,
		logger:   log.GetLogger("/monica/core/context"),
	}
	c.Map(c)
	return c
}
