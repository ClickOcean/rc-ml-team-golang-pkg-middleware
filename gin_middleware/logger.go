package ginmiddleware

import (
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/exp/slices"
)

type logger interface {
	Info(msg string, args ...any)
	Error(msg string, err error, args ...any)
}

func StructuredLogger(
	l logger,
	excludedPath ...string,
) gin.HandlerFunc {
	msg := "middleware logger"
	return func(c *gin.Context) {

		start := time.Now()
		c.Next()
		finish := time.Now()

		path := c.Request.URL.Path
		raw := c.Request.URL.RawQuery
		if slices.Contains(excludedPath, path) {
			return
		}

		latency := finish.Sub(start).Milliseconds()

		if raw != "" {
			path = path + "?" + raw
		}

		args := []any{
			"client_id", c.ClientIP(),
			"method", c.Request.Method,
			"status_code", c.Writer.Status(),
			"body_size", c.Writer.Size(),
			"path", path,
			"latency", latency,
		}

		if c.Writer.Status() >= 500 {
			var err error
			if c.Errors.Last() != nil {
				err = c.Errors.Last().Err
			}
			l.Error(msg, err, args...)
		} else {
			l.Info(msg, args...)
		}
	}
}
