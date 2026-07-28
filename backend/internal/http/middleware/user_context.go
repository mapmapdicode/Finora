package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func currentUserID(c *gin.Context) string {
	if value, ok := c.Get("user_id"); ok {
		if id, ok := value.(string); ok {
			return strings.TrimSpace(id)
		}
	}
	return ""
}
