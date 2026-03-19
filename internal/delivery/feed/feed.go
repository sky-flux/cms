// Package feed provides XML feed and sitemap types for the Delivery BC.
package feed

import "encoding/xml"

// --- RSS 2.0 ---

// RSSFeed is the root RSS 2.0 document.
type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	DC      string     `xml:"xmlns:dc,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Channel RSSChannel `xml:"channel"`
}

// RSSChannel holds feed metadata and items.
type RSSChannel struct {
	Title         string    `xml:"title"`
	Link          string    `xml:"link"`
	Description   string    `xml:"description"`
	Language      string    `xml:"language"`
	LastBuildDate string    `xml:"lastBuildDate"`
	Generator     string    `xml:"generator"`
	AtomLink      AtomLink  `xml:"atom:link"`
	Items         []RSSItem `xml:"item"`
}

// AtomLink is the atom:link self-referencing element in RSS.
type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

// RSSItem represents a single post entry in RSS.
type RSSItem struct {
	Title          string   `xml:"title"`
	Link           string   `xml:"link"`
	GUID           GUID     `xml:"guid"`
	Description    string   `xml:"description"`
	ContentEncoded string   `xml:"content:encoded"`
	Creator        string   `xml:"dc:creator"`
	PubDate        string   `xml:"pubDate"`
	Categories     []string `xml:"category"`
}

// GUID is the globally unique identifier for an RSS item.
type GUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

// --- Atom 1.0 ---

// AtomFeed is the root Atom 1.0 document.
type AtomFeed struct {
	XMLName   xml.Name       `xml:"feed"`
	XMLNS     string         `xml:"xmlns,attr"`
	Title     string         `xml:"title"`
	Link      []AtomFeedLink `xml:"link"`
	Updated   string         `xml:"updated"`
	ID        string         `xml:"id"`
	Author    *AtomAuthor    `xml:"author,omitempty"`
	Generator string         `xml:"generator"`
	Entries   []AtomEntry    `xml:"entry"`
}

// AtomFeedLink is a link element within an Atom feed or entry.
type AtomFeedLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

// AtomAuthor holds the author's name for Atom entries.
type AtomAuthor struct {
	Name string `xml:"name"`
}

// AtomEntry represents a single post entry in Atom.
type AtomEntry struct {
	Title     string       `xml:"title"`
	Link      AtomFeedLink `xml:"link"`
	ID        string       `xml:"id"`
	Updated   string       `xml:"updated"`
	Published string       `xml:"published"`
	Author    *AtomAuthor  `xml:"author,omitempty"`
	Summary   string       `xml:"summary,omitempty"`
	Content   *AtomContent `xml:"content,omitempty"`
}

// AtomContent wraps the full post content in CDATA.
type AtomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",cdata"`
}

// --- Sitemap ---

// SitemapIndex is the root sitemap index document.
type SitemapIndex struct {
	XMLName  xml.Name       `xml:"sitemapindex"`
	XMLNS    string         `xml:"xmlns,attr"`
	Sitemaps []SitemapEntry `xml:"sitemap"`
}

// SitemapEntry is a single sitemap reference in an index.
type SitemapEntry struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

// URLSet is the root sitemap URL set document.
type URLSet struct {
	XMLName xml.Name     `xml:"urlset"`
	XMLNS   string       `xml:"xmlns,attr"`
	URLs    []SitemapURL `xml:"url"`
}

// SitemapURL is a single URL entry in a sitemap.
type SitemapURL struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}
