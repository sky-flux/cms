package feed

import (
	"crypto/md5"
	"encoding/xml"
	"fmt"
	"net/http"
	"strconv"
	"time"

	"github.com/go-chi/chi/v5"
)

const (
	xmlHeader     = `<?xml version="1.0" encoding="UTF-8"?>`
	sitemapNS     = "http://www.sitemaps.org/schemas/sitemap/0.9"
	atomNS        = "http://www.w3.org/2005/Atom"
	feedGenerator = "Sky Flux CMS"
	maxFeedLimit  = 50
	feedCacheSecs = 3600
)

// Handler serves XML feed and sitemap endpoints via plain Chi.
type Handler struct {
	posts      FeedPostQuery
	categories FeedCategoryQuery
	tags       FeedTagQuery
	site       FeedSiteQuery
}

// NewHandler creates a feed handler with the required query ports.
func NewHandler(
	posts FeedPostQuery,
	categories FeedCategoryQuery,
	tags FeedTagQuery,
	site FeedSiteQuery,
) *Handler {
	return &Handler{posts: posts, categories: categories, tags: tags, site: site}
}

// RegisterRoutes wires feed endpoints onto a Chi router.
func RegisterRoutes(r chi.Router, h *Handler) {
	r.Get("/feed/rss", h.RSS)
	r.Get("/feed/atom", h.Atom)
	r.Get("/sitemap.xml", h.SitemapIndex)
	r.Get("/sitemap-posts.xml", h.SitemapPosts)
	r.Get("/sitemap-categories.xml", h.SitemapCategories)
	r.Get("/sitemap-tags.xml", h.SitemapTags)
}

// RSS serves GET /feed/rss — RSS 2.0 XML.
func (h *Handler) RSS(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, maxFeedLimit)
	ctx := r.Context()

	posts, err := h.posts.ListPublished(ctx, limit)
	if err != nil {
		http.Error(w, "feed generation error", http.StatusInternalServerError)
		return
	}
	siteURL := h.site.GetSiteURL(ctx)
	feedURL := siteURL + "/feed/rss"

	ch := RSSChannel{
		Title:         h.site.GetSiteTitle(ctx),
		Link:          siteURL,
		Description:   h.site.GetSiteDescription(ctx),
		Language:      h.site.GetSiteLanguage(ctx),
		LastBuildDate: time.Now().UTC().Format(time.RFC1123Z),
		Generator:     feedGenerator,
		AtomLink:      AtomLink{Href: feedURL, Rel: "self", Type: "application/rss+xml"},
	}
	for _, p := range posts {
		ch.Items = append(ch.Items, buildRSSItem(siteURL, p))
	}
	doc := RSSFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		DC:      "http://purl.org/dc/elements/1.1/",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Channel: ch,
	}
	writeXMLDoc(w, "application/rss+xml; charset=utf-8", doc)
}

// Atom serves GET /feed/atom — Atom 1.0 XML.
func (h *Handler) Atom(w http.ResponseWriter, r *http.Request) {
	limit := parseLimit(r, maxFeedLimit)
	ctx := r.Context()

	posts, err := h.posts.ListPublished(ctx, limit)
	if err != nil {
		http.Error(w, "feed generation error", http.StatusInternalServerError)
		return
	}
	siteURL := h.site.GetSiteURL(ctx)
	feedURL := siteURL + "/feed/atom"

	doc := AtomFeed{
		XMLNS:     atomNS,
		Title:     h.site.GetSiteTitle(ctx),
		ID:        siteURL,
		Updated:   time.Now().UTC().Format(time.RFC3339),
		Generator: feedGenerator,
		Link: []AtomFeedLink{
			{Href: siteURL, Rel: "alternate", Type: "text/html"},
			{Href: feedURL, Rel: "self", Type: "application/atom+xml"},
		},
	}
	for _, p := range posts {
		doc.Entries = append(doc.Entries, buildAtomEntry(siteURL, p))
	}
	writeXMLDoc(w, "application/atom+xml; charset=utf-8", doc)
}

// SitemapIndex serves GET /sitemap.xml — index referencing sub-sitemaps.
func (h *Handler) SitemapIndex(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	siteURL := h.site.GetSiteURL(ctx)
	now := time.Now().UTC().Format("2006-01-02")

	doc := SitemapIndex{
		XMLNS: sitemapNS,
		Sitemaps: []SitemapEntry{
			{Loc: siteURL + "/sitemap-posts.xml", Lastmod: now},
			{Loc: siteURL + "/sitemap-categories.xml", Lastmod: now},
			{Loc: siteURL + "/sitemap-tags.xml", Lastmod: now},
		},
	}
	writeXMLDoc(w, "application/xml; charset=utf-8", doc)
}

