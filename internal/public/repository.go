package public

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// ---------------------------------------------------------------------------
// PostRepoAdapter implements PostReader
// ---------------------------------------------------------------------------

// PostRepoAdapter adapts the bun DB to the PostReader interface for public queries.
type PostRepoAdapter struct {
	db *bun.DB
}

// NewPostRepoAdapter creates a new PostRepoAdapter.
func NewPostRepoAdapter(db *bun.DB) *PostRepoAdapter {
	return &PostRepoAdapter{db: db}
}

// List returns published posts with optional category/tag slug filtering, paginated.
func (r *PostRepoAdapter) List(ctx context.Context, f PostListFilter) ([]model.Post, int64, error) {
	var posts []model.Post
	q := r.db.NewSelect().
		Model(&posts).
		Relation("Author").
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL")

	// Filter by category slug via subquery on mapping + categories table.
	if f.Category != "" {
		q = q.Where(`p.id IN (
			SELECT pcm.post_id FROM sfc_site_post_category_map AS pcm
			JOIN sfc_site_categories AS c ON c.id = pcm.category_id
			WHERE c.slug = ?
		)`, f.Category)
	}

	// Filter by tag slug via subquery on mapping + tags table.
	if f.Tag != "" {
		q = q.Where(`p.id IN (
			SELECT ptm.post_id FROM sfc_site_post_tag_map AS ptm
			JOIN sfc_site_tags AS t ON t.id = ptm.tag_id
			WHERE t.slug = ?
		)`, f.Tag)
	}

	// Sort.
	switch f.Sort {
	case "title:asc":
		q = q.OrderExpr("p.title ASC")
	case "view_count:desc":
		q = q.OrderExpr("p.view_count DESC")
	default:
		q = q.OrderExpr("p.published_at DESC NULLS LAST")
	}

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public post list count: %w", err)
	}

	perPage := f.PerPage
	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	page := f.Page
	if page <= 0 {
		page = 1
	}

	err = q.Limit(perPage).Offset((page - 1) * perPage).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public post list: %w", err)
	}

	return posts, int64(count), nil
}

// GetBySlug returns a single post by slug with Author loaded.
func (r *PostRepoAdapter) GetBySlug(ctx context.Context, slug string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.slug = ?", slug).
		Where("p.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("public post get by slug: %w", err)
	}
	return post, nil
}

// GetByID returns a single post by ID with Author loaded.
func (r *PostRepoAdapter) GetByID(ctx context.Context, id string) (*model.Post, error) {
	post := new(model.Post)
	err := r.db.NewSelect().
		Model(post).
		Relation("Author").
		Where("p.id = ?", id).
		Where("p.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("post not found", err)
		}
		return nil, fmt.Errorf("public post get by id: %w", err)
	}
	return post, nil
}

// LoadRelations is a no-op because Author is already loaded via Relation.
func (r *PostRepoAdapter) LoadRelations(_ context.Context, _ *model.Post) error {
	return nil
}

// IncrementViewCount atomically increments the view_count of a post by 1.
func (r *PostRepoAdapter) IncrementViewCount(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().
		Model((*model.Post)(nil)).
		Set(`"view_count" = view_count + 1`).
		Where("id = ?", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("public post increment view count: %w", err)
	}
	return nil
}

// ---------------------------------------------------------------------------
// CategoryRepoAdapter implements CategoryReader
// ---------------------------------------------------------------------------

// CategoryRepoAdapter adapts the bun DB to the CategoryReader interface.
type CategoryRepoAdapter struct {
	db *bun.DB
}

// NewCategoryRepoAdapter creates a new CategoryRepoAdapter.
func NewCategoryRepoAdapter(db *bun.DB) *CategoryRepoAdapter {
	return &CategoryRepoAdapter{db: db}
}

// List returns all categories ordered by sort_order, name.
func (r *CategoryRepoAdapter) List(ctx context.Context) ([]model.Category, error) {
	var cats []model.Category
	err := r.db.NewSelect().
		Model(&cats).
		OrderExpr("c.sort_order ASC, c.name ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("public category list: %w", err)
	}
	return cats, nil
}

