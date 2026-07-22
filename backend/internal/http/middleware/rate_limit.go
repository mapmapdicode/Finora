package middleware

import (
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type rateLimitBucket struct {
	windowStart time.Time
	count       int
}

var rateLimitStore sync.Map
var nowFunc = time.Now

func LoginRateLimit(maxPerMinute int) gin.HandlerFunc {
	if maxPerMinute <= 0 {
		maxPerMinute = 60
	}
	return rateLimitMiddleware(
		"auth/login",
		maxPerMinute,
		time.Minute,
		func(c *gin.Context) string {
			remoteIP := strings.TrimSpace(c.ClientIP())
			if remoteIP == "" {
				remoteIP = "anonymous"
			}
			return remoteIP
		},
	)
}

func rateLimitMiddleware(scope string, maxRequests int, window time.Duration, keyFn func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := strings.TrimSpace(keyFn(c))
		if key == "" {
			key = "anonymous"
		}
		marker := scope + ":" + key
		now := nowFunc().UTC()

		var (
			bucket rateLimitBucket
			ok     bool
		)

		raw, exists := rateLimitStore.Load(marker)
		if exists {
			bucket, ok = raw.(rateLimitBucket)
			if ok {
				if now.Sub(bucket.windowStart) > window {
					bucket = rateLimitBucket{
						windowStart: now,
						count:       0,
					}
				}
			} else {
				bucket = rateLimitBucket{
					windowStart: now,
					count:       0,
				}
			}
		} else {
			bucket = rateLimitBucket{
				windowStart: now,
				count:       0,
			}
		}

		resetAt := bucket.windowStart.Add(window)
		remainingBefore := maxRequests - bucket.count
		if remainingBefore < 0 {
			remainingBefore = 0
		}
		c.Header("X-RateLimit-Limit", fmt.Sprintf("%d", maxRequests))
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", remainingBefore))
		c.Header("X-RateLimit-Reset", fmt.Sprintf("%d", resetAt.Unix()))

		allow := bucket.count < maxRequests
		if !allow {
			retryAfter := int(time.Until(resetAt).Seconds())
			if retryAfter < 1 {
				retryAfter = 1
			}
			c.Header("Retry-After", fmt.Sprintf("%d", retryAfter))
			c.JSON(http.StatusTooManyRequests, gin.H{
				"code":    "RATE_LIMIT_EXCEEDED",
				"message": "too many requests, please try again later",
				"retryIn": retryAfter,
			})
			c.Abort()
			return
		}

		bucket.count++
		rateLimitStore.Store(marker, bucket)
		c.Header("X-RateLimit-Remaining", fmt.Sprintf("%d", maxRequests-bucket.count))
		c.Next()
	}
}

func clearRateLimitStore() {
	rateLimitStore = sync.Map{}
}
