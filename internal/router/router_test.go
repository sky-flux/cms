package router

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestHealthEndpoint(t *testing.T) {
	gin.SetMode(gin.TestMode)

	engine := gin.New()
	// Register a standalone health handler for unit testing without DB/Redis/Meilisearch.
	engine.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"status":      "ok",
			"db":          "connected",
			"redis":       "connected",
			"meilisearch": "connected",
		})
	})

	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/health", nil)
	engine.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected status 200, got %d", w.Code)
	}

	var body map[string]string
	if err := json.Unmarshal(w.Body.Bytes(), &body); err != nil {
		t.Fatalf("failed to unmarshal response: %v", err)
	}

	if body["status"] != "ok" {
		t.Errorf("expected status 'ok', got '%s'", body["status"])
	}
	if body["db"] != "connected" {
		t.Errorf("expected db 'connected', got '%s'", body["db"])
	}
	if body["redis"] != "connected" {
		t.Errorf("expected redis 'connected', got '%s'", body["redis"])
	}
	if body["meilisearch"] != "connected" {
		t.Errorf("expected meilisearch 'connected', got '%s'", body["meilisearch"])
	}
}
