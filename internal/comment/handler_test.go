package comment

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/stretchr/testify/assert"
)

func setupRouter() (*gin.Engine, *mockRepo) {
	gin.SetMode(gin.TestMode)
	repo := &mockRepo{comments: make(map[string]*model.Comment)}
	a := &mockAudit{}
	mailer := &mail.NoopSender{}
	svc := NewService(repo, a, mailer)
	h := NewHandler(svc)

	r := gin.New()
	g := r.Group("/comments")
	g.GET("", h.List)
	g.PUT("/batch-status", h.BatchStatus)
	g.GET("/:id", h.Get)
	g.PUT("/:id/status", h.UpdateStatus)
	g.PUT("/:id/pin", h.TogglePin)
	g.POST("/:id/reply", func(c *gin.Context) {
		c.Set("user_id", "uid-1")
		c.Set("user_name", "Admin")
		c.Set("user_email", "admin@test.com")
		h.Reply(c)
	})
	g.DELETE("/:id", h.Delete)

	return r, repo
}

func TestHandler_List(t *testing.T) {
	r, repo := setupRouter()
	repo.listRows = []CommentRow{}
	repo.listTotal = 0

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments?page=1&per_page=10", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Get(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "a@b.com"}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("GET", "/comments/c1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_UpdateStatus(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1"}

	body, _ := json.Marshal(UpdateStatusReq{Status: "approved"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/c1/status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_BatchStatus(t *testing.T) {
	r, repo := setupRouter()
	repo.batchAffected = 2

	body, _ := json.Marshal(BatchStatusReq{IDs: []string{"c1", "c2"}, Status: "spam"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("PUT", "/comments/batch-status", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_Reply(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1", PostID: "p1", AuthorEmail: "guest@test.com"}

	body, _ := json.Marshal(ReplyReq{Content: "thank you"})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/comments/c1/reply", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)
}

func TestHandler_Delete(t *testing.T) {
	r, repo := setupRouter()
	repo.comments["c1"] = &model.Comment{ID: "c1"}

	w := httptest.NewRecorder()
	req, _ := http.NewRequest("DELETE", "/comments/c1", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}
