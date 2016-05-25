package middleware

import (
	"runtime"

	"github.com/DrWrong/monica/log"
	"github.com/DrWrong/monica/core"

)


func ReconverHandler() core.Handler {
	errorLogger := log.GetLogger("/monica/middleware/errorLogger")
	return func (ctx *core.Context) {
		defer func() {
			if r := recover(); r != nil {
				trace := make([]byte, 10240)
				runtime.Stack(trace, false)
				errorLogger.Errorf(
					"server request %+v error: %+v\n %s", ctx.Request, r, trace)

				ctx.Resp.Write([]byte("monica recover.... error happend"))
				ctx.Resp.WriteHeader(500)
				ctx.Stop()
			}

		}()
		ctx.Next()
	}

}
