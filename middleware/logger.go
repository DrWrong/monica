package middleware

import (
	"time"

	"github.com/DrWrong/monica/core"
	"github.com/DrWrong/monica/log"
)

func LoggerHandler() core.Handler {
	logger := log.GetLogger("/monica/middleware/logger")
	return func(ctx *core.Context) {
		t := time.Now()
		ctx.Next()
		logger.Infof("%s %dms",
			ctx.Request.URL.String(),
			int(time.Since(t).Seconds()*1000),
		)
	}
}
