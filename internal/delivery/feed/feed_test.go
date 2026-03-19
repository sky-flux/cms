package feed_test

import (
	"encoding/xml"
	"strings"
	"testing"

	"github.com/sky-flux/cms/internal/delivery/feed"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRSSItem_MarshalXML(t *testing.T) {
	item := feed.RSSItem{
		Title:   "Hello World",
		Link:    "https://example.com/posts/hello",
		PubDate: "Mon, 01 Jan 2026 00:00:00 +0000",
		GUID: feed.GUID{
			IsPermaLink: "true",
			Value:       "https://example.com/posts/hello",
		},
	}
	data, err := xml.Marshal(item)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "Hello World"))
	assert.True(t, strings.Contains(string(data), "https://example.com/posts/hello"))
}

func TestAtomEntry_MarshalXML(t *testing.T) {
	entry := feed.AtomEntry{
		Title: "Test Post",
		ID:    "urn:uuid:abc-123",
		Link: feed.AtomFeedLink{
			Href: "https://example.com/posts/test",
			Rel:  "alternate",
		},
		Updated: "2026-01-01T00:00:00Z",
	}
	data, err := xml.Marshal(entry)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "Test Post"))
}

func TestSitemapURL_MarshalXML(t *testing.T) {
	u := feed.SitemapURL{
		Loc:        "https://example.com/posts/hello",
		Lastmod:    "2026-01-01",
		Changefreq: "weekly",
		Priority:   "0.8",
	}
	data, err := xml.Marshal(u)
	require.NoError(t, err)
	assert.True(t, strings.Contains(string(data), "https://example.com/posts/hello"))
	assert.True(t, strings.Contains(string(data), "weekly"))
}
