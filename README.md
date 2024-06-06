## Getting started

1. go env -w GOPRIVATE=gitlab.ml.rc.dating.com/rc-ml-team/golang-pkg
2. go get gitlab.ml.rc.dating.com/rc-ml-team/golang-pkg/middleware

## Example of usage (Gin middlewares)

```go
	import (
        "github.com/gin-gonic/gin"
        ginmiddleware "gitlab.ml.rc.dating.com/rc-ml-team/golang-pkg/middleware/gin_middleware"
    )


    ginEngine := gin.New()
    ginEngine.Use(ginmiddleware.Trace("appName"))
     
```