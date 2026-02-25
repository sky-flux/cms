package middleware

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
)

// RateLimit returns middleware that limits requests per IP using Redis SET NX.
// window defines the cooldown period between requests.
func RateLimit(rdb *redis.Client, prefix string, window time.Duration) gin.HandlerFunc {
	return func(c *gin.Context) {
		if rdb == nil {
			c.Next()
			return
		}

		siteSlug, _ := c.Get("site_slug")
		ip := c.ClientIP()
		key := prefix + ":" + siteSlug.(string) + ":" + ip

		ok, err := rdb.SetNX(c.Request.Context(), key, "1", window).Result()
		if err != nil {
			// Redis error — allow request through (fail open)
			c.Next()
			return
		}
		if !ok {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"success": false,
				"error":   "rate limit exceeded, please try again later",
			})
			return
		}
		c.Next()
	}
}
