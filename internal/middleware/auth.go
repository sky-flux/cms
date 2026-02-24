package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/jwt"
)

func Auth(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		header := c.GetHeader("Authorization")
		if header == "" || !strings.HasPrefix(header, "Bearer ") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "missing or invalid authorization header",
			})
			return
		}
		tokenStr := strings.TrimPrefix(header, "Bearer ")
		claims, err := jwtMgr.Verify(tokenStr)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "invalid or expired token",
			})
			return
		}
		blacklisted, err := jwtMgr.IsBlacklisted(c.Request.Context(), claims.JTI)
		if err != nil || blacklisted {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"success": false,
				"error":   "token has been revoked",
			})
			return
		}
		c.Set("user_id", claims.Subject)
		c.Set("token_jti", claims.JTI)
		if claims.Purpose != "" {
			c.Set("token_purpose", claims.Purpose)
		}
		c.Next()
	}
}
