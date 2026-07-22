package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestLoginRateLimitRejectsAfterMaxRequests(t *testing.T) {
	clearRateLimitStore()

	r := gin.New()
	hitCount := 0
	r.POST("/auth/login", LoginRateLimit(2), func(c *gin.Context) {
		hitCount++
		c.Status(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)
	if got, want := resp1.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("first request expected %d, got %d", want, got)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)
	if got, want := resp2.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("second request expected %d, got %d", want, got)
	}

	req3 := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req3.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp3 := httptest.NewRecorder()
	r.ServeHTTP(resp3, req3)
	if got, want := resp3.Result().StatusCode, http.StatusTooManyRequests; got != want {
		t.Fatalf("third request expected %d, got %d", want, got)
	}

	if got := resp3.Result().Header.Get("Retry-After"); got == "" {
		t.Fatalf("expected Retry-After header")
	}

	if hitCount != 2 {
		t.Fatalf("expected handler run twice, got %d", hitCount)
	}
}

func TestLoginRateLimitIsPerClientIP(t *testing.T) {
	clearRateLimitStore()

	r := gin.New()
	hitCount := 0
	r.POST("/auth/login", LoginRateLimit(1), func(c *gin.Context) {
		hitCount++
		c.Status(http.StatusOK)
	})

	req1 := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req1.Header.Set("X-Forwarded-For", "203.0.113.10")
	resp1 := httptest.NewRecorder()
	r.ServeHTTP(resp1, req1)
	if resp1.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected first request 200, got %d", resp1.Result().StatusCode)
	}

	req2 := httptest.NewRequest(http.MethodPost, "/auth/login", nil)
	req2.Header.Set("X-Forwarded-For", "203.0.113.20")
	resp2 := httptest.NewRecorder()
	r.ServeHTTP(resp2, req2)
	if resp2.Result().StatusCode != http.StatusOK {
		t.Fatalf("expected different client request 200, got %d", resp2.Result().StatusCode)
	}

	if hitCount != 2 {
		t.Fatalf("expected two clients to be handled independently, got %d", hitCount)
	}
}
