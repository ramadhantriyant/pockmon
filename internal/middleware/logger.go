package middleware

import (
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

func Logger() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		c.Next()

		errMsg := ""
		if len(c.Errors) > 0 {
			last := c.Errors.Last()
			if appErr, ok := asType[*AppError](last.Err); ok && appErr.Internal != "" {
				errMsg = appErr.Internal
			} else {
				errMsg = last.Error()
			}
		}

		fmt.Printf(`{"timestamp":%q,"method":%q,"path":%q,"status":%d,"latency":%q,"error":%q}`+"\n",
			start.Format(time.RFC3339),
			c.Request.Method,
			c.Request.URL.Path,
			c.Writer.Status(),
			time.Since(start).String(),
			errMsg,
		)
	}
}