// CountPosts counts published, non-deleted posts in a given category.
func (r *CategoryRepoAdapter) CountPosts(ctx context.Context, categoryID string) (int64, error) {
	count, err := r.db.NewSelect().
		TableExpr("sfc_site_post_category_map AS pcm").
		Join("JOIN sfc_site_posts AS p ON p.id = pcm.post_id").
		Where("pcm.category_id = ?", categoryID).
		Where("p.status = ?", model.PostStatusPublished).
		Where("p.deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("public category count posts: %w", err)
	}
	return int64(count), nil
}

// ---------------------------------------------------------------------------
// TagRepoAdapter implements TagReader
// ---------------------------------------------------------------------------

// TagRepoAdapter adapts the bun DB to the TagReader interface.
type TagRepoAdapter struct {
	db *bun.DB
}

// NewTagRepoAdapter creates a new TagRepoAdapter.
func NewTagRepoAdapter(db *bun.DB) *TagRepoAdapter {
	return &TagRepoAdapter{db: db}
}

// ListPublic returns all tags with their published post counts.
// Sort options: "post_count:desc" (default) or "name:asc".
func (r *TagRepoAdapter) ListPublic(ctx context.Context, sort string) ([]TagWithCount, error) {
	var tags []TagWithCount
	q := r.db.NewSelect().
		TableExpr("sfc_site_tags AS t").
		ColumnExpr("t.*").
		ColumnExpr("COUNT(p.id) AS post_count").
		Join("LEFT JOIN sfc_site_post_tag_map AS ptm ON ptm.tag_id = t.id").
		Join("LEFT JOIN sfc_site_posts AS p ON p.id = ptm.post_id AND p.status = ? AND p.deleted_at IS NULL", model.PostStatusPublished).
		GroupExpr("t.id")

	switch sort {
	case "name:asc":
		q = q.OrderExpr("t.name ASC")
	default:
		q = q.OrderExpr("post_count DESC, t.name ASC")
	}

	err := q.Scan(ctx, &tags)
	if err != nil {
		return nil, fmt.Errorf("public tag list: %w", err)
	}
	return tags, nil
}

// ---------------------------------------------------------------------------
// CommentRepoAdapter implements CommentReader
// ---------------------------------------------------------------------------

// CommentRepoAdapter adapts the bun DB to the CommentReader interface.
type CommentRepoAdapter struct {
	db *bun.DB
}

// NewCommentRepoAdapter creates a new CommentRepoAdapter.
func NewCommentRepoAdapter(db *bun.DB) *CommentRepoAdapter {
	return &CommentRepoAdapter{db: db}
}

// ListByPost returns top-level approved comments for a post with recursive replies loaded.
func (r *CommentRepoAdapter) ListByPost(ctx context.Context, postID string, page, perPage int) ([]model.Comment, int64, error) {
	var comments []model.Comment

	q := r.db.NewSelect().
		Model(&comments).
		Where("cm.post_id = ?", postID).
		Where("cm.parent_id IS NULL").
		Where("cm.status = ?", model.CommentStatusApproved).
		Where("cm.deleted_at IS NULL").
		OrderExpr("cm.pinned DESC, cm.created_at ASC")

	count, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public comment list count: %w", err)
	}

	if perPage <= 0 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	if page <= 0 {
		page = 1
	}

	err = q.Limit(perPage).Offset((page - 1) * perPage).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("public comment list: %w", err)
	}

	// Recursively load approved replies for each top-level comment.
	for i := range comments {
		if err := r.loadReplies(ctx, &comments[i]); err != nil {
			return nil, 0, fmt.Errorf("public comment load replies: %w", err)
		}
	}

	return comments, int64(count), nil
}

// loadReplies recursively loads approved, non-deleted child comments.
func (r *CommentRepoAdapter) loadReplies(ctx context.Context, parent *model.Comment) error {
	var children []*model.Comment
	err := r.db.NewSelect().
		Model(&children).
		Where("cm.parent_id = ?", parent.ID).
		Where("cm.status = ?", model.CommentStatusApproved).
		Where("cm.deleted_at IS NULL").
		OrderExpr("cm.created_at ASC").
		Scan(ctx)
	if err != nil {
		return fmt.Errorf("load replies for %s: %w", parent.ID, err)
	}

	parent.Children = children

	for _, child := range children {
		if err := r.loadReplies(ctx, child); err != nil {
			return err
		}
	}
	return nil
}