// SitemapPosts serves GET /sitemap-posts.xml.
func (h *Handler) SitemapPosts(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	posts, err := h.posts.ListPublished(ctx, 1000)
	if err != nil {
		http.Error(w, "sitemap generation error", http.StatusInternalServerError)
		return
	}
	siteURL := h.site.GetSiteURL(ctx)
	doc := URLSet{XMLNS: sitemapNS}
	for _, p := range posts {
		doc.URLs = append(doc.URLs, SitemapURL{
			Loc:        siteURL + "/posts/" + p.Slug,
			Lastmod:    p.UpdatedAt.UTC().Format("2006-01-02"),
			Changefreq: "weekly",
			Priority:   "0.8",
		})
	}
	writeXMLDoc(w, "application/xml; charset=utf-8", doc)
}

// SitemapCategories serves GET /sitemap-categories.xml.
func (h *Handler) SitemapCategories(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	cats, err := h.categories.ListAll(ctx)
	if err != nil {
		http.Error(w, "sitemap generation error", http.StatusInternalServerError)
		return
	}
	siteURL := h.site.GetSiteURL(ctx)
	doc := URLSet{XMLNS: sitemapNS}
	for _, c := range cats {
		u := SitemapURL{
			Loc:        siteURL + "/categories/" + c.Slug,
			Changefreq: "weekly",
			Priority:   "0.6",
		}
		if c.LastPostAt != nil {
			u.Lastmod = c.LastPostAt.UTC().Format("2006-01-02")
		}
		doc.URLs = append(doc.URLs, u)
	}
	writeXMLDoc(w, "application/xml; charset=utf-8", doc)
}

// SitemapTags serves GET /sitemap-tags.xml.
func (h *Handler) SitemapTags(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	tags, err := h.tags.ListWithPosts(ctx)
	if err != nil {
		http.Error(w, "sitemap generation error", http.StatusInternalServerError)
		return
	}
	siteURL := h.site.GetSiteURL(ctx)
	doc := URLSet{XMLNS: sitemapNS}
	for _, tag := range tags {
		u := SitemapURL{
			Loc:        siteURL + "/tags/" + tag.Slug,
			Changefreq: "weekly",
			Priority:   "0.5",
		}
		if tag.LastPostAt != nil {
			u.Lastmod = tag.LastPostAt.UTC().Format("2006-01-02")
		}
		doc.URLs = append(doc.URLs, u)
	}
	writeXMLDoc(w, "application/xml; charset=utf-8", doc)
}

// buildRSSItem converts a FeedPost into an RSSItem.
func buildRSSItem(siteURL string, p FeedPost) RSSItem {
	postURL := siteURL + "/posts/" + p.Slug
	return RSSItem{
		Title:          p.Title,
		Link:           postURL,
		GUID:           GUID{IsPermaLink: "true", Value: postURL},
		Description:    p.Excerpt,
		ContentEncoded: p.Content,
		Creator:        p.AuthorName,
		PubDate:        p.PublishedAt.UTC().Format(time.RFC1123Z),
		Categories:     p.Categories,
	}
}

// buildAtomEntry converts a FeedPost into an AtomEntry.
func buildAtomEntry(siteURL string, p FeedPost) AtomEntry {
	postURL := siteURL + "/posts/" + p.Slug
	entry := AtomEntry{
		Title:     p.Title,
		ID:        postURL,
		Updated:   p.UpdatedAt.UTC().Format(time.RFC3339),
		Published: p.PublishedAt.UTC().Format(time.RFC3339),
		Link:      AtomFeedLink{Href: postURL, Rel: "alternate", Type: "text/html"},
		Summary:   p.Excerpt,
	}
	if p.Content != "" {
		entry.Content = &AtomContent{Type: "html", Value: p.Content}
	}
	if p.AuthorName != "" {
		entry.Author = &AtomAuthor{Name: p.AuthorName}
	}
	return entry
}

// writeXMLDoc marshals v to XML and writes it with ETag + Cache-Control headers.
func writeXMLDoc(w http.ResponseWriter, contentType string, v any) {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		http.Error(w, "xml marshal error", http.StatusInternalServerError)
		return
	}
	out := []byte(xmlHeader + "\n")
	out = append(out, data...)
	etag := fmt.Sprintf(`"%x"`, md5.Sum(out))
	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", fmt.Sprintf("public, max-age=%d", feedCacheSecs))
	w.Header().Set("ETag", etag)
	w.WriteHeader(http.StatusOK)
	w.Write(out) //nolint:errcheck
}

// parseLimit reads ?limit= from the request, caps at maxLimit, defaults to 20.
func parseLimit(r *http.Request, maxLimit int) int {
	limit, _ := strconv.Atoi(r.URL.Query().Get("limit"))
	if limit <= 0 {
		limit = 20
	}
	if limit > maxLimit {
		limit = maxLimit
	}
	return limit
}
