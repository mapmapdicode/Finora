package middleware

import (
	"bytes"
	"crypto/sha256"
	"encoding/hex"
	"io"
	"net/http"
	"strings"
	"sync"

	"github.com/gin-gonic/gin"
	"wealthos-backend/internal/storage"
)

type idempotencyResponse struct {
	Status int
	Body   []byte
}

type idempotencyResponseWriter struct {
	gin.ResponseWriter
	status int
	body   *bytes.Buffer
}

func (w *idempotencyResponseWriter) Write(data []byte) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.Write(data)
	_, bodyErr := w.body.Write(data)
	if bodyErr != nil {
		return n, bodyErr
	}
	return n, err
}

func (w *idempotencyResponseWriter) WriteString(s string) (int, error) {
	if w.status == 0 {
		w.status = http.StatusOK
	}
	n, err := w.ResponseWriter.WriteString(s)
	_, bodyErr := w.body.WriteString(s)
	if bodyErr != nil {
		return n, bodyErr
	}
	return n, err
}

func (w *idempotencyResponseWriter) WriteHeader(code int) {
	w.status = code
	w.ResponseWriter.WriteHeader(code)
}

func (w *idempotencyResponseWriter) Status() int {
	if w.status == 0 {
		return http.StatusOK
	}
	return w.status
}

func (w *idempotencyResponseWriter) result() idempotencyResponse {
	body := make([]byte, len(w.body.Bytes()))
	copy(body, w.body.Bytes())
	return idempotencyResponse{
		Status: w.Status(),
		Body:   body,
	}
}

var idempotencyStore sync.Map
var idempotencyResponseStore sync.Map

func IdempotencyGuard(stores ...storage.Store) gin.HandlerFunc {
	var store storage.Store
	if len(stores) > 0 {
		store = stores[0]
	}

	return func(c *gin.Context) {
		var rawBody []byte
		var err error
		if c.Request.Body != nil {
			rawBody, err = io.ReadAll(c.Request.Body)
			if err != nil {
				c.JSON(http.StatusInternalServerError, gin.H{
					"code":    "IDEMPOTENCY_BODY_READ_ERROR",
					"message": "failed to read request body for idempotency",
				})
				c.Abort()
				return
			}
		}
		c.Request.Body = io.NopCloser(bytes.NewReader(rawBody))

		key := strings.TrimSpace(c.GetHeader("Idempotency-Key"))
		if key == "" {
			c.JSON(http.StatusBadRequest, gin.H{
				"code":    "MISSING_IDEMPOTENCY_KEY",
				"message": "Idempotency-Key is required",
			})
			c.Abort()
			return
		}

		userID := currentContextString(c, "user_id")
		sum := sha256.Sum256(rawBody)
		bodyHash := hex.EncodeToString(sum[:])
		marker := c.Request.Method + ":" + c.Request.URL.Path + ":" + userID + ":" + key + ":" + bodyHash

		if cached, ok := idempotencyResponseStore.Load(marker); ok {
			if value, ok := cached.(idempotencyResponse); ok {
				c.Data(value.Status, "application/json", value.Body)
				c.Abort()
				return
			}
		}

		duplicate := false
		if store != nil {
			if !store.RecordIdempotency(marker) {
				duplicate = true
			}
		} else if _, ok := idempotencyStore.LoadOrStore(marker, true); ok {
			duplicate = true
		}

		if duplicate {
			if cached, ok := idempotencyResponseStore.Load(marker); ok {
				if value, ok := cached.(idempotencyResponse); ok {
					c.Data(value.Status, "application/json", value.Body)
					c.Abort()
					return
				}
			}
			c.JSON(http.StatusConflict, gin.H{
				"code":    "DUPLICATE_IDEMPOTENCY_KEY",
				"message": "this idempotency key was already used",
			})
			c.Abort()
			return
		}

		writer := &idempotencyResponseWriter{
			ResponseWriter: c.Writer,
			body:           &bytes.Buffer{},
		}
		c.Writer = writer
		c.Next()
		response := writer.result()
		c.Writer = writer.ResponseWriter
		idempotencyResponseStore.Store(marker, response)
	}
}

func currentContextString(c *gin.Context, key string) string {
	if raw, ok := c.Get(key); ok {
		if value, ok := raw.(string); ok {
			value = strings.TrimSpace(value)
			if value != "" {
				return value
			}
		}
	}
	return "anonymous"
}
