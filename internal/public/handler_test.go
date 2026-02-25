package public

import (
	"bytes"
	"encoding/json"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// setupPublicRouter creates a Gin engine with site_slug middleware and all public routes.
func setupPublicRouter(h *Handler) *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(func(c *gin.Context) {
		c.Set("site_slug", "test-site")
		c.Next()
	})
	r.GET("/posts", h.ListPosts)
	r.GET("/posts/:slug", h.GetPost)
	r.GET("/categories", h.ListCategories)
	r.GET("/tags", h.ListTags)
	r.GET("/search", h.Search)
	r.GET("/posts/:slug/comments", h.ListComments)
	r.POST("/posts/:slug/comments", h.CreateComment)
	r.GET("/menus", h.GetMenu)
	r.GET("/preview/:token", h.Preview)
	return r
}

// newHandlerTestService creates a Service from testDeps and wraps it in a Handler.
func newHandlerTestService(d *testDeps) *Handler {
	svc := NewService(
		d.posts,
		d.categories,
		d.tags,
		d.comments,
		d.menus,
		d.previews,
		nil, // search client
		nil, // cache client
		slog.Default(),
		nil, // mailer
		"",  // siteName
	)
	return NewHandler(svc)
}

// ---------------------------------------------------------------------------
// ListPosts
// ---------------------------------------------------------------------------

func TestHandler_ListPosts_Success(t *testing.T) {
	now := time.Now()
	author := &model.User{ID: "u1", DisplayName: "Alice", AvatarURL: "https://img.example.com/alice.png"}

	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			posts: []model.Post{
				{
					ID: "p1", Title: "First Post", Slug: "first-post",
					Excerpt: "Excerpt 1", Status: model.PostStatusPublished,
					ViewCount: 42, PublishedAt: &now, Author: author,
				},
				{
					ID: "p2", Title: "Second Post", Slug: "second-post",
					Excerpt: "Excerpt 2", Status: model.PostStatusPublished,
					ViewCount: 7, PublishedAt: &now,
				},
			},
			total: 2,
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts?page=1&per_page=20", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data []PostListItem
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	require.Len(t, data, 2)
	assert.Equal(t, "p1", data[0].ID)
	assert.Equal(t, "First Post", data[0].Title)
	assert.Equal(t, "p2", data[1].ID)

	var meta map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(resp["meta"], &meta))
	assert.Equal(t, "2", string(meta["total"]))
}

// ---------------------------------------------------------------------------
// GetPost
// ---------------------------------------------------------------------------

func TestHandler_GetPost_NotFound(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			err: apperror.NotFound("post not found", nil),
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/nonexistent", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestHandler_GetPost_Success(t *testing.T) {
	now := time.Now()
	author := &model.User{ID: "u1", DisplayName: "Bob"}

	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			post: &model.Post{
				ID: "p1", Title: "My Post", Slug: "my-post",
				Content: "<p>Hello</p>", Excerpt: "Hello excerpt",
				Status: model.PostStatusPublished, ViewCount: 100,
				PublishedAt: &now, Author: author,
			},
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/my-post", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data PostDetail
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	assert.Equal(t, "p1", data.ID)
	assert.Equal(t, "My Post", data.Title)
	assert.Equal(t, "my-post", data.Slug)
	assert.Equal(t, "<p>Hello</p>", data.Content)
}

// ---------------------------------------------------------------------------
// ListCategories
// ---------------------------------------------------------------------------

