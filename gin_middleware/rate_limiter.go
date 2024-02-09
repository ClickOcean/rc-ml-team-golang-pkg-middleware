package ginmiddleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

func RateLimit(limit int) gin.HandlerFunc {
	l := rate.Every(time.Second)
	limiter := rate.NewLimiter(l, limit)

	return func(ctx *gin.Context) {
		if !limiter.Allow() {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
			return
		}
		ctx.Next()
	}
}
