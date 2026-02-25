package feed

import (
	"context"
	"encoding/xml"
	"fmt"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

const generator = "Sky Flux CMS"

// Service generates RSS, Atom, and Sitemap XML feeds.
type Service struct {
	posts      FeedPostReader
	categories FeedCategoryReader
	tags       FeedTagReader
	site       SiteConfigReader
}

// NewService creates a new feed Service.
func NewService(posts FeedPostReader, categories FeedCategoryReader, tags FeedTagReader, site SiteConfigReader) *Service {
	return &Service{
		posts:      posts,
		categories: categories,
		tags:       tags,
		site:       site,
	}
}

// GenerateRSS produces an RSS 2.0 XML document.
func (s *Service) GenerateRSS(ctx context.Context, limit int, categorySlug, tagSlug string) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, limit, categorySlug, tagSlug)
	if err != nil {
		return nil, fmt.Errorf("feed: list published posts: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	feedURL := siteURL + "/feed/rss"

	var lastBuild string
	if len(posts) > 0 {
		lastBuild = formatPubDate(posts[0].PublishedAt)
	}

	feed := RSSFeed{
		Version: "2.0",
		Atom:    "http://www.w3.org/2005/Atom",
		DC:      "http://purl.org/dc/elements/1.1/",
		Content: "http://purl.org/rss/1.0/modules/content/",
		Channel: RSSChannel{
			Title:         s.site.GetSiteTitle(ctx),
			Link:          siteURL,
			Description:   s.site.GetSiteDescription(ctx),
			Language:      s.site.GetSiteLanguage(ctx),
			LastBuildDate: lastBuild,
			Generator:     generator,
			AtomLink: AtomLink{
				Href: feedURL,
				Rel:  "self",
				Type: "application/rss+xml",
			},
			Items: postsToRSSItems(posts, siteURL),
		},
	}

	return marshalXML(feed)
}

// GenerateAtom produces an Atom 1.0 XML document.
func (s *Service) GenerateAtom(ctx context.Context, limit int, categorySlug, tagSlug string) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, limit, categorySlug, tagSlug)
	if err != nil {
		return nil, fmt.Errorf("feed: list published posts: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	feedURL := siteURL + "/feed/atom"
	siteTitle := s.site.GetSiteTitle(ctx)

	var updated string
	if len(posts) > 0 {
		updated = formatRFC3339(posts[0].PublishedAt)
	} else {
		now := time.Now().UTC()
		updated = formatRFC3339(&now)
	}

	feed := AtomFeed{
		XMLNS:     "http://www.w3.org/2005/Atom",
		Title:     siteTitle,
		ID:        siteURL + "/",
		Updated:   updated,
		Generator: generator,
		Link: []AtomFeedLink{
			{Href: siteURL, Rel: "alternate", Type: "text/html"},
			{Href: feedURL, Rel: "self", Type: "application/atom+xml"},
		},
		Entries: postsToAtomEntries(posts, siteURL),
	}

	return marshalXML(feed)
}

// GenerateSitemapIndex produces a sitemap index linking to sub-sitemaps.
func (s *Service) GenerateSitemapIndex(ctx context.Context) ([]byte, error) {
	siteURL := s.site.GetSiteURL(ctx)
	now := time.Now().UTC().Format(time.RFC3339)

	idx := SitemapIndex{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		Sitemaps: []Sitemap{
			{Loc: siteURL + "/sitemap-posts.xml", Lastmod: now},
			{Loc: siteURL + "/sitemap-categories.xml", Lastmod: now},
			{Loc: siteURL + "/sitemap-tags.xml", Lastmod: now},
		},
	}

	return marshalXML(idx)
}

// GeneratePostsSitemap produces a URL set of all published posts.
func (s *Service) GeneratePostsSitemap(ctx context.Context) ([]byte, error) {
	posts, err := s.posts.ListPublished(ctx, 0, "", "")
	if err != nil {
		return nil, fmt.Errorf("feed: list posts for sitemap: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)
	now := time.Now().UTC()

	urls := make([]URL, 0, len(posts))
	for _, p := range posts {
		priority, changefreq := postPriority(p, now)
		urls = append(urls, URL{
			Loc:        siteURL + "/" + p.Slug,
			Lastmod:    formatLastmod(&p.UpdatedAt),
			Changefreq: changefreq,
			Priority:   priority,
		})
	}

	set := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	return marshalXML(set)
}

// GenerateCategoriesSitemap produces a URL set of all categories.
func (s *Service) GenerateCategoriesSitemap(ctx context.Context) ([]byte, error) {
	cats, err := s.categories.ListAll(ctx)
	if err != nil {
		return nil, fmt.Errorf("feed: list categories for sitemap: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)

	urls := make([]URL, 0, len(cats))
	for _, c := range cats {
		lastmod, _ := s.categories.LatestPostDate(ctx, c.ID)

		priority := "0.6"
		if c.ParentID != nil {
			priority = "0.5"
		}

		u := URL{
			Loc:        siteURL + "/category/" + c.Slug,
			Changefreq: "weekly",
			Priority:   priority,
		}
		if lastmod != nil {
			u.Lastmod = formatLastmod(lastmod)
		}
		urls = append(urls, u)
	}

	set := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	return marshalXML(set)
}

// GenerateTagsSitemap produces a URL set of tags that have posts.
func (s *Service) GenerateTagsSitemap(ctx context.Context) ([]byte, error) {
	tags, err := s.tags.ListWithPosts(ctx)
	if err != nil {
		return nil, fmt.Errorf("feed: list tags for sitemap: %w", err)
	}

	siteURL := s.site.GetSiteURL(ctx)

	urls := make([]URL, 0, len(tags))
	for _, t := range tags {
		u := URL{
			Loc:        siteURL + "/tag/" + t.Slug,
			Changefreq: "weekly",
			Priority:   "0.4",
		}
		if t.LastPostDate != nil {
			u.Lastmod = formatLastmod(t.LastPostDate)
		}
		urls = append(urls, u)
	}

	set := URLSet{
		XMLNS: "http://www.sitemaps.org/schemas/sitemap/0.9",
		URLs:  urls,
	}

	return marshalXML(set)
}

// --- helpers ---

func postsToRSSItems(posts []model.Post, siteURL string) []RSSItem {
	items := make([]RSSItem, 0, len(posts))
	for _, p := range posts {
		link := siteURL + "/" + p.Slug

		creator := ""
		if p.Author != nil {
			creator = p.Author.DisplayName
		}

		items = append(items, RSSItem{
			Title:          p.Title,
			Link:           link,
			GUID:           GUID{IsPermaLink: "true", Value: link},
			Description:    p.Excerpt,
			ContentEncoded: p.Content,
			Creator:        creator,
			PubDate:        formatPubDate(p.PublishedAt),
		})
	}
	return items
}

func postsToAtomEntries(posts []model.Post, siteURL string) []AtomEntry {
	entries := make([]AtomEntry, 0, len(posts))
	for _, p := range posts {
		link := siteURL + "/" + p.Slug

		var author *AtomAuthor
		if p.Author != nil {
			author = &AtomAuthor{Name: p.Author.DisplayName}
		}

		entry := AtomEntry{
			Title:     p.Title,
			Link:      AtomFeedLink{Href: link, Rel: "alternate", Type: "text/html"},
			ID:        link,
			Updated:   formatRFC3339(&p.UpdatedAt),
			Published: formatRFC3339(p.PublishedAt),
			Author:    author,
			Summary:   p.Excerpt,
		}
		if p.Content != "" {
			entry.Content = &AtomContent{
				Type:  "html",
				Value: p.Content,
			}
		}
		entries = append(entries, entry)
	}
	return entries
}

// postPriority returns sitemap priority and changefreq based on post age and type.
func postPriority(p model.Post, now time.Time) (priority, changefreq string) {
	if p.PostType == "page" {
		return "0.6", "monthly"
	}

	if p.PublishedAt == nil {
		return "0.5", "monthly"
	}

	age := now.Sub(*p.PublishedAt)
	switch {
	case age <= 7*24*time.Hour:
		return "0.9", "daily"
	case age <= 30*24*time.Hour:
		return "0.8", "weekly"
	case age <= 90*24*time.Hour:
		return "0.7", "weekly"
	default:
		return "0.5", "monthly"
	}
}

func formatPubDate(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC1123Z)
}

func formatRFC3339(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func formatLastmod(t *time.Time) string {
	if t == nil {
		return ""
	}
	return t.UTC().Format(time.RFC3339)
}

func marshalXML(v any) ([]byte, error) {
	data, err := xml.MarshalIndent(v, "", "  ")
	if err != nil {
		return nil, fmt.Errorf("feed: marshal xml: %w", err)
	}
	return append([]byte(xml.Header), data...), nil
}