// Create inserts a new comment.
func (r *CommentRepoAdapter) Create(ctx context.Context, c *model.Comment) error {
	_, err := r.db.NewInsert().Model(c).Exec(ctx)
	if err != nil {
		return fmt.Errorf("public comment create: %w", err)
	}
	return nil
}

// GetByID returns a comment by ID (non-deleted only).
func (r *CommentRepoAdapter) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	comment := new(model.Comment)
	err := r.db.NewSelect().
		Model(comment).
		Where("cm.id = ?", id).
		Where("cm.deleted_at IS NULL").
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("comment not found", err)
		}
		return nil, fmt.Errorf("public comment get by id: %w", err)
	}
	return comment, nil
}

// GetParentChainDepth walks up the parent chain and returns the depth (0 = root).
func (r *CommentRepoAdapter) GetParentChainDepth(ctx context.Context, commentID string) (int, error) {
	depth := 0
	currentID := commentID
	for depth < 5 { // safety limit
		comment := new(model.Comment)
		err := r.db.NewSelect().
			Model(comment).
			Column("parent_id").
			Where("id = ?", currentID).
			Scan(ctx)
		if err != nil {
			if errors.Is(err, sql.ErrNoRows) {
				break
			}
			return 0, fmt.Errorf("public comment parent chain: %w", err)
		}
		if comment.ParentID == nil {
			break
		}
		depth++
		currentID = *comment.ParentID
	}
	return depth, nil
}

// ---------------------------------------------------------------------------
// MenuRepoAdapter implements MenuReader
// ---------------------------------------------------------------------------

// MenuRepoAdapter adapts the bun DB to the MenuReader interface.
type MenuRepoAdapter struct {
	db *bun.DB
}

// NewMenuRepoAdapter creates a new MenuRepoAdapter.
func NewMenuRepoAdapter(db *bun.DB) *MenuRepoAdapter {
	return &MenuRepoAdapter{db: db}
}

// GetByLocation returns a site menu by its location field.
func (r *MenuRepoAdapter) GetByLocation(ctx context.Context, location string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().
		Model(menu).
		Where("sm.location = ?", location).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("public menu get by location: %w", err)
	}
	return menu, nil
}

// GetBySlug returns a site menu by its slug.
func (r *MenuRepoAdapter) GetBySlug(ctx context.Context, slug string) (*model.SiteMenu, error) {
	menu := new(model.SiteMenu)
	err := r.db.NewSelect().
		Model(menu).
		Where("sm.slug = ?", slug).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("menu not found", err)
		}
		return nil, fmt.Errorf("public menu get by slug: %w", err)
	}
	return menu, nil
}

// ListItemsByMenuID returns all items for a menu, ordered by sort_order.
func (r *MenuRepoAdapter) ListItemsByMenuID(ctx context.Context, menuID string) ([]*model.SiteMenuItem, error) {
	var items []*model.SiteMenuItem
	err := r.db.NewSelect().
		Model(&items).
		Where("mi.menu_id = ?", menuID).
		OrderExpr("mi.sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("public menu list items: %w", err)
	}
	return items, nil
}

// ---------------------------------------------------------------------------
// PreviewRepoAdapter implements PreviewReader
// ---------------------------------------------------------------------------

// PreviewRepoAdapter adapts the bun DB to the PreviewReader interface.
type PreviewRepoAdapter struct {
	db *bun.DB
}

// NewPreviewRepoAdapter creates a new PreviewRepoAdapter.
func NewPreviewRepoAdapter(db *bun.DB) *PreviewRepoAdapter {
	return &PreviewRepoAdapter{db: db}
}

// GetByHash returns a preview token by its hash.
func (r *PreviewRepoAdapter) GetByHash(ctx context.Context, hash string) (*model.PreviewToken, error) {
	token := new(model.PreviewToken)
	err := r.db.NewSelect().
		Model(token).
		Where("pvt.token_hash = ?", hash).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("preview token not found", err)
		}
		return nil, fmt.Errorf("public preview get by hash: %w", err)
	}
	return token, nil
}
