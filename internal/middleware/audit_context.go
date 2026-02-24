package middleware

import "github.com/gin-gonic/gin"

// AuditContext extracts client IP and User-Agent from the request
// and stores them in the gin context for the audit service to read.
func AuditContext() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Set("audit_ip", c.ClientIP())
		c.Set("audit_ua", c.GetHeader("User-Agent"))
		c.Next()
	}
}
