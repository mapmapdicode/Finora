package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type responseBufferWriter struct {
	gin.ResponseWriter
	body   *bytes.Buffer
	status int
}

func (w *responseBufferWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.Write(data)
}

func (w *responseBufferWriter) WriteString(s string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	return w.body.WriteString(s)
}

func (w *responseBufferWriter) WriteHeader(code int) {
	w.status = code
}

func ErrorEnvelope() gin.HandlerFunc {
	return func(c *gin.Context) {
		downstream := c.Writer
		writer := &responseBufferWriter{
			ResponseWriter: downstream,
			body:           &bytes.Buffer{},
			status:         http.StatusOK,
		}
		c.Writer = writer
		c.Next()
		c.Writer = downstream

		response := writer.body.Bytes()
		if writer.status >= http.StatusBadRequest {
			var payload map[string]any
			if err := json.Unmarshal(response, &payload); err == nil && payload != nil {
				if _, hasCode := payload["code"]; hasCode {
					if _, hasTrace := payload["traceId"]; !hasTrace {
						payload["traceId"] = requestIDFromContext(c)
					}
					if encoded, err := json.Marshal(payload); err == nil {
						response = encoded
					}
				}
			}
		}

		downstream.Header().Set("Content-Type", "application/json")
		downstream.WriteHeader(writer.status)
		if len(response) > 0 {
			_, _ = downstream.Write(response)
		}
	}
}

func requestIDFromContext(c *gin.Context) string {
	traceID, ok := c.Get("request_id")
	if !ok {
		return "trace-id-missing"
	}
	if value, ok := traceID.(string); ok {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return "trace-id-missing"
}
