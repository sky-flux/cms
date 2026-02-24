package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func TestRequestID_Generated(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	var ctxID string
	r.GET("/test", func(c *gin.Context) {
		ctxID = c.GetString("request_id")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.NotEmpty(t, ctxID, "should generate request_id")
	assert.Equal(t, ctxID, w.Header().Get("X-Request-ID"))
}

func TestRequestID_Preserved(t *testing.T) {
	r := gin.New()
	r.Use(RequestID())
	var ctxID string
	r.GET("/test", func(c *gin.Context) {
		ctxID = c.GetString("request_id")
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-Request-ID", "my-custom-id-123")
	r.ServeHTTP(w, req)

	assert.Equal(t, "my-custom-id-123", ctxID)
	assert.Equal(t, "my-custom-id-123", w.Header().Get("X-Request-ID"))
}
