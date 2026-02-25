package comment

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements CommentRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new comment repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context, filter ListFilter) ([]CommentRow, int64, error) {
	var rows []CommentRow

	q := r.db.NewSelect().
		TableExpr("sfc_site_comments AS cm").
		ColumnExpr("cm.*").
		ColumnExpr("p.title AS post_title").
		ColumnExpr("p.slug AS post_slug").
		Join("LEFT JOIN sfc_site_posts AS p ON p.id = cm.post_id").
		Where("cm.deleted_at IS NULL")

	if filter.PostID != "" {
		q = q.Where("cm.post_id = ?", filter.PostID)
	}
	if filter.Status != "" {
		q = q.Where("cm.status = ?", StringToCommentStatus(filter.Status))
	}
	if filter.Query != "" {
		q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			like := "%" + filter.Query + "%"
			return sq.Where("cm.content ILIKE ?", like).WhereOr("cm.author_name ILIKE ?", like)
		})
	}

	q = q.OrderExpr("cm.created_at DESC")

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("comment list count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	err = q.Limit(filter.PerPage).Offset(offset).Scan(ctx, &rows)
	if err != nil {
		return nil, 0, fmt.Errorf("comment list: %w", err)
	}

	return rows, int64(total), nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Comment, error) {
	comment := new(model.Comment)
	err := r.db.NewSelect().Model(comment).Where("id = ? AND deleted_at IS NULL", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("comment not found", err)
		}
		return nil, fmt.Errorf("comment get by id: %w", err)
	}
	return comment, nil
}

func (r *Repo) GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error) {
	var children []*model.Comment
	err := r.db.NewSelect().
		Model(&children).
		Where("parent_id = ? AND deleted_at IS NULL", parentID).
		OrderExpr("created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("comment get children: %w", err)
	}
	return children, nil
}

func (r *Repo) UpdateStatus(ctx context.Context, id string, status model.CommentStatus) error {
	_, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("status = ?", status).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment update status: %w", err)
	}
	return nil
}

func (r *Repo) UpdatePinned(ctx context.Context, id string, pinned model.Toggle) error {
	_, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("pinned = ?", pinned).
		Where("id = ? AND deleted_at IS NULL", id).
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment update pinned: %w", err)
	}
	return nil
}

func (r *Repo) Create(ctx context.Context, comment *model.Comment) error {
	_, err := r.db.NewInsert().Model(comment).Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment create: %w", err)
	}
	return nil
}

func (r *Repo) BatchUpdateStatus(ctx context.Context, ids []string, status model.CommentStatus) (int64, error) {
	res, err := r.db.NewUpdate().
		Model((*model.Comment)(nil)).
		Set("status = ?", status).
		Where("id IN (?)", bun.In(ids)).
		Where("deleted_at IS NULL").
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("comment batch update status: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.Comment)(nil)).
		Where("id = ?", id).
		ForceDelete().
		Exec(ctx)
	if err != nil {
		return fmt.Errorf("comment delete: %w", err)
	}
	return nil
}

func (r *Repo) CountPinnedByPost(ctx context.Context, postID string) (int64, error) {
	count, err := r.db.NewSelect().
		Model((*model.Comment)(nil)).
		Where("post_id = ?", postID).
		Where("pinned = ?", model.ToggleYes).
		Where("parent_id IS NULL").
		Where("deleted_at IS NULL").
		Count(ctx)
	if err != nil {
		return 0, fmt.Errorf("comment count pinned: %w", err)
	}
	return int64(count), nil
}

func (r *Repo) GetParentChainDepth(ctx context.Context, commentID string) (int, error) {
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
			return 0, fmt.Errorf("comment parent chain: %w", err)
		}
		if comment.ParentID == nil {
			break
		}
		depth++
		currentID = *comment.ParentID
	}
	return depth, nil
}
