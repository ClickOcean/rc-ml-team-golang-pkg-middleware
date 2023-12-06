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
	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		duration := time.Since(start).Seconds()

		status, method, path := c.Writer.Status(), c.Request.Method, c.FullPath()

		monitor.ObserveIncomingRequests(status, method, path)
		monitor.ObserveRequestsDuration(duration, status, method, path)
	}
}
