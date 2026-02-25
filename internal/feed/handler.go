package feed

import (
	"crypto/md5"
	"fmt"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

// Handler serves XML feed and sitemap endpoints.
type Handler struct {
	svc *Service
}

// NewHandler creates a feed handler.
func NewHandler(svc *Service) *Handler {
	return &Handler{svc: svc}
}

// RSSFeed serves RSS 2.0 XML.
func (h *Handler) RSSFeed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}
	data, err := h.svc.GenerateRSS(c.Request.Context(), limit, c.Query("category"), c.Query("tag"))
	if err != nil {
		c.String(http.StatusInternalServerError, "feed generation error")
		return
	}
	writeXML(c, "application/rss+xml; charset=utf-8", data, 3600)
}

// AtomFeed serves Atom 1.0 XML.
func (h *Handler) AtomFeed(c *gin.Context) {
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "20"))
	if limit > 50 {
		limit = 50
	}
	data, err := h.svc.GenerateAtom(c.Request.Context(), limit, c.Query("category"), c.Query("tag"))
	if err != nil {
		c.String(http.StatusInternalServerError, "feed generation error")
		return
	}
	writeXML(c, "application/atom+xml; charset=utf-8", data, 3600)
}

// SitemapIndex serves the sitemap index XML.
func (h *Handler) SitemapIndex(c *gin.Context) {
	data, err := h.svc.GenerateSitemapIndex(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

// SitemapPosts serves the posts sitemap XML.
func (h *Handler) SitemapPosts(c *gin.Context) {
	data, err := h.svc.GeneratePostsSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

// SitemapCategories serves the categories sitemap XML.
func (h *Handler) SitemapCategories(c *gin.Context) {
	data, err := h.svc.GenerateCategoriesSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

// SitemapTags serves the tags sitemap XML.
func (h *Handler) SitemapTags(c *gin.Context) {
	data, err := h.svc.GenerateTagsSitemap(c.Request.Context())
	if err != nil {
		c.String(http.StatusInternalServerError, "sitemap generation error")
		return
	}
	writeXML(c, "application/xml; charset=utf-8", data, 3600)
}

func writeXML(c *gin.Context, contentType string, data []byte, maxAge int) {
	etag := fmt.Sprintf(`"%x"`, md5.Sum(data))
	c.Header("Cache-Control", fmt.Sprintf("public, max-age=%d", maxAge))
	c.Header("ETag", etag)
	c.Data(http.StatusOK, contentType, data)
}
