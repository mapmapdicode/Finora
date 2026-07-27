package middleware

import (
	"bytes"
	"encoding/json"
	"fmt"
	"log"
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
			code := ""
			msg := ""

			if err := json.Unmarshal(response, &payload); err == nil && payload != nil {
				if cCode, hasCode := payload["code"]; hasCode {
					code = fmt.Sprintf("%v", cCode)
				}
				if cMsg, hasMsg := payload["message"]; hasMsg {
					msg = fmt.Sprintf("%v", cMsg)
				}
				if _, hasCode := payload["code"]; hasCode {
					if _, hasTrace := payload["traceId"]; !hasTrace {
						payload["traceId"] = requestIDFromContext(c)
					}
					if encoded, err := json.Marshal(payload); err == nil {
						response = encoded
					}
				}
			}

			if msg == "" && len(c.Errors) > 0 {
				msg = c.Errors.String()
			}
			if code == "" {
				code = strings.ReplaceAll(strings.ToUpper(http.StatusText(writer.status)), " ", "_")
			}

			traceID := requestIDFromContext(c)
			userID := currentUserID(c)
			if userID == "" {
				userID = "anonymous"
			}
			wsID := ""
			if val, ok := c.Get("workspace_id"); ok {
				if s, ok2 := val.(string); ok2 {
					wsID = s
				}
			}
			if wsID == "" {
				wsID = "none"
			}

			statusText := strings.ToUpper(http.StatusText(writer.status))
			if statusText == "" {
				statusText = "UNKNOWN"
			}

			reqMethod := "UNKNOWN"
			reqPath := "UNKNOWN"
			clientIP := "UNKNOWN"
			if c.Request != nil {
				reqMethod = c.Request.Method
				if c.Request.URL != nil {
					reqPath = c.Request.URL.Path
				}
				clientIP = c.ClientIP()
			}

			log.Printf("[HTTP ERROR %d %s] TraceID: %s | %s %s | IP: %s | User: %s | Workspace: %s | Code: %s | Message: %s",
				writer.status, statusText, traceID, reqMethod, reqPath, clientIP, userID, wsID, code, msg)
		}

		downstream.Header().Set("Content-Type", "application/json")
		downstream.WriteHeader(writer.status)
		if len(response) > 0 {
			_, _ = downstream.Write(response)
		}
	}
}

func RequestIDFromContext(c *gin.Context) string {
	return requestIDFromContext(c)
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
