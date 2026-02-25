package middleware

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
)

// APIKeyLookup abstracts API key validation.
type APIKeyLookup interface {
	GetByHash(ctx context.Context, hash string) (*model.APIKey, error)
}

// APIKey validates the X-API-Key header against sfc_site_api_keys.
func APIKey(lookup APIKeyLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		raw := c.GetHeader("X-API-Key")
		if raw == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing X-API-Key header",
			})
			return
		}

		h := sha256.Sum256([]byte(raw))
		hash := hex.EncodeToString(h[:])

		key, err := lookup.GetByHash(c.Request.Context(), hash)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid api key",
			})
			return
		}

		if key.Status != model.APIKeyStatusActive {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "api key revoked",
			})
			return
		}

		if key.ExpiresAt != nil && key.ExpiresAt.Before(time.Now()) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "api key expired",
			})
			return
		}

		c.Set("api_key_id", key.ID)
		c.Next()
	}
}
