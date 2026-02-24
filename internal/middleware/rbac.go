package middleware

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

// PermissionChecker abstracts the RBAC permission check.
// Implemented by rbac.Service.
type PermissionChecker interface {
	CheckPermission(ctx context.Context, userID, method, path string) (bool, error)
}

// RBAC returns middleware that checks API-level permissions.
// It reads user_id from Gin context (set by JWT auth middleware),
// then delegates to PermissionChecker.CheckPermission().
func RBAC(checker PermissionChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		userID := c.GetString("user_id")
		if userID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "unauthorized",
			})
			return
		}

		method := c.Request.Method
		path := c.FullPath() // Gin route pattern, e.g. /api/v1/posts/:id

		allowed, err := checker.CheckPermission(c.Request.Context(), userID, method, path)
		if err != nil {
			slog.Error("rbac check failed",
				"error", err,
				"user_id", userID,
				"method", method,
				"path", path,
			)
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"success": false,
				"error":   "permission check failed",
			})
			return
		}

		if !allowed {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"success": false,
				"error":   "forbidden",
			})
			return
		}

		c.Next()
	}
}
