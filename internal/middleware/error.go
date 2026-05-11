package middleware

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type AppError struct {
	Code     int
	Err      string
	Message  string // sent to client
	Internal string // logged server-side only, never sent to client
}

func (a *AppError) Error() string {
	return a.Message
}

func NewAppError(code int, err, message string) *AppError {
	return &AppError{
		Code:    code,
		Err:     err,
		Message: message,
	}
}

func (a *AppError) WithInternal(detail string) *AppError {
	a.Internal = detail
	return a
}

func ErrorHandler() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next()

		if len(c.Errors) == 0 {
			return
		}

		err := c.Errors.Last()

		if appErr, ok := asType[*AppError](err.Err); ok {
			c.JSON(appErr.Code, gin.H{
				"status":    appErr.Code,
				"error":     appErr.Err,
				"message":   appErr.Message,
				"path":      c.Request.URL.Path,
				"timestamp": time.Now().UTC().Format(time.RFC3339),
			})
			return
		}

		c.JSON(http.StatusInternalServerError, gin.H{
			"status":    http.StatusInternalServerError,
			"error":     "internal server error",
			"message":   "an unexpected error occured",
			"path":      c.Request.URL.Path,
			"timestamp": time.Now().UTC().Format(time.RFC3339),
		})
	}
}

func asType[T error](err error) (T, bool) {
	var target T
	ok := errors.As(err, &target)
	return target, ok
}
