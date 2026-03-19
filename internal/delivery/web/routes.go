package web

import "github.com/go-chi/chi/v5"

// RegisterRoutes mounts all public website routes onto the given router.
// IMPORTANT: Call this AFTER mounting /api and /console routes.
func (h *WebHandler) RegisterRoutes(r chi.Router) {
	// Static assets (app.css, htmx.min.js) served by the embed handler upstream.
	// These routes handle the HTML pages only.

	r.Get("/", h.Home)

	// Posts
	r.Get("/posts/{slug}", h.PostDetail)
	r.Get("/posts/partial", h.PostsPartial) // HTMX infinite scroll fragment

	// Category archive + HTMX pagination fragment
	r.Get("/categories/{slug}", h.CategoryArchive)
	r.Get("/categories/{slug}/partial", h.CategoryPartial)

	// Tag archive + HTMX pagination fragment
	r.Get("/tags/{slug}", h.TagArchive)
	r.Get("/tags/{slug}/partial", h.TagPartial)

	// Search (full page + HTMX partial via HX-Request header detection)
	r.Get("/search", h.Search)

	// Comment submission (HTMX form POST)
	r.Post("/posts/{slug}/comments", h.SubmitComment)

	// Custom pages / catch-all (must be LAST)
	r.Get("/{slug}", h.CustomPage)
}
