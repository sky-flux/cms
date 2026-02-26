package dashboard

import (
	"context"
	"database/sql"

	"github.com/sky-flux/cms/internal/model"
	"github.com/uptrace/bun"
)

// Repository reads dashboard statistics from the database.
// It relies on search_path being set by the Schema middleware.
type Repository struct {
	db *bun.DB
}

func NewRepository(db *bun.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) GetStats(ctx context.Context, _ string) (*DashboardStats, error) {
	stats := &DashboardStats{}

	// Posts grouped by status
	type statusCount struct {
		Status int8  `bun:"status"`
		Count  int64 `bun:"count"`
	}
	var postCounts []statusCount
	err := r.db.NewRaw(
		"SELECT COALESCE(status, 0) AS status, COUNT(*) AS count FROM sfc_site_posts WHERE deleted_at IS NULL GROUP BY status",
	).Scan(ctx, &postCounts)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	for _, pc := range postCounts {
		stats.Posts.Total += pc.Count
		switch model.PostStatus(pc.Status) {
		case model.PostStatusPublished:
			stats.Posts.Published = pc.Count
		case model.PostStatusDraft:
			stats.Posts.Draft = pc.Count
		case model.PostStatusScheduled:
			stats.Posts.Scheduled = pc.Count
		}
	}

	// Comments grouped by status
	var commentCounts []statusCount
	err = r.db.NewRaw(
		"SELECT COALESCE(status, 0) AS status, COUNT(*) AS count FROM sfc_site_comments WHERE deleted_at IS NULL GROUP BY status",
	).Scan(ctx, &commentCounts)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}
	for _, cc := range commentCounts {
		stats.Comments.Total += cc.Count
		switch model.CommentStatus(cc.Status) {
		case model.CommentStatusPending:
			stats.Comments.Pending = cc.Count
		case model.CommentStatusApproved:
			stats.Comments.Approved = cc.Count
		case model.CommentStatusSpam:
			stats.Comments.Spam = cc.Count
		}
	}

	// Media: total count and storage used
	err = r.db.NewRaw(
		"SELECT COUNT(*) AS total, COALESCE(SUM(file_size), 0) AS storage_used FROM sfc_site_media_files WHERE deleted_at IS NULL",
	).Scan(ctx, &stats.Media)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	// Users: global scope (public schema)
	err = r.db.NewRaw(
		"SELECT COUNT(*) AS total, COUNT(*) FILTER (WHERE status = ?) AS active, COUNT(*) FILTER (WHERE status = ?) AS inactive FROM public.sfc_users",
		model.UserStatusActive, model.UserStatusDisabled,
	).Scan(ctx, &stats.Users)
	if err != nil && err != sql.ErrNoRows {
		return nil, err
	}

	return stats, nil
}
