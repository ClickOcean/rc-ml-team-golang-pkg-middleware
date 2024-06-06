package ginmiddleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
)

func Trace(appName string) gin.HandlerFunc {
	return otelgin.Middleware(appName)
}
