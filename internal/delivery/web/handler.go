package web

import (
	"log/slog"
	"net/http"
	"strconv"

	"github.com/go-chi/chi/v5"
	"github.com/sky-flux/cms/web/templates"
)

const defaultPageSize = 10

// WebHandler handles HTTP requests for public Templ SSR pages.
type WebHandler struct {
	postQuery     PostQuery
	categoryQuery CategoryQuery
	tagQuery      TagQuery
	commentWriter CommentWriter
	siteConfig    SiteConfigLoader
	log           *slog.Logger
}

// NewWebHandler constructs a WebHandler with all required dependencies.
func NewWebHandler(
	postQuery PostQuery,
	categoryQuery CategoryQuery,
	tagQuery TagQuery,
	commentWriter CommentWriter,
	siteConfig SiteConfigLoader,
	log *slog.Logger,
) *WebHandler {
	return &WebHandler{
		postQuery:     postQuery,
		categoryQuery: categoryQuery,
		tagQuery:      tagQuery,
		commentWriter: commentWriter,
		siteConfig:    siteConfig,
		log:           log,
	}
}

// Home renders the homepage with the latest posts.
// GET /
func (h *WebHandler) Home(w http.ResponseWriter, r *http.Request) {
	cfg, err := h.siteConfig.Load(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	posts, err := h.postQuery.ListLatest(r.Context(), 1, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.HomePage(cfg, posts).Render(r.Context(), w)
}

// PostDetail renders a single post page.
// GET /posts/:slug
func (h *WebHandler) PostDetail(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	cfg, err := h.siteConfig.Load(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	post, err := h.postQuery.GetBySlug(r.Context(), slug)
	if err != nil {
		h.notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.PostPage(cfg, *post).Render(r.Context(), w)
}

// PostsPartial returns only the post card fragment for HTMX infinite scroll.
// GET /posts/partial?page=N
func (h *WebHandler) PostsPartial(w http.ResponseWriter, r *http.Request) {
	page := parsePageParam(r)
	posts, err := h.postQuery.ListLatest(r.Context(), page, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	nextURL := ""
	if len(posts) >= defaultPageSize {
		nextURL = "/posts/partial?page=" + strconv.Itoa(page+1)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
}

// CategoryArchive renders the category archive page.
// GET /categories/:slug
func (h *WebHandler) CategoryArchive(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page := parsePageParam(r)
	cfg, err := h.siteConfig.Load(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	name, err := h.categoryQuery.GetBySlug(r.Context(), slug)
	if err != nil {
		h.notFound(w, r)
		return
	}
	posts, err := h.postQuery.ListByCategory(r.Context(), slug, page, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.CategoryArchivePage(cfg, name, slug, posts, page).Render(r.Context(), w)
}

// CategoryPartial returns HTMX post card fragment for category pagination.
// GET /categories/:slug/partial?page=N
func (h *WebHandler) CategoryPartial(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page := parsePageParam(r)
	posts, err := h.postQuery.ListByCategory(r.Context(), slug, page, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	nextURL := ""
	if len(posts) >= defaultPageSize {
		nextURL = "/categories/" + slug + "/partial?page=" + strconv.Itoa(page+1)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
}

// TagArchive renders the tag archive page.
// GET /tags/:slug
func (h *WebHandler) TagArchive(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page := parsePageParam(r)
	cfg, err := h.siteConfig.Load(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	name, err := h.tagQuery.GetBySlug(r.Context(), slug)
	if err != nil {
		h.notFound(w, r)
		return
	}
	posts, err := h.postQuery.ListByTag(r.Context(), slug, page, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.TagArchivePage(cfg, name, slug, posts, page).Render(r.Context(), w)
}

// TagPartial returns HTMX post card fragment for tag pagination.
// GET /tags/:slug/partial?page=N
func (h *WebHandler) TagPartial(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	page := parsePageParam(r)
	posts, err := h.postQuery.ListByTag(r.Context(), slug, page, defaultPageSize)
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	nextURL := ""
	if len(posts) >= defaultPageSize {
		nextURL = "/tags/" + slug + "/partial?page=" + strconv.Itoa(page+1)
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.PostsPartial(posts, page+1, nextURL).Render(r.Context(), w)
}

// Search renders the search results page or partial.
// GET /search?q=
// When HX-Request header is present, returns SearchResults partial only.
func (h *WebHandler) Search(w http.ResponseWriter, r *http.Request) {
	query := r.URL.Query().Get("q")
	cfg, _ := h.siteConfig.Load(r.Context())

	var results []templates.PostSummary
	if query != "" {
		var err error
		results, err = h.postQuery.Search(r.Context(), query, 20)
		if err != nil {
			h.log.ErrorContext(r.Context(), "search failed", "query", query, "err", err)
			results = nil
		}
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")

	// HTMX partial: return only the results fragment
	if r.Header.Get("HX-Request") == "true" {
		templates.SearchResults(query, results).Render(r.Context(), w)
		return
	}

	templates.SearchPage(cfg, query, results).Render(r.Context(), w)
}

// SubmitComment handles HTMX comment form submission.
// POST /posts/:slug/comments
// Returns an HTML fragment (success or error message) for #comment-form-status.
func (h *WebHandler) SubmitComment(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	if err := r.ParseForm(); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(`<span class="text-red-600">Invalid form submission.</span>`))
		return
	}
	name := r.FormValue("author_name")
	email := r.FormValue("author_email")
	body := r.FormValue("body")

	if name == "" || email == "" || body == "" {
		w.WriteHeader(http.StatusUnprocessableEntity)
		w.Write([]byte(`<span class="text-red-600">All fields are required.</span>`))
		return
	}

	if err := h.commentWriter.Submit(r.Context(), slug, name, email, body); err != nil {
		h.log.ErrorContext(r.Context(), "comment submit failed", "slug", slug, "err", err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`<span class="text-red-600">Failed to submit comment. Please try again.</span>`))
		return
	}

	w.WriteHeader(http.StatusCreated)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	w.Write([]byte(`<span class="text-green-600">Thank you! Your comment is awaiting moderation.</span>`))
}

// CustomPage handles /:slug for Page-type posts (e.g. /about).
// GET /:slug
func (h *WebHandler) CustomPage(w http.ResponseWriter, r *http.Request) {
	slug := chi.URLParam(r, "slug")
	cfg, err := h.siteConfig.Load(r.Context())
	if err != nil {
		h.serverError(w, r, err)
		return
	}
	post, err := h.postQuery.GetBySlug(r.Context(), slug)
	if err != nil {
		h.notFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	templates.PostPage(cfg, *post).Render(r.Context(), w)
}

// serverError writes a plain 500 response. In production, render a Templ 500 page instead.
func (h *WebHandler) serverError(w http.ResponseWriter, r *http.Request, err error) {
	h.log.ErrorContext(r.Context(), "internal server error", "err", err)
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
}

// notFound writes a plain 404 response. In production, render a Templ 404 page instead.
func (h *WebHandler) notFound(w http.ResponseWriter, r *http.Request) {
	http.Error(w, "Not Found", http.StatusNotFound)
}

// parsePageParam extracts the ?page= query param, defaulting to 1.
func parsePageParam(r *http.Request) int {
	p, err := strconv.Atoi(r.URL.Query().Get("page"))
	if err != nil || p < 1 {
		return 1
	}
	return p
}
