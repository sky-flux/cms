package rbac

import (
	"context"
	"database/sql"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/uptrace/bun"
)

// TemplateRepo handles persistence for role templates and their API/menu associations.
type TemplateRepo struct {
	db bun.IDB
}

// NewTemplateRepo creates a TemplateRepo backed by the given bun database handle.
func NewTemplateRepo(db bun.IDB) *TemplateRepo {
	return &TemplateRepo{db: db}
}

// List returns all role templates ordered by built_in DESC, created_at ASC.
func (r *TemplateRepo) List(ctx context.Context) ([]model.RoleTemplate, error) {
	var templates []model.RoleTemplate
	err := r.db.NewSelect().
		Model(&templates).
		OrderExpr("built_in DESC, created_at ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return templates, nil
}

// GetByID returns a single role template or apperror.NotFound.
func (r *TemplateRepo) GetByID(ctx context.Context, id string) (*model.RoleTemplate, error) {
	tmpl := new(model.RoleTemplate)
	err := r.db.NewSelect().
		Model(tmpl).
		Where("id = ?", id).
		Scan(ctx)
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, apperror.NotFound("role template not found", err)
		}
		return nil, err
	}
	return tmpl, nil
}

// Create inserts a new role template.
func (r *TemplateRepo) Create(ctx context.Context, tmpl *model.RoleTemplate) error {
	_, err := r.db.NewInsert().Model(tmpl).Exec(ctx)
	return err
}

// Update modifies an existing role template by primary key.
func (r *TemplateRepo) Update(ctx context.Context, tmpl *model.RoleTemplate) error {
	_, err := r.db.NewUpdate().Model(tmpl).WherePK().Exec(ctx)
	return err
}

// Delete removes a non-built-in role template by ID.
// Built-in templates are protected and cannot be deleted.
func (r *TemplateRepo) Delete(ctx context.Context, id string) error {
	_, err := r.db.NewDelete().
		Model((*model.RoleTemplate)(nil)).
		Where("id = ?", id).
		Where("built_in = ?", model.ToggleNo).
		Exec(ctx)
	return err
}

// GetTemplateAPIs returns all API endpoints associated with a template.
func (r *TemplateRepo) GetTemplateAPIs(ctx context.Context, templateID string) ([]model.APIEndpoint, error) {
	var apis []model.APIEndpoint
	err := r.db.NewSelect().
		Model(&apis).
		Join("JOIN sfc_role_template_apis AS rta ON rta.api_id = api.id").
		Where("rta.template_id = ?", templateID).
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return apis, nil
}

// SetTemplateAPIs replaces all API associations for a template in a single transaction.
func (r *TemplateRepo) SetTemplateAPIs(ctx context.Context, templateID string, apiIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*model.RoleTemplateAPI)(nil)).
			Where("template_id = ?", templateID).
			Exec(ctx)
		if err != nil {
			return err
		}

		if len(apiIDs) == 0 {
			return nil
		}

		items := make([]model.RoleTemplateAPI, len(apiIDs))
		for i, aid := range apiIDs {
			items[i] = model.RoleTemplateAPI{TemplateID: templateID, APIID: aid}
		}
		_, err = tx.NewInsert().Model(&items).Exec(ctx)
		return err
	})
}

// GetTemplateMenus returns all admin menus associated with a template.
func (r *TemplateRepo) GetTemplateMenus(ctx context.Context, templateID string) ([]model.AdminMenu, error) {
	var menus []model.AdminMenu
	err := r.db.NewSelect().
		Model(&menus).
		Join("JOIN sfc_role_template_menus AS rtm ON rtm.menu_id = m.id").
		Where("rtm.template_id = ?", templateID).
		OrderExpr("m.sort_order ASC").
		Scan(ctx)
	if err != nil {
		return nil, err
	}
	return menus, nil
}

// SetTemplateMenus replaces all menu associations for a template in a single transaction.
func (r *TemplateRepo) SetTemplateMenus(ctx context.Context, templateID string, menuIDs []string) error {
	return r.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		_, err := tx.NewDelete().
			Model((*model.RoleTemplateMenu)(nil)).
			Where("template_id = ?", templateID).
			Exec(ctx)
		if err != nil {
			return err
		}

		if len(menuIDs) == 0 {
			return nil
		}

		items := make([]model.RoleTemplateMenu, len(menuIDs))
		for i, mid := range menuIDs {
			items[i] = model.RoleTemplateMenu{TemplateID: templateID, MenuID: mid}
		}
		_, err = tx.NewInsert().Model(&items).Exec(ctx)
		return err
	})
}
