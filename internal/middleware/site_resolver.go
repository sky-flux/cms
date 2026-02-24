package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// SiteLookup abstracts site resolution by slug or domain.
type SiteLookup interface {
	GetIDBySlug(ctx context.Context, slug string) (string, error)
	GetSlugByDomain(ctx context.Context, domain string) (string, string, error) // slug, id, error
}

// SiteResolver reads the site slug from X-Site-Slug header (preferred)
// or falls back to domain-based lookup via the Host header.
// It stores "site_slug" and "site_id" in the gin context for downstream handlers.
func SiteResolver(lookup SiteLookup) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug := c.GetHeader("X-Site-Slug")

		if slug == "" {
			host := c.Request.Host
			if host != "" {
				s, id, err := lookup.GetSlugByDomain(c.Request.Context(), host)
				if err == nil && s != "" {
					c.Set("site_slug", s)
					c.Set("site_id", id)
					c.Next()
					return
				}
			}
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "missing X-Site-Slug header or valid domain",
			})
			return
		}

		id, err := lookup.GetIDBySlug(c.Request.Context(), slug)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
				"success": false,
				"error":   "site not found",
			})
			return
		}

		c.Set("site_slug", slug)
		c.Set("site_id", id)
		c.Next()
	}
}
