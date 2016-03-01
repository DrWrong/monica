package core

import (
	"fmt"
	"net"
	"net/http"
	"net/http/fcgi"
	"strings"

	"github.com/DrWrong/monica/config"
)

type MonicaWebServer struct {
	Addr       string
	ServerMode string
	URLPrefix  string
}

func (server *MonicaWebServer) ServeHTTP(rw http.ResponseWriter, req *http.Request) {
	req.URL.Path = strings.TrimPrefix(req.URL.Path, server.URLPrefix)
	handler, err := GetProcessor(req.URL.Path)
	if err != nil {
		http.NotFound(rw, req)
	} else {
		handler(rw, req)
	}
}

func (server *MonicaWebServer) Run() {
	switch server.ServerMode {
	case "http":
		fmt.Printf("run http server on %s\n", server.Addr)
		err := http.ListenAndServe(server.Addr, server)
		if err != nil {
			panic(err)
		}
	case "fastcgi":
		listener, err := net.Listen("tcp", server.Addr)
		if err != nil {
			panic(err)
		}
		err = fcgi.Serve(listener, server)
		if err != nil {
			panic(err)
		}
	default:
		panic("not implement server mode")

	}
}

func NewServer() *MonicaWebServer {
	serverConfig := config.GetGlobalServerOption()
	return &MonicaWebServer{
		Addr:       fmt.Sprintf(":%d", serverConfig.ServerPort),
		ServerMode: serverConfig.ServerMode,
		URLPrefix:  serverConfig.URLPrefix,
	}
}
