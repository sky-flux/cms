package feed

import "encoding/xml"

// --- RSS 2.0 ---

type RSSFeed struct {
	XMLName xml.Name   `xml:"rss"`
	Version string     `xml:"version,attr"`
	Atom    string     `xml:"xmlns:atom,attr"`
	DC      string     `xml:"xmlns:dc,attr"`
	Content string     `xml:"xmlns:content,attr"`
	Channel RSSChannel `xml:"channel"`
}

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

type AtomLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr"`
	Type string `xml:"type,attr"`
}

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

type GUID struct {
	IsPermaLink string `xml:"isPermaLink,attr"`
	Value       string `xml:",chardata"`
}

// --- Atom 1.0 ---

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

type AtomFeedLink struct {
	Href string `xml:"href,attr"`
	Rel  string `xml:"rel,attr,omitempty"`
	Type string `xml:"type,attr,omitempty"`
}

type AtomAuthor struct {
	Name string `xml:"name"`
}

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

type AtomContent struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",cdata"`
}

// --- Sitemap ---

type SitemapIndex struct {
	XMLName  xml.Name  `xml:"sitemapindex"`
	XMLNS    string    `xml:"xmlns,attr"`
	Sitemaps []Sitemap `xml:"sitemap"`
}

type Sitemap struct {
	Loc     string `xml:"loc"`
	Lastmod string `xml:"lastmod,omitempty"`
}

type URLSet struct {
	XMLName xml.Name `xml:"urlset"`
	XMLNS   string   `xml:"xmlns,attr"`
	URLs    []URL    `xml:"url"`
}

type URL struct {
	Loc        string `xml:"loc"`
	Lastmod    string `xml:"lastmod,omitempty"`
	Changefreq string `xml:"changefreq,omitempty"`
	Priority   string `xml:"priority,omitempty"`
}
