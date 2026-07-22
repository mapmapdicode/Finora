package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func RequestID(headerName string) gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(headerName)
		if requestID == "" {
			requestID = uuid.NewString()
		}
		c.Header(headerName, requestID)
		c.Set("request_id", requestID)
		c.Next()
	}
}
