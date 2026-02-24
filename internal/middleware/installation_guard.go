package middleware

import (
	"context"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
)

type InstallChecker interface {
	IsInstalled(ctx context.Context) bool
	MarkInstalled()
}

func InstallationGuard(checker InstallChecker, exemptPrefixes ...string) gin.HandlerFunc {
	return func(c *gin.Context) {
		for _, prefix := range exemptPrefixes {
			if strings.HasPrefix(c.Request.URL.Path, prefix) {
				c.Next()
				return
			}
		}
		if checker.IsInstalled(c.Request.Context()) {
			c.Next()
			return
		}
		c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
			"success": false,
			"error": gin.H{
				"code":    "NOT_INSTALLED",
				"message": "CMS is not installed. Please complete setup first.",
			},
		})
	}
}
