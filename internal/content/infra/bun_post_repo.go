package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/content/domain"
)

// bunPost is the private ORM model for the infra layer.
type bunPost struct {
	bun.BaseModel `bun:"table:sfc_posts,alias:p"`

	ID          string            `bun:"id,pk,type:uuid,default:gen_random_uuid()"`
	AuthorID    string            `bun:"author_id,notnull,type:uuid"`
	Title       string            `bun:"title,notnull"`
	Slug        string            `bun:"slug,notnull,unique"`
	Excerpt     string            `bun:"excerpt,notnull,default:''"`
	Content     string            `bun:"content,notnull,default:''"`
	Status      domain.PostStatus `bun:"status,notnull,type:smallint,default:1"`
	Version     int               `bun:"version,notnull,default:1"`
	PublishedAt *time.Time        `bun:"published_at"`
	ScheduledAt *time.Time        `bun:"scheduled_at"`
	CreatedAt   time.Time         `bun:"created_at,notnull,default:current_timestamp"`
	UpdatedAt   time.Time         `bun:"updated_at,notnull,default:current_timestamp"`
	DeletedAt   *time.Time        `bun:"deleted_at,soft_delete,nullzero"`
}

// BunPostRepo implements domain.PostRepository.
type BunPostRepo struct {
	db *bun.DB
}

func NewBunPostRepo(db *bun.DB) *BunPostRepo {
	return &BunPostRepo{db: db}
}

func (r *BunPostRepo) Save(ctx context.Context, p *domain.Post) error {
	row := postDomainToRow(p)
	_, err := r.db.NewInsert().Model(row).Exec(ctx)
	if err != nil {
		return fmt.Errorf("post_repo.Save: %w", err)
	}
	p.ID = row.ID
	return nil
}

func (r *BunPostRepo) FindByID(ctx context.Context, id string) (*domain.Post, error) {
	row := new(bunPost)
	err := r.db.NewSelect().Model(row).Where("p.id = ?", id).Scan(ctx)
	return r.scanResult(row, err)
}

func (r *BunPostRepo) FindBySlug(ctx context.Context, slug string) (*domain.Post, error) {
	row := new(bunPost)
	err := r.db.NewSelect().Model(row).Where("p.slug = ?", slug).Scan(ctx)
	return r.scanResult(row, err)
}

func (r *BunPostRepo) Update(ctx context.Context, p *domain.Post, expectedVersion int) error {
	p.IncrementVersion()
	res, err := r.db.NewUpdate().
		TableExpr("sfc_posts").
		Set("title = ?", p.Title).
		Set("slug = ?", p.Slug).
		Set("excerpt = ?", p.Excerpt).
		Set("content = ?", p.Content).
		Set("status = ?", p.Status).
		Set("version = ?", p.Version).
		Set("published_at = ?", p.PublishedAt).
		Set("scheduled_at = ?", p.ScheduledAt).
		Set("updated_at = NOW()").
		Where("id = ? AND version = ?", p.ID, expectedVersion).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("post_repo.Update: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return domain.ErrVersionConflict
	}
	return nil
}

func (r *BunPostRepo) SoftDelete(ctx context.Context, id string) error {
	_, err := r.db.NewUpdate().
		TableExpr("sfc_posts").
		Set("deleted_at = NOW()").
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	return err
}

func (r *BunPostRepo) List(ctx context.Context, f domain.PostFilter) ([]*domain.Post, int64, error) {
	var rows []bunPost
	q := r.db.NewSelect().Model(&rows)
	if f.Status != nil {
		q = q.Where("p.status = ?", *f.Status)
	}
	if f.AuthorID != "" {
		q = q.Where("p.author_id = ?", f.AuthorID)
	}
	q = q.OrderExpr("p.created_at DESC")

	offset := (f.Page - 1) * f.PerPage
	total, err := q.Limit(f.PerPage).Offset(offset).ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("post_repo.List: %w", err)
	}

	posts := make([]*domain.Post, len(rows))
	for i := range rows {
		posts[i] = postRowToDomain(&rows[i])
	}
	return posts, int64(total), nil
}

func (r *BunPostRepo) SlugExists(ctx context.Context, slug, excludeID string) (bool, error) {
	q := r.db.NewSelect().TableExpr("sfc_posts").Where("slug = ?", slug)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	return exists, err
}

func (r *BunPostRepo) scanResult(row *bunPost, err error) (*domain.Post, error) {
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, domain.ErrPostNotFound
		}
		return nil, err
	}
	return postRowToDomain(row), nil
}

func postDomainToRow(p *domain.Post) *bunPost {
	return &bunPost{
		ID:          p.ID,
		AuthorID:    p.AuthorID,
		Title:       p.Title,
		Slug:        p.Slug,
		Excerpt:     p.Excerpt,
		Content:     p.Content,
		Status:      p.Status,
		Version:     p.Version,
		PublishedAt: p.PublishedAt,
		ScheduledAt: p.ScheduledAt,
	}
}

func postRowToDomain(r *bunPost) *domain.Post {
	return &domain.Post{
		ID:          r.ID,
		AuthorID:    r.AuthorID,
		Title:       r.Title,
		Slug:        r.Slug,
		Excerpt:     r.Excerpt,
		Content:     r.Content,
		Status:      r.Status,
		Version:     r.Version,
		PublishedAt: r.PublishedAt,
		ScheduledAt: r.ScheduledAt,
		CreatedAt:   r.CreatedAt,
		UpdatedAt:   r.UpdatedAt,
	}
}
