package main

import (
	"fmt"

	"github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica/middleware"
	"github.com/DrWrong/monica"
)

func routerConfigure() {
	fn := func(context *core.Context) {
		println("i am the handler")
		fmt.Printf("%+v\n", context)
	}
	core.AddMiddleware(middleware.LoggerHandler())
	core.Group("^/product",
		func() {
			core.Handle(`^/update/(?P<id>\d+)/$`, fn)
			core.Handle(`^/create/$`, fn)
			core.Group("^/partial/", func(){
				core.Handle("^/test/$", fn)
			})
		},
	)
	core.DebugRoute()
}

func main() {
	monica.BootStrap(routerConfigure)
}
