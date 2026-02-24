package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestAuditContext(t *testing.T) {
	gin.SetMode(gin.TestMode)
	w := httptest.NewRecorder()
	_, r := gin.CreateTestContext(w)

	var capturedIP, capturedUA string
	r.Use(AuditContext())
	r.GET("/test", func(c *gin.Context) {
		v, _ := c.Get("audit_ip")
		capturedIP = v.(string)
		v2, _ := c.Get("audit_ua")
		capturedUA = v2.(string)
		c.Status(200)
	})

	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("User-Agent", "TestAgent/1.0")
	req.RemoteAddr = "192.168.1.1:1234"
	r.ServeHTTP(w, req)

	assert.Equal(t, 200, w.Code)
	assert.Equal(t, "192.168.1.1", capturedIP)
	assert.Equal(t, "TestAgent/1.0", capturedUA)
}
