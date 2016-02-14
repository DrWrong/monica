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
	Req      *http.Request
	Resp     http.ResponseWriter
	Kwargs   map[string]string
	// Data is used for a json response which
	Data map[string]interface{}
	// a bool value which control wheather will go on processing
	stopProcess bool
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
		Injector: inject.New(),
		Req:      req,
		Resp:     resp,
	}
	c.Map(c)
	return c
}
