package middleware

import (
	"net/http"
	"strings"

	"firebase.google.com/go/v4/auth"
	"github.com/gin-gonic/gin"
)

func Auth(authClient *auth.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if !strings.HasPrefix(header, "Bearer ") {
			c.Error(&gin.Error{
				Err:  NewAppError(http.StatusUnauthorized, "unauthorized", "bearer token is required"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}

		token, err := authClient.VerifyIDTokenAndCheckRevoked(c.Request.Context(), strings.TrimPrefix(header, "Bearer "))
		if err != nil {
			c.Error(&gin.Error{
				Err:  NewAppError(http.StatusUnauthorized, "unauthorized", "invalid token"),
				Type: gin.ErrorTypePublic,
			})
			c.Abort()
			return
		}

		c.Set("firebaseToken", token)
		c.Next()
	}
}