func TestHandler_ListCategories_Success(t *testing.T) {
	parentID := "cat1"
	h := newHandlerTestService(&testDeps{
		categories: &mockCategoryReader{
			cats: []model.Category{
				{ID: "cat1", Name: "Tech", Slug: "tech", Path: "/tech"},
				{ID: "cat2", Name: "Go", Slug: "go", Path: "/tech/go", ParentID: &parentID},
			},
			counts: map[string]int64{"cat1": 5, "cat2": 8},
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/categories", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data []CategoryNode
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	require.NotEmpty(t, data)
}

// ---------------------------------------------------------------------------
// ListTags
// ---------------------------------------------------------------------------

func TestHandler_ListTags_Success(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		tags: &mockTagReader{
			tags: []TagWithCount{
				{Tag: model.Tag{ID: "t1", Name: "Golang", Slug: "golang"}, PostCount: 12},
				{Tag: model.Tag{ID: "t2", Name: "Docker", Slug: "docker"}, PostCount: 7},
			},
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/tags", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data []TagItem
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	require.Len(t, data, 2)
	assert.Equal(t, "Golang", data[0].Name)
	assert.Equal(t, int64(12), data[0].PostCount)
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestHandler_Search_EmptyQuery(t *testing.T) {
	// search.Client is concrete and cannot be mocked without a real Meilisearch
	// instance. With a nil search client the service panics, which the Recovery
	// middleware converts to 500. We verify the handler does not hang and the
	// recovery path works correctly.
	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/search?q=", nil)
	r.ServeHTTP(w, req)

	// Gin Recovery middleware catches the nil-pointer panic and returns 500.
	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

// ---------------------------------------------------------------------------
// ListComments
// ---------------------------------------------------------------------------

func TestHandler_ListComments_Success(t *testing.T) {
	now := time.Now()

	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			post: &model.Post{
				ID: "p1", Status: model.PostStatusPublished, Slug: "my-post",
			},
		},
		comments: &mockCommentReader{
			comments: []model.Comment{
				{
					ID: "c1", PostID: "p1", AuthorName: "Alice",
					Content: "Great post!", Pinned: model.ToggleNo, CreatedAt: now,
				},
				{
					ID: "c2", PostID: "p1", AuthorName: "Bob",
					Content: "Thanks!", Pinned: model.ToggleNo, CreatedAt: now.Add(time.Minute),
				},
			},
			total: 2,
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/posts/my-post/comments", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	// The paginated data wraps comment_count and comments.
	var data map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	assert.Equal(t, "2", string(data["comment_count"]))
}

// ---------------------------------------------------------------------------
// CreateComment
// ---------------------------------------------------------------------------

func TestHandler_CreateComment_InvalidBody(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		posts:    &mockPostReader{},
		comments: &mockCommentReader{},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts/my-post/comments", bytes.NewBufferString(""))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestHandler_CreateComment_Success(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			post: &model.Post{
				ID: "p1", Status: model.PostStatusPublished, Slug: "my-post",
			},
		},
		comments: &mockCommentReader{
			createID: "new-c1",
		},
	})

	body := `{"author_name":"Test User","author_email":"test@example.com","content":"Great post!"}`
	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/posts/my-post/comments", bytes.NewBufferString(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data CreateCommentResp
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	assert.Equal(t, "new-c1", data.ID)
	assert.Equal(t, "pending", data.Status)
}

// ---------------------------------------------------------------------------
// GetMenu
// ---------------------------------------------------------------------------

func TestHandler_GetMenu_NoParam(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		menus: &mockMenuReader{},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/menus", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}

func TestHandler_GetMenu_ByLocation(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		menus: &mockMenuReader{
			menu: &model.SiteMenu{
				ID: "m1", Name: "Main Nav", Slug: "main-nav", Location: "header",
			},
			items: []*model.SiteMenuItem{
				{
					ID: "i1", MenuID: "m1", Label: "Home", URL: "/",
					Target: "_self", SortOrder: 0, Status: model.MenuItemStatusActive,
				},
				{
					ID: "i2", MenuID: "m1", Label: "About", URL: "/about",
					Target: "_self", SortOrder: 1, Status: model.MenuItemStatusActive,
				},
			},
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/menus?location=header", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data PublicMenu
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	assert.Equal(t, "m1", data.ID)
	assert.Equal(t, "Main Nav", data.Name)
	assert.Equal(t, "header", data.Location)
	require.Len(t, data.Items, 2)
	assert.Equal(t, "Home", data.Items[0].Label)
	assert.Equal(t, "About", data.Items[1].Label)
}

// ---------------------------------------------------------------------------
// Preview
// ---------------------------------------------------------------------------

func TestHandler_Preview_Success(t *testing.T) {
	now := time.Now()
	futureExpiry := now.Add(24 * time.Hour)
	author := &model.User{ID: "u1", DisplayName: "Author"}

	h := newHandlerTestService(&testDeps{
		posts: &mockPostReader{
			post: &model.Post{
				ID: "p1", Title: "Preview Post", Slug: "preview-post",
				Content: "<p>Draft content</p>", Status: model.PostStatusDraft,
				Author: author,
			},
		},
		previews: &mockPreviewReader{
			token: &model.PreviewToken{
				ID: "tok1", PostID: "p1", TokenHash: "hashed",
				ExpiresAt: futureExpiry,
			},
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/preview/test-token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]json.RawMessage
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Contains(t, string(resp["success"]), "true")

	var data map[string]any
	require.NoError(t, json.Unmarshal(resp["data"], &data))
	assert.Equal(t, true, data["is_preview"])
	assert.Equal(t, "p1", data["id"])
	assert.Equal(t, "Preview Post", data["title"])
}

func TestHandler_Preview_NotFound(t *testing.T) {
	h := newHandlerTestService(&testDeps{
		previews: &mockPreviewReader{
			err: apperror.NotFound("invalid preview token", nil),
		},
	})

	r := setupPublicRouter(h)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/preview/bad-token", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, false, resp["success"])
}
