package redirect

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// Repo implements RedirectRepository using uptrace/bun.
type Repo struct {
	db *bun.DB
}

// NewRepo creates a new redirect repository.
func NewRepo(db *bun.DB) *Repo {
	return &Repo{db: db}
}

func (r *Repo) List(ctx context.Context, filter ListFilter) ([]model.Redirect, int64, error) {
	var redirects []model.Redirect

	q := r.db.NewSelect().Model(&redirects)

	if filter.Query != "" {
		like := "%" + filter.Query + "%"
		q = q.WhereGroup(" AND ", func(sq *bun.SelectQuery) *bun.SelectQuery {
			return sq.Where("source_path ILIKE ?", like).WhereOr("target_url ILIKE ?", like)
		})
	}
	if filter.StatusCode != 0 {
		q = q.Where("status_code = ?", filter.StatusCode)
	}
	if filter.Status != "" {
		q = q.Where("status = ?", stringToRedirectStatus(filter.Status))
	}

	switch filter.Sort {
	case "hit_count:desc":
		q = q.OrderExpr("hit_count DESC")
	case "last_hit_at:desc":
		q = q.OrderExpr("last_hit_at DESC NULLS LAST")
	case "source_path:asc":
		q = q.OrderExpr("source_path ASC")
	default: // "created_at:desc" or empty
		q = q.OrderExpr("created_at DESC")
	}

	total, err := q.Count(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("redirect list count: %w", err)
	}

	offset := (filter.Page - 1) * filter.PerPage
	err = q.Limit(filter.PerPage).Offset(offset).Scan(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("redirect list: %w", err)
	}

	return redirects, int64(total), nil
}

func (r *Repo) GetByID(ctx context.Context, id string) (*model.Redirect, error) {
	rd := new(model.Redirect)
	err := r.db.NewSelect().Model(rd).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("redirect not found", err)
		}
		return nil, fmt.Errorf("redirect get by id: %w", err)
	}
	return rd, nil
}

func (r *Repo) Create(ctx context.Context, redirect *model.Redirect) error {
	_, err := r.db.NewInsert().Model(redirect).Exec(ctx)
	if err != nil {
		return fmt.Errorf("redirect create: %w", err)
	}
	return nil
}

func (r *Repo) Update(ctx context.Context, redirect *model.Redirect) error {
	_, err := r.db.NewUpdate().Model(redirect).WherePK().Exec(ctx)
	if err != nil {
		return fmt.Errorf("redirect update: %w", err)
	}
	return nil
}

func (r *Repo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().Model((*model.Redirect)(nil)).Where("id = ?", id).Exec(ctx)
	if err != nil {
		return fmt.Errorf("redirect delete: %w", err)
	}
	return nil
}

func (r *Repo) BatchDelete(ctx context.Context, ids []string) (int64, error) {
	res, err := r.db.NewDelete().
		Model((*model.Redirect)(nil)).
		Where("id IN (?)", bun.In(ids)).
		Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("redirect batch delete: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) SourcePathExists(ctx context.Context, path string, excludeID string) (bool, error) {
	q := r.db.NewSelect().Model((*model.Redirect)(nil)).Where("source_path = ?", path)
	if excludeID != "" {
		q = q.Where("id != ?", excludeID)
	}
	exists, err := q.Exists(ctx)
	if err != nil {
		return false, fmt.Errorf("redirect source path exists: %w", err)
	}
	return exists, nil
}

func (r *Repo) BulkInsert(ctx context.Context, redirects []*model.Redirect) (int64, error) {
	if len(redirects) == 0 {
		return 0, nil
	}
	res, err := r.db.NewInsert().Model(&redirects).Exec(ctx)
	if err != nil {
		return 0, fmt.Errorf("redirect bulk insert: %w", err)
	}
	n, _ := res.RowsAffected()
	return n, nil
}

func (r *Repo) ListAll(ctx context.Context) ([]model.Redirect, error) {
	var redirects []model.Redirect
	err := r.db.NewSelect().Model(&redirects).OrderExpr("source_path ASC").Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("redirect list all: %w", err)
	}
	return redirects, nil
}

func stringToRedirectStatus(s string) model.RedirectStatus {
	switch s {
	case "active":
		return model.RedirectStatusActive
	case "disabled":
		return model.RedirectStatusDisabled
	default:
		return model.RedirectStatusActive
	}
}
