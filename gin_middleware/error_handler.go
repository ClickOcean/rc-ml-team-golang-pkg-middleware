package ginmiddleware

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type APIError interface {
	StatusCode() int
	JSONObj() any
}

func ErrorHandling() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		ctx.Next()

		if ctx.Errors.Last() == nil {
			return
		}

		if apiErr, ok := ctx.Errors.Last().Err.(APIError); ok {
			ctx.AbortWithStatusJSON(apiErr.StatusCode(), apiErr.JSONObj())
			return
		}

		ctx.AbortWithStatus(http.StatusInternalServerError)
	}
}
