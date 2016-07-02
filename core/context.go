package core

import (
	"encoding/json"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/DrWrong/monica/core/inject"
	"github.com/DrWrong/monica/form"
	"github.com/DrWrong/monica/log"
	"github.com/astaxie/beego/validation"
)

type Handler interface{}

var privateIpNets [3]*net.IPNet

func init() {
	_, aNet, _ := net.ParseCIDR("10.0.0.0/8")
	_, bNet, _ := net.ParseCIDR("172.16.0.0/12")
	_, cNet, _ := net.ParseCIDR("192.168.0.0/16")
	privateIpNets = [3]*net.IPNet{aNet, bNet, cNet}
}

func isPrivateIp(ip net.IP) bool {
	for _, net := range privateIpNets {
		if net.Contains(ip) {
			return true
		}
	}
	return false
}

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
	timerStart  time.Time
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
	// firstly get remote addr and judge if it is a global ip address
	ipAddress := strings.Split(c.RemoteAddr, ":")[0]
	ip := net.ParseIP(ipAddress)
	if ip.IsGlobalUnicast() && !isPrivateIp(ip) {
		return ipAddress
	}
	// if not get ip from X-Forwarded-For
	forwardedIpAddress := c.Header.Get("X-Forwarded-For")
	// if X-Forwarded-For is blank then return real ip
	if forwardedIpAddress == "" {
		return ipAddress
	}
	// cause X-Forwarded-For can be set by user so we examine if it is a validate value
	forwardedIp := net.ParseIP(forwardedIpAddress)
	if forwardedIp == nil || forwardedIp.IsUnspecified() {
		panic("X-Forwarded-For not validate")
	}
	return forwardedIpAddress
}

func (c *Context) RenderJson(data interface{}) {
	c.Resp.Header().Set("Content-type", "application/json; charset=utf-8")
	res, err := json.Marshal(data)

	if err != nil {
		panic(err)
	}

	c.logger.Debugf(
		"REQUESET: %+v RESPONSE: %s COST: %f ms",
		c.Request, res, time.Since(c.timerStart).Seconds()*1000)

	c.Resp.Write(res)
	c.stopProcess = true
}

func (c *Context) Render(data []byte) {
	c.Resp.Write(data)
	c.stopProcess = true
}

// when error response 500
func (c *Context) ErrorResponse(err error) {
	c.Resp.Write([]byte(err.Error()))
	c.Resp.WriteHeader(500)
	c.Stop()
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

func (c *Context) Bind(formStruct interface{}, fields ...string) error {
	form.Bind(c.Form, formStruct, fields...)
	valid := validation.Validation{}
	b, err := valid.Valid(formStruct)
	if err != nil {
		return err
	}
	if b {
		return nil
	}
	res := ""
	for _, err := range valid.Errors {
		res += fmt.Sprintf("%s\n", err.Message)
	}
	return errors.New(res)
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
		Injector:   inject.New(),
		Request:    req,
		Resp:       resp,
		logger:     log.GetLogger("/monica/core/context"),
		timerStart: time.Now(),
	}
	c.Map(c)
	return c
}
