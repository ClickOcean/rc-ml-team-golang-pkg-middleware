package ginmiddleware

import (
	"time"

	"github.com/gin-gonic/gin"
)

type MetricsMonitor interface {
	ObserveIncomingRequests(statusCode int, method, path string)
	ObserveRequestsDuration(duration float64, statusCode int, method, path string)
}

func Metrics(monitor MetricsMonitor) gin.HandlerFunc {
	return func(ctx *gin.Context) {

		start := time.Now()
		ctx.Next()
		duration := time.Since(start).Seconds()

		status, method, path := ctx.Writer.Status(), ctx.Request.Method, ctx.FullPath()

		monitor.ObserveIncomingRequests(status, method, path)
		monitor.ObserveRequestsDuration(duration, status, method, path)
	}
}
