package middleware

import (
	"net/http"
	"strings"

	"wealthos-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func UserContextMiddleware(staticToken, jwtSecret string) gin.HandlerFunc {
	staticToken = strings.TrimSpace(staticToken)
	jwtSecret = strings.TrimSpace(jwtSecret)
	return func(c *gin.Context) {
		token := strings.TrimSpace(c.GetHeader("Authorization"))
		if strings.HasPrefix(strings.ToLower(token), "bearer ") {
			token = strings.TrimSpace(token[len("bearer "):])
		}
		if token == "" {
			token = c.GetHeader("X-Auth-Token")
		}
		token = strings.TrimSpace(token)
		if strings.HasPrefix(token, "token-") {
			userID := strings.TrimPrefix(token, "token-")
			if userID != "" {
				c.Set("user_id", userID)
				c.Next()
				return
			}
		}

		if staticToken != "" && token == staticToken {
			c.Set("user_id", "demo-user")
			c.Next()
			return
		}

		if token != "" && jwtSecret != "" {
			userID, err := auth.ParseToken(jwtSecret, token)
			if err == nil && strings.TrimSpace(userID) != "" {
				c.Set("user_id", userID)
				c.Next()
				return
			}
		}

		c.JSON(http.StatusUnauthorized, gin.H{"code": "UNAUTHORIZED", "message": "missing or invalid authentication"})
		c.Abort()
	}
}
