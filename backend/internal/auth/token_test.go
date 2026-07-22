package auth

import (
	"encoding/base64"
	"encoding/json"
	"testing"
	"time"
)

func TestIssueAndParseTokenRoundTrip(t *testing.T) {
	token, err := IssueToken("test-secret", "user-123", 2*time.Minute)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	subject, err := ParseToken("test-secret", token)
	if err != nil {
		t.Fatalf("parse token: %v", err)
	}
	if subject != "user-123" {
		t.Fatalf("expected user-123, got %q", subject)
	}
}

func TestParseTokenRejectsInvalidSignature(t *testing.T) {
	token, err := IssueToken("test-secret", "user-123", 2*time.Minute)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	token = token + "x"
	if _, err := ParseToken("test-secret", token); err == nil {
		t.Fatal("expected error for tampered token")
	}
}

func TestParseTokenRejectsExpiredToken(t *testing.T) {
	header := base64.RawURLEncoding.EncodeToString([]byte(jwtHeaderJSON))
	claims := tokenClaims{
		Subject:   "user-123",
		IssuedAt:  time.Now().UTC().Add(-5 * time.Minute).Unix(),
		ExpiresAt: time.Now().UTC().Add(-1 * time.Minute).Unix(),
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		t.Fatalf("marshal expired claims: %v", err)
	}
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	token := header + "." + payload + "." + signToken("test-secret", header+"."+payload)
	if _, err := ParseToken("test-secret", token); err == nil {
		t.Fatal("expected expired token error")
	}
}

func TestParseTokenRejectsWrongSecret(t *testing.T) {
	token, err := IssueToken("test-secret", "user-123", 2*time.Minute)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}
	if _, err := ParseToken("wrong-secret", token); err == nil {
		t.Fatal("expected error for wrong secret")
	}
}
