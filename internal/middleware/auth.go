package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/pkg/jwt"
)

func Auth(jwtMgr *jwt.Manager) gin.HandlerFunc {
	return func(c *gin.Context) {
		// Try Authorization header first
		header := c.GetHeader("Authorization")
		if header != "" && strings.HasPrefix(header, "Bearer ") {
			tokenStr := strings.TrimPrefix(header, "Bearer ")
			if claims, err := jwtMgr.Verify(tokenStr); err == nil {
				if blacklisted, _ := jwtMgr.IsBlacklisted(c.Request.Context(), claims.JTI); !blacklisted {
					c.Set("user_id", claims.Subject)
					c.Set("token_jti", claims.JTI)
					if claims.Purpose != "" {
						c.Set("token_purpose", claims.Purpose)
					}
					c.Next()
					return
				}
			}
		}

		// Try access_token cookie
		if accessToken, err := c.Cookie("access_token"); err == nil {
			if claims, err := jwtMgr.Verify(accessToken); err == nil {
				if blacklisted, _ := jwtMgr.IsBlacklisted(c.Request.Context(), claims.JTI); !blacklisted {
					c.Set("user_id", claims.Subject)
					c.Set("token_jti", claims.JTI)
					c.Next()
					return
				}
			}
		}

		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"success": false,
			"error":   "missing or invalid token",
		})
	}
}
