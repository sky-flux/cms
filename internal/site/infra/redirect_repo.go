package infra

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/site/domain"
)

type redirectRow struct {
	bun.BaseModel `bun:"table:sfc_redirects,alias:rd"`

	ID         string    `bun:"id,pk,type:uuid,default:uuidv7()"`
	FromPath   string    `bun:"from_path,notnull,unique"`
	ToPath     string    `bun:"to_path,notnull"`
	StatusCode int       `bun:"status_code,notnull,default:301"`
	CreatedAt  time.Time `bun:"created_at,nullzero,default:now()"`
	UpdatedAt  time.Time `bun:"updated_at,nullzero,default:now()"`
}

// BunRedirectRepo implements domain.RedirectRepository using uptrace/bun.
type BunRedirectRepo struct {
	db *bun.DB
}

func NewBunRedirectRepo(db *bun.DB) *BunRedirectRepo {
	return &BunRedirectRepo{db: db}
}

func (r *BunRedirectRepo) Save(ctx context.Context, redirect *domain.Redirect) error {
	row := &redirectRow{
		ID: redirect.ID, FromPath: redirect.FromPath,
		ToPath: redirect.ToPath, StatusCode: redirect.StatusCode,
	}
	if redirect.ID == "" {
		if _, err := r.db.NewInsert().Model(row).Exec(ctx); err != nil {
			return fmt.Errorf("insert redirect: %w", err)
		}
		redirect.ID = row.ID
		return nil
	}
	if _, err := r.db.NewUpdate().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("update redirect: %w", err)
	}
	return nil
}

func (r *BunRedirectRepo) FindByPath(ctx context.Context, fromPath string) (*domain.Redirect, error) {
	row := &redirectRow{}
	err := r.db.NewSelect().Model(row).Where("rd.from_path = ?", fromPath).Scan(ctx)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, apperror.ErrNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("find redirect: %w", err)
	}
	return &domain.Redirect{
		ID: row.ID, FromPath: row.FromPath, ToPath: row.ToPath,
		StatusCode: row.StatusCode, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
	}, nil
}

func (r *BunRedirectRepo) List(ctx context.Context, offset, limit int) ([]*domain.Redirect, int, error) {
	var rows []redirectRow
	count, err := r.db.NewSelect().Model(&rows).
		OrderExpr("created_at DESC").
		Offset(offset).Limit(limit).
		ScanAndCount(ctx)
	if err != nil {
		return nil, 0, fmt.Errorf("list redirects: %w", err)
	}
	redirects := make([]*domain.Redirect, len(rows))
	for i, row := range rows {
		redirects[i] = &domain.Redirect{
			ID: row.ID, FromPath: row.FromPath, ToPath: row.ToPath,
			StatusCode: row.StatusCode, CreatedAt: row.CreatedAt, UpdatedAt: row.UpdatedAt,
		}
	}
	return redirects, count, nil
}

func (r *BunRedirectRepo) Delete(ctx context.Context, id string) error {
	row := &redirectRow{ID: id}
	if _, err := r.db.NewDelete().Model(row).WherePK().Exec(ctx); err != nil {
		return fmt.Errorf("delete redirect: %w", err)
	}
	return nil
}
