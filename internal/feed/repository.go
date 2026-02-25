package feed

import (
	"context"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/model"
)

// --- PostRepoAdapter ---

// PostRepoAdapter implements FeedPostReader using bun.
type PostRepoAdapter struct {
	db *bun.DB
}

// NewPostRepoAdapter creates a new PostRepoAdapter.
func NewPostRepoAdapter(db *bun.DB) *PostRepoAdapter {
	return &PostRepoAdapter{db: db}
}

// ListPublished returns published posts ordered by published_at DESC.
// limit=0 means no limit. categorySlug and tagSlug filter by join tables when non-empty.
func (r *PostRepoAdapter) ListPublished(ctx context.Context, limit int, categorySlug, tagSlug string) ([]model.Post, error) {
	var posts []model.Post

	q := r.db.NewSelect().
		Model(&posts).
		Relation("Author").
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL").
		OrderExpr("p.published_at DESC")

	if categorySlug != "" {
		q = q.Where(`p.id IN (
			SELECT pcm.post_id FROM sfc_site_post_category_map AS pcm
			JOIN sfc_site_categories AS cat ON cat.id = pcm.category_id
			WHERE cat.slug = ?
		)`, categorySlug)
	}

	if tagSlug != "" {
		q = q.Where(`p.id IN (
			SELECT ptm.post_id FROM sfc_site_post_tag_map AS ptm
			JOIN sfc_site_tags AS tag ON tag.id = ptm.tag_id
			WHERE tag.slug = ?
		)`, tagSlug)
	}

	if limit > 0 {
		q = q.Limit(limit)
	}

	if err := q.Scan(ctx); err != nil {
		return nil, err
	}
	return posts, nil
}

// LatestPublishedAt returns the most recent published_at timestamp.
func (r *PostRepoAdapter) LatestPublishedAt(ctx context.Context) (*time.Time, error) {
	var t time.Time
	err := r.db.NewSelect().
		TableExpr("sfc_site_posts AS p").
		ColumnExpr("MAX(p.published_at)").
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL").
		Scan(ctx, &t)
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		return nil, nil
	}
	return &t, nil
}

// --- CategoryRepoAdapter ---

// CategoryRepoAdapter implements FeedCategoryReader using bun.
type CategoryRepoAdapter struct {
	db *bun.DB
}

// NewCategoryRepoAdapter creates a new CategoryRepoAdapter.
func NewCategoryRepoAdapter(db *bun.DB) *CategoryRepoAdapter {
	return &CategoryRepoAdapter{db: db}
}

// ListAll returns all categories ordered by sort_order, name.
func (r *CategoryRepoAdapter) ListAll(ctx context.Context) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.NewSelect().
		Model(&cats).
		OrderExpr("c.sort_order ASC, c.name ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return cats, nil
}

// LatestPostDate returns the most recent published_at for posts in a category.
func (r *CategoryRepoAdapter) LatestPostDate(ctx context.Context, categoryID string) (*time.Time, error) {
	var t time.Time
	err := r.db.NewSelect().
		TableExpr("sfc_site_post_category_map AS pcm").
		Join("JOIN sfc_site_posts AS p ON p.id = pcm.post_id").
		ColumnExpr("MAX(p.published_at)").
		Where("pcm.category_id = ?", categoryID).
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL").
		Scan(ctx, &t)
	if err != nil {
		return nil, err
	}
	if t.IsZero() {
		return nil, nil
	}
	return &t, nil
}

// --- TagRepoAdapter ---

// TagRepoAdapter implements FeedTagReader using bun.
type TagRepoAdapter struct {
	db *bun.DB
}

// NewTagRepoAdapter creates a new TagRepoAdapter.
func NewTagRepoAdapter(db *bun.DB) *TagRepoAdapter {
	return &TagRepoAdapter{db: db}
}

// ListWithPosts returns tags that have at least one published post,
// along with each tag's latest post date and post count.
func (r *TagRepoAdapter) ListWithPosts(ctx context.Context) ([]TagWithLastmod, error) {
	var tags []TagWithLastmod
	err := r.db.NewSelect().
		TableExpr("sfc_site_tags AS t").
		ColumnExpr("t.*").
		ColumnExpr("MAX(p.published_at) AS last_post_date").
		ColumnExpr("COUNT(p.id) AS post_count").
		Join("LEFT JOIN sfc_site_post_tag_map AS ptm ON ptm.tag_id = t.id").
		Join("LEFT JOIN sfc_site_posts AS p ON p.id = ptm.post_id AND p.status = ? AND p.deleted_at IS NULL", model.PostStatusPublished).
		GroupExpr("t.id").
		Having("COUNT(p.id) > 0").
		OrderExpr("t.name ASC").
		Scan(ctx, &tags)
	if err != nil {
		return nil, err
	}
	return tags, nil
}

// --- SiteConfigAdapter ---

// SiteConfigAdapter implements SiteConfigReader with static values.
type SiteConfigAdapter struct {
	title       string
	url         string
	description string
	language    string
}

// NewSiteConfigAdapter creates a new SiteConfigAdapter.
func NewSiteConfigAdapter(title, url, description, language string) *SiteConfigAdapter {
	return &SiteConfigAdapter{
		title:       title,
		url:         url,
		description: description,
		language:    language,
	}
}

func (a *SiteConfigAdapter) GetSiteTitle(_ context.Context) string       { return a.title }
func (a *SiteConfigAdapter) GetSiteURL(_ context.Context) string         { return a.url }
func (a *SiteConfigAdapter) GetSiteDescription(_ context.Context) string { return a.description }
func (a *SiteConfigAdapter) GetSiteLanguage(_ context.Context) string    { return a.language }
