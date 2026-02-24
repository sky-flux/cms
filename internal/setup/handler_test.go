package setup_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/setup"
	"github.com/stretchr/testify/assert"
)

func TestInitializeReq_Validation(t *testing.T) {
	gin.SetMode(gin.TestMode)

	tests := []struct {
		name string
		body map[string]string
		code int
	}{
		{
			name: "missing required fields",
			body: map[string]string{},
			code: 422,
		},
		{
			name: "invalid email",
			body: map[string]string{
				"site_name": "Blog", "site_slug": "blog", "site_url": "https://blog.com",
				"admin_email": "not-an-email", "admin_password": "Pass123!", "admin_display_name": "Admin",
			},
			code: 422,
		},
		{
			name: "password too short",
			body: map[string]string{
				"site_name": "Blog", "site_slug": "blog", "site_url": "https://blog.com",
				"admin_email": "a@b.com", "admin_password": "short", "admin_display_name": "Admin",
			},
			code: 422,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			r := gin.New()
			r.POST("/api/v1/setup/initialize", func(c *gin.Context) {
				var req setup.InitializeReq
				if err := c.ShouldBindJSON(&req); err != nil {
					c.JSON(422, gin.H{"success": false, "error": err.Error()})
					return
				}
				c.JSON(200, gin.H{"success": true})
			})
			body, _ := json.Marshal(tt.body)
			w := httptest.NewRecorder()
			req, _ := http.NewRequest("POST", "/api/v1/setup/initialize", bytes.NewReader(body))
			req.Header.Set("Content-Type", "application/json")
			r.ServeHTTP(w, req)
			assert.Equal(t, tt.code, w.Code)
		})
	}
}
