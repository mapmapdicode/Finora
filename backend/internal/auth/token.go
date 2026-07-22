package auth

import (
	"crypto/hmac"
	"crypto/sha256"
	"crypto/subtle"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

const (
	jwtHeaderJSON = `{"alg":"HS256","typ":"JWT"}`
)

type tokenClaims struct {
	Subject   string `json:"sub"`
	IssuedAt  int64  `json:"iat"`
	ExpiresAt int64  `json:"exp"`
}

func IssueToken(secret, userID string, ttl time.Duration) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("jwt secret is required")
	}
	userID = strings.TrimSpace(userID)
	if userID == "" {
		return "", errors.New("userID is required")
	}

	if ttl <= 0 {
		ttl = 24 * time.Hour
	}

	now := time.Now().UTC().Unix()
	claims := tokenClaims{
		Subject:   userID,
		IssuedAt:  now,
		ExpiresAt: now + int64(ttl.Seconds()),
	}
	payloadJSON, err := json.Marshal(claims)
	if err != nil {
		return "", fmt.Errorf("marshal jwt claims: %w", err)
	}

	header := base64.RawURLEncoding.EncodeToString([]byte(jwtHeaderJSON))
	payload := base64.RawURLEncoding.EncodeToString(payloadJSON)
	signingInput := header + "." + payload
	signature := signToken(secret, signingInput)
	return signingInput + "." + signature, nil
}

func ParseToken(secret, token string) (string, error) {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return "", errors.New("jwt secret is required")
	}
	token = strings.TrimSpace(token)
	if token == "" {
		return "", errors.New("missing token")
	}

	parts := strings.Split(token, ".")
	if len(parts) != 3 {
		return "", errors.New("invalid token format")
	}
	signingInput := parts[0] + "." + parts[1]
	expectedSig := signToken(secret, signingInput)
	if subtle.ConstantTimeCompare([]byte(parts[2]), []byte(expectedSig)) != 1 {
		return "", errors.New("invalid signature")
	}

	payloadJSON, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return "", errors.New("invalid token payload")
	}

	var claims tokenClaims
	if err := json.Unmarshal(payloadJSON, &claims); err != nil {
		return "", errors.New("invalid token payload")
	}
	if strings.TrimSpace(claims.Subject) == "" {
		return "", errors.New("missing subject")
	}
	if claims.ExpiresAt <= 0 || claims.ExpiresAt < time.Now().UTC().Unix() {
		return "", errors.New("token expired")
	}

	return claims.Subject, nil
}

func signToken(secret, signingInput string) string {
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	return base64.RawURLEncoding.EncodeToString(mac.Sum(nil))
}
