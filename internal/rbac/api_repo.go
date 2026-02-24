package rbac

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"strings"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// APIRepo handles persistence for the sfc_apis table.
type APIRepo struct {
	db *bun.DB
}

// NewAPIRepo creates an APIRepo backed by the given bun.DB.
func NewAPIRepo(db *bun.DB) *APIRepo {
	return &APIRepo{db: db}
}

// UpsertBatch inserts or updates a batch of API endpoints.
// On conflict (method, path), existing rows are updated.
func (r *APIRepo) UpsertBatch(ctx context.Context, endpoints []model.APIEndpoint) error {
	for i := range endpoints {
		endpoints[i].Status = true
		_, err := r.db.NewInsert().
			Model(&endpoints[i]).
			On("CONFLICT (method, path) DO UPDATE").
			Set("name = EXCLUDED.name").
			Set("description = EXCLUDED.description").
			Set(`"group" = EXCLUDED."group"`).
			Set("status = EXCLUDED.status").
			Set("updated_at = NOW()").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("api upsert %s %s: %w", endpoints[i].Method, endpoints[i].Path, err)
		}
	}
	return nil
}

// DisableStale marks API endpoints as inactive when they are no longer
// present in the running application. activeKeys is a list of "METHOD:path"
// strings representing the currently registered routes.
func (r *APIRepo) DisableStale(ctx context.Context, activeKeys []string) error {
	if len(activeKeys) == 0 {
		// No active keys means disable everything.
		_, err := r.db.NewUpdate().
			Model((*model.APIEndpoint)(nil)).
			Set("status = false").
			Where("status = true").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("api disable all: %w", err)
		}
		return nil
	}

	// Build placeholders for the IN clause using concatenation.
	placeholders := make([]string, len(activeKeys))
	args := make([]interface{}, len(activeKeys))
	for i, key := range activeKeys {
		placeholders[i] = "?"
		args[i] = key
	}

	query := fmt.Sprintf(
		`UPDATE sfc_apis SET status = false, updated_at = NOW() WHERE method || ':' || path NOT IN (%s) AND status = true`,
		strings.Join(placeholders, ", "),
	)
	_, err := r.db.ExecContext(ctx, query, args...)
	if err != nil {
		return fmt.Errorf("api disable stale: %w", err)
	}
	return nil
}

// List returns all API endpoints ordered by group, method, path.
func (r *APIRepo) List(ctx context.Context) ([]model.APIEndpoint, error) {
	var apis []model.APIEndpoint
	err := r.db.NewSelect().
		Model(&apis).
		OrderExpr(`"group" ASC, method ASC, path ASC`).
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("api list: %w", err)
	}
	return apis, nil
}

// ListByGroup returns API endpoints filtered by the given group name.
func (r *APIRepo) ListByGroup(ctx context.Context, group string) ([]model.APIEndpoint, error) {
	var apis []model.APIEndpoint
	err := r.db.NewSelect().
		Model(&apis).
		Where(`"group" = ?`, group).
		OrderExpr("method ASC, path ASC").
		Scan(ctx)
	if err != nil {
		return nil, fmt.Errorf("api list by group: %w", err)
	}
	return apis, nil
}

// GetByMethodPath returns a single API endpoint matching the given HTTP
// method and path. Returns apperror.NotFound if no match exists.
func (r *APIRepo) GetByMethodPath(ctx context.Context, method, path string) (*model.APIEndpoint, error) {
	api := new(model.APIEndpoint)
	err := r.db.NewSelect().
		Model(api).
		Where("method = ?", method).
		Where("path = ?", path).
		Scan(ctx)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, apperror.NotFound("api endpoint not found", err)
		}
		return nil, fmt.Errorf("api get by method path: %w", err)
	}
	return api, nil
}
