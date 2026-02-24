package search

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/meilisearch/meilisearch-go"
)

// Client wraps meilisearch.ServiceManager with simplified interfaces.
type Client struct {
	ms meilisearch.ServiceManager
}

// NewClient creates a Meilisearch client wrapper.
// Pass nil for ms to create a no-op client (graceful degradation).
func NewClient(ms meilisearch.ServiceManager) *Client {
	return &Client{ms: ms}
}

// Available returns true if the Meilisearch connection is live.
func (c *Client) Available() bool {
	return c.ms != nil && c.ms.IsHealthy()
}

// IndexSettings configures a Meilisearch index.
type IndexSettings struct {
	SearchableAttributes []string
	DisplayedAttributes  []string
	FilterableAttributes []string
	SortableAttributes   []string
}

// EnsureIndex creates an index if it doesn't exist and applies settings.
func (c *Client) EnsureIndex(ctx context.Context, uid string, settings *IndexSettings) error {
	if c.ms == nil {
		return nil
	}

	_, err := c.ms.CreateIndex(&meilisearch.IndexConfig{Uid: uid, PrimaryKey: "id"})
	if err != nil {
		// Index may already exist — not an error.
	}

	if settings != nil {
		s := &meilisearch.Settings{}
		if len(settings.SearchableAttributes) > 0 {
			s.SearchableAttributes = settings.SearchableAttributes
		}
		if len(settings.DisplayedAttributes) > 0 {
			s.DisplayedAttributes = settings.DisplayedAttributes
		}
		if len(settings.FilterableAttributes) > 0 {
			s.FilterableAttributes = settings.FilterableAttributes
		}
		if len(settings.SortableAttributes) > 0 {
			s.SortableAttributes = settings.SortableAttributes
		}
		if _, err := c.ms.Index(uid).UpdateSettings(s); err != nil {
			return fmt.Errorf("update index settings %s: %w", uid, err)
		}
	}
	return nil
}

// UpsertDocuments adds or updates documents in an index.
func (c *Client) UpsertDocuments(ctx context.Context, uid string, docs any) error {
	if c.ms == nil {
		return nil
	}
	pk := "id"
	_, err := c.ms.Index(uid).AddDocuments(docs, &meilisearch.DocumentOptions{PrimaryKey: &pk})
	if err != nil {
		return fmt.Errorf("upsert documents %s: %w", uid, err)
	}
	return nil
}

// DeleteDocuments removes documents by ID from an index.
func (c *Client) DeleteDocuments(ctx context.Context, uid string, ids []string) error {
	if c.ms == nil {
		return nil
	}
	_, err := c.ms.Index(uid).DeleteDocuments(ids, nil)
	if err != nil {
		return fmt.Errorf("delete documents %s: %w", uid, err)
	}
	return nil
}

// SearchOpts configures a search query.
type SearchOpts struct {
	Limit  int64
	Offset int64
	Filter string
}

// SearchResult contains search results.
type SearchResult struct {
	Hits           []map[string]any
	EstimatedTotal int64
}

// Search queries an index.
func (c *Client) Search(ctx context.Context, uid, query string, opts *SearchOpts) (*SearchResult, error) {
	if c.ms == nil {
		return &SearchResult{}, nil
	}

	req := &meilisearch.SearchRequest{Query: query}
	if opts != nil {
		if opts.Limit > 0 {
			req.Limit = opts.Limit
		}
		if opts.Offset > 0 {
			req.Offset = opts.Offset
		}
		if opts.Filter != "" {
			req.Filter = opts.Filter
		}
	}

	resp, err := c.ms.Index(uid).Search(query, req)
	if err != nil {
		return nil, fmt.Errorf("search %s: %w", uid, err)
	}

	hits := make([]map[string]any, len(resp.Hits))
	for i, h := range resp.Hits {
		m := make(map[string]any, len(h))
		for k, raw := range h {
			var v any
			if err := json.Unmarshal(raw, &v); err == nil {
				m[k] = v
			}
		}
		hits[i] = m
	}

	return &SearchResult{
		Hits:           hits,
		EstimatedTotal: resp.EstimatedTotalHits,
	}, nil
}

// DeleteIndex removes an entire index (used when deleting a site).
func (c *Client) DeleteIndex(ctx context.Context, uid string) error {
	if c.ms == nil {
		return nil
	}
	_, err := c.ms.DeleteIndex(uid)
	if err != nil {
		return fmt.Errorf("delete index %s: %w", uid, err)
	}
	return nil
}
