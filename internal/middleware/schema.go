package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/schema"
	"github.com/uptrace/bun"
)

func Schema(db *bun.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		slug, exists := c.Get("site_slug")
		if !exists {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "site context not set",
			})
			return
		}

		slugStr, ok := slug.(string)
		if !ok || slugStr == "" {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "site context not set",
			})
			return
		}

		if !schema.ValidateSlug(slugStr) {
			c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{
				"success": false,
				"error":   "invalid site slug format",
			})
			return
		}

		// Convert hyphen to underscore for schema name (PostgreSQL schema names don't support hyphens)
		schemaName := "site_" + strings.ReplaceAll(slugStr, "-", "_")
		_, err := db.ExecContext(c.Request.Context(),
			fmt.Sprintf("SET search_path TO '%s', 'public'", schemaName))
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "failed to set schema context",
			})
			return
		}

		c.Next()
	}
}
