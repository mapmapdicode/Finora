package middleware

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"wealthos-backend/internal/auth"

	"github.com/gin-gonic/gin"
)

func clearIdempotencyStore() {
	idempotencyStore = sync.Map{}
	idempotencyResponseStore = sync.Map{}
}

func TestIdempotencyGuardRequiresKey(t *testing.T) {
	clearIdempotencyStore()
	r := gin.New()
	r.POST("/transactions", IdempotencyGuard(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodPost, "/transactions", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
}

func TestIdempotencyGuardRejectsDuplicateKeys(t *testing.T) {
	clearIdempotencyStore()
	r := gin.New()
	hitCount := 0
	r.POST("/transactions", IdempotencyGuard(), func(c *gin.Context) {
		hitCount++
		c.JSON(http.StatusOK, gin.H{"ok": true, "idempotent": false})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/transactions", nil)
	req1.Header.Set("Idempotency-Key", "dup-key")
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/transactions", nil)
	req2.Header.Set("Idempotency-Key", "dup-key")
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)

	if got, want := resp1.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("first request expected %d, got %d", want, got)
	}
	if got, want := resp2.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("second request expected %d, got %d", want, got)
	}
	if got, want := resp1.Body.String(), resp2.Body.String(); got != want {
		t.Fatalf("replayed response body should match first request: %q != %q", want, got)
	}
	if hitCount != 1 {
		t.Fatalf("expected handler to run once, got %d", hitCount)
	}
}

func TestIdempotencyGuardAllowsDifferentBodies(t *testing.T) {
	clearIdempotencyStore()
	r := gin.New()
	hitCount := 0
	r.POST("/transactions", IdempotencyGuard(), func(c *gin.Context) {
		hitCount++
		c.JSON(http.StatusOK, gin.H{"hits": hitCount})
	})

	req1 := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"amount": 100}`))
	req1.Header.Set("Idempotency-Key", "same-body")
	req1.Header.Set("Content-Type", "application/json")
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)

	req2 := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"amount": 200}`))
	req2.Header.Set("Idempotency-Key", "same-body")
	req2.Header.Set("Content-Type", "application/json")
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)

	if resp1.Result().StatusCode != http.StatusOK {
		t.Fatalf("first request expected %d, got %d", http.StatusOK, resp1.Result().StatusCode)
	}
	if resp2.Result().StatusCode != http.StatusOK {
		t.Fatalf("second request expected %d, got %d", http.StatusOK, resp2.Result().StatusCode)
	}
	if hitCount != 2 {
		t.Fatalf("expected handler to run twice, got %d", hitCount)
	}
}

func TestIdempotencyGuardPassesBodyToHandler(t *testing.T) {
	clearIdempotencyStore()
	r := gin.New()
	r.POST("/transactions", IdempotencyGuard(), func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"read": "failed"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"body": string(body)})
	})

	req := httptest.NewRequest(http.MethodPost, "/transactions", strings.NewReader(`{"note":"hello"}`))
	req.Header.Set("Idempotency-Key", "body-read")
	req.Header.Set("Content-Type", "application/json")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	var got map[string]any
	if err := json.NewDecoder(resp.Result().Body).Decode(&got); err != nil {
		t.Fatalf("decode response: %v", err)
	}
	if got["body"] != `{"note":"hello"}` {
		t.Fatalf("expected handler to read original body, got: %v", got["body"])
	}
}

func TestUserContextMiddlewareAcceptsJWT(t *testing.T) {
	r := gin.New()
	r.POST("/x", UserContextMiddleware("", "unit-test-secret"), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": currentUserID(c)})
	})

	token, err := auth.IssueToken("unit-test-secret", "user-jwt", time.Minute)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var got map[string]any
	if err := json.NewDecoder(resp.Result().Body).Decode(&got); err != nil {
		t.Fatalf("decode: %v", err)
	}
	if got["user"] != "user-jwt" {
		t.Fatalf("expected user-jwt, got %v", got["user"])
	}
}

func TestUserContextMiddlewareAcceptsLegacyTokens(t *testing.T) {
	r := gin.New()
	r.POST("/x", UserContextMiddleware("legacy-static", ""), func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"user": currentUserID(c)})
	})

	req := httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("Authorization", "token-user")
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for legacy token, got %d", resp.Result().StatusCode)
	}

	req = httptest.NewRequest(http.MethodPost, "/x", nil)
	req.Header.Set("X-Auth-Token", "legacy-static")
	resp = httptest.NewRecorder()
	r.ServeHTTP(resp, req)
	if resp.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected 200 for static token header, got %d", resp.Result().StatusCode)
	}
}
