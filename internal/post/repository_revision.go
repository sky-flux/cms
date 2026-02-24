package post

import (
	"context"
	"database/sql"
	"errors"
	"fmt"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

type RevisionRepo struct {
	db *bun.DB
}

func NewRevisionRepo(db *bun.DB) *RevisionRepo {
	return &RevisionRepo{db: db}
}

func (r *RevisionRepo) List(ctx context.Context, postID string) ([]model.PostRevision, error) {
	var revs []model.PostRevision
	err := r.db.NewSelect().
		Model(&revs).
		Where("post_id = ?", postID).
		OrderExpr("version DESC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("revision list: %w", err)
	}
	return revs, nil
}

func (r *RevisionRepo) GetByID(ctx context.Context, id string) (*model.PostRevision, error) {
	rev := new(model.PostRevision)
	err := r.db.NewSelect().Model(rev).Where("id = ?", id).Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("revision not found", err)
		}
		return nil, fmt.Errorf("revision get by id: %w", err)
	}
	return rev, nil
}

func (r *RevisionRepo) Create(ctx context.Context, rev *model.PostRevision) error {
	_, err := r.db.NewInsert().Model(rev).Exec(ctx)
	if err != nil {
		return fmt.Errorf("revision create: %w", err)
	}
	return nil
}
