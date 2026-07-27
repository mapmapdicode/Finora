package middleware

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestErrorEnvelopeKeepsSuccessResponse(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorEnvelope())
	r.GET("/ok", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})

	req := httptest.NewRequest(http.MethodGet, "/ok", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusOK; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var got map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &got); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if got["status"] != "ok" {
		t.Fatalf("expected status payload, got %#v", got)
	}
	if _, ok := got["traceId"]; ok {
		t.Fatalf("expected no traceId for success response")
	}
}

func TestErrorEnvelopeInjectsTraceIdForCodeErrors(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorEnvelope())
	r.GET("/bad", func(c *gin.Context) {
		c.Set("request_id", "trace-xyz")
		c.JSON(http.StatusBadRequest, gin.H{"code": "BAD_REQUEST", "message": "bad input"})
	})

	req := httptest.NewRequest(http.MethodGet, "/bad", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	var gotPayload map[string]any
	if err := json.Unmarshal(resp.Body.Bytes(), &gotPayload); err != nil {
		t.Fatalf("invalid response body: %v", err)
	}
	if gotPayload["code"] != "BAD_REQUEST" {
		t.Fatalf("expected code field, got %#v", gotPayload)
	}
	if gotPayload["traceId"] != "trace-xyz" {
		t.Fatalf("expected traceId trace-xyz, got %v", gotPayload["traceId"])
	}
}

func TestErrorEnvelopeSkipsTraceForNonCodeErrorPayloads(t *testing.T) {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(ErrorEnvelope())
	r.GET("/text", func(c *gin.Context) {
		c.Set("request_id", "trace-text")
		c.String(http.StatusBadRequest, "plain error")
	})

	req := httptest.NewRequest(http.MethodGet, "/text", nil)
	resp := httptest.NewRecorder()
	r.ServeHTTP(resp, req)

	if got, want := resp.Result().StatusCode, http.StatusBadRequest; got != want {
		t.Fatalf("expected status %d, got %d", want, got)
	}
	if resp.Body.String() != "plain error" {
		t.Fatalf("expected original plain body, got: %q", resp.Body.String())
	}
}

func TestErrorEnvelopeLogsStatusCodes(t *testing.T) {
	gin.SetMode(gin.TestMode)
	statuses := []int{http.StatusBadRequest, http.StatusUnauthorized, http.StatusForbidden, http.StatusInternalServerError}

	for _, st := range statuses {
		r := gin.New()
		r.Use(ErrorEnvelope())
		r.GET("/err", func(c *gin.Context) {
			c.Set("user_id", "user-123")
			c.Set("request_id", "trace-test")
			c.JSON(st, gin.H{"code": "ERROR_CODE", "message": "error occurred"})
		})

		req := httptest.NewRequest(http.MethodGet, "/err", nil)
		resp := httptest.NewRecorder()
		r.ServeHTTP(resp, req)

		if got := resp.Result().StatusCode; got != st {
			t.Fatalf("status %d mismatch, got %d", st, got)
		}
	}
}
