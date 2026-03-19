package delivery

import (
	"context"
	"errors"
	"net/http"

	"github.com/danielgtaylor/huma/v2"

	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/site/app"
	"github.com/sky-flux/cms/internal/site/domain"
)

// Executor interfaces — delivery depends on app layer via these ports.

type GetSiteExecutor interface {
	Execute(ctx context.Context) (*domain.Site, error)
}

type UpdateSiteExecutor interface {
	Execute(ctx context.Context, in app.UpdateSiteConfigInput) error
}

type CreateMenuExecutor interface {
	Execute(ctx context.Context, in app.CreateMenuInput) (*domain.Menu, error)
}

type AddMenuItemExecutor interface {
	Execute(ctx context.Context, in app.AddMenuItemInput) (*domain.MenuItem, error)
}

type CreateRedirectExecutor interface {
	Execute(ctx context.Context, in app.CreateRedirectInput) (*domain.Redirect, error)
}

// Handler holds all site delivery dependencies.
type Handler struct {
	getSite        GetSiteExecutor
	updateSite     UpdateSiteExecutor
	createMenu     CreateMenuExecutor
	addMenuItem    AddMenuItemExecutor
	createRedirect CreateRedirectExecutor
}

func NewHandler(
	getSite GetSiteExecutor,
	updateSite UpdateSiteExecutor,
	createMenu CreateMenuExecutor,
	addMenuItem AddMenuItemExecutor,
	createRedirect CreateRedirectExecutor,
) *Handler {
	return &Handler{
		getSite:        getSite,
		updateSite:     updateSite,
		createMenu:     createMenu,
		addMenuItem:    addMenuItem,
		createRedirect: createRedirect,
	}
}

// RegisterRoutes wires all site endpoints onto the Huma API.
func RegisterRoutes(api huma.API,
	getSite GetSiteExecutor,
	updateSite UpdateSiteExecutor,
	createMenu CreateMenuExecutor,
	addItem AddMenuItemExecutor,
	createRedirect CreateRedirectExecutor,
) {
	h := NewHandler(getSite, updateSite, createMenu, addItem, createRedirect)

	huma.Register(api, huma.Operation{
		OperationID: "settings-get",
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/settings",
		Summary:     "Get site configuration",
		Tags:        []string{"Settings"},
	}, h.GetSettings)

	huma.Register(api, huma.Operation{
		OperationID: "settings-update",
		Method:      http.MethodPut,
		Path:        "/api/v1/admin/settings",
		Summary:     "Update site configuration",
		Tags:        []string{"Settings"},
	}, h.UpdateSettings)

	huma.Register(api, huma.Operation{
		OperationID:   "menus-create",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/menus",
		Summary:       "Create a navigation menu",
		Tags:          []string{"Menus"},
		DefaultStatus: http.StatusCreated,
	}, h.CreateMenu)

	huma.Register(api, huma.Operation{
		OperationID:   "menus-add-item",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/menus/{id}/items",
		Summary:       "Add an item to a menu",
		Tags:          []string{"Menus"},
		DefaultStatus: http.StatusCreated,
	}, h.AddMenuItem)

	huma.Register(api, huma.Operation{
		OperationID:   "redirects-create",
		Method:        http.MethodPost,
		Path:          "/api/v1/admin/redirects",
		Summary:       "Create a URL redirect rule",
		Tags:          []string{"Redirects"},
		DefaultStatus: http.StatusCreated,
	}, h.CreateRedirect)
}

// --- Request / Response DTOs ---

type SiteBody struct {
	ID          int    `json:"id"`
	Name        string `json:"name"`
	Slug        string `json:"slug,omitempty"`
	Description string `json:"description,omitempty"`
	Language    string `json:"language"`
	Timezone    string `json:"timezone"`
	BaseURL     string `json:"base_url,omitempty"`
	LogoURL     string `json:"logo_url,omitempty"`
}

type GetSettingsResponse struct{ Body *SiteBody }

type UpdateSettingsRequest struct {
	Body struct {
		Name        string `json:"name" minLength:"1"`
		Language    string `json:"language" minLength:"2"`
		Timezone    string `json:"timezone"`
		Description string `json:"description,omitempty"`
		BaseURL     string `json:"base_url,omitempty"`
		LogoURL     string `json:"logo_url,omitempty"`
	}
}

type UpdateSettingsResponse struct{ Body *SiteBody }

type CreateMenuRequest struct {
	Body struct {
		Name string `json:"name" minLength:"1"`
		Slug string `json:"slug" minLength:"1"`
	}
}

type MenuBody struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type CreateMenuResponse struct{ Body *MenuBody }

type AddMenuItemRequest struct {
	ID   string `path:"id"`
	Body struct {
		ParentID string `json:"parent_id,omitempty"`
		Label    string `json:"label" minLength:"1"`
		URL      string `json:"url" minLength:"1"`
		Order    int    `json:"order"`
	}
}

type MenuItemBody struct {
	ID       string `json:"id"`
	MenuID   string `json:"menu_id"`
	ParentID string `json:"parent_id,omitempty"`
	Label    string `json:"label"`
	URL      string `json:"url"`
	Order    int    `json:"order"`
}

type AddMenuItemResponse struct{ Body *MenuItemBody }

type CreateRedirectRequest struct {
	Body struct {
		FromPath   string `json:"from_path" minLength:"2"`
		ToPath     string `json:"to_path" minLength:"1"`
		StatusCode int    `json:"status_code" minimum:"301" maximum:"302"`
	}
}

type RedirectBody struct {
	ID         string `json:"id"`
	FromPath   string `json:"from_path"`
	ToPath     string `json:"to_path"`
	StatusCode int    `json:"status_code"`
}

type CreateRedirectResponse struct{ Body *RedirectBody }

// --- Handlers ---

func (h *Handler) GetSettings(ctx context.Context, _ *struct{}) (*GetSettingsResponse, error) {
	site, err := h.getSite.Execute(ctx)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "get settings")
	}
	return &GetSettingsResponse{Body: toSiteBody(site)}, nil
}

func (h *Handler) UpdateSettings(ctx context.Context, req *UpdateSettingsRequest) (*UpdateSettingsResponse, error) {
	err := h.updateSite.Execute(ctx, app.UpdateSiteConfigInput{
		Name:        req.Body.Name,
		Language:    req.Body.Language,
		Timezone:    req.Body.Timezone,
		Description: req.Body.Description,
		BaseURL:     req.Body.BaseURL,
		LogoURL:     req.Body.LogoURL,
	})
	if err != nil {
		return nil, mapError(err)
	}
	site, _ := h.getSite.Execute(ctx)
	return &UpdateSettingsResponse{Body: toSiteBody(site)}, nil
}

func (h *Handler) CreateMenu(ctx context.Context, req *CreateMenuRequest) (*CreateMenuResponse, error) {
	menu, err := h.createMenu.Execute(ctx, app.CreateMenuInput{
		Name: req.Body.Name,
		Slug: req.Body.Slug,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &CreateMenuResponse{Body: &MenuBody{ID: menu.ID, Name: menu.Name, Slug: menu.Slug}}, nil
}

func (h *Handler) AddMenuItem(ctx context.Context, req *AddMenuItemRequest) (*AddMenuItemResponse, error) {
	item, err := h.addMenuItem.Execute(ctx, app.AddMenuItemInput{
		MenuID:   req.ID,
		ParentID: req.Body.ParentID,
		Label:    req.Body.Label,
		URL:      req.Body.URL,
		Order:    req.Body.Order,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &AddMenuItemResponse{Body: &MenuItemBody{
		ID: item.ID, MenuID: item.MenuID, ParentID: item.ParentID,
		Label: item.Label, URL: item.URL, Order: item.Order,
	}}, nil
}

func (h *Handler) CreateRedirect(ctx context.Context, req *CreateRedirectRequest) (*CreateRedirectResponse, error) {
	r, err := h.createRedirect.Execute(ctx, app.CreateRedirectInput{
		FromPath:   req.Body.FromPath,
		ToPath:     req.Body.ToPath,
		StatusCode: req.Body.StatusCode,
	})
	if err != nil {
		return nil, mapError(err)
	}
	return &CreateRedirectResponse{Body: &RedirectBody{
		ID: r.ID, FromPath: r.FromPath, ToPath: r.ToPath, StatusCode: r.StatusCode,
	}}, nil
}

func toSiteBody(s *domain.Site) *SiteBody {
	if s == nil {
		return &SiteBody{}
	}
	return &SiteBody{
		ID:          s.ID,
		Name:        s.Name,
		Slug:        s.Slug,
		Description: s.Description,
		Language:    s.Language,
		Timezone:    s.Timezone,
		BaseURL:     s.BaseURL,
		LogoURL:     s.LogoURL,
	}
}

func mapError(err error) error {
	switch {
	case errors.Is(err, apperror.ErrNotFound):
		return huma.NewError(http.StatusNotFound, err.Error())
	case errors.Is(err, domain.ErrEmptySiteName),
		errors.Is(err, domain.ErrEmptyLanguage),
		errors.Is(err, domain.ErrEmptyMenuName),
		errors.Is(err, domain.ErrEmptyMenuSlug),
		errors.Is(err, domain.ErrEmptyMenuItemLabel),
		errors.Is(err, domain.ErrEmptyMenuItemURL),
		errors.Is(err, domain.ErrInvalidFromPath),
		errors.Is(err, domain.ErrInvalidStatusCode),
		errors.Is(err, domain.ErrEmptyToPath):
		return huma.NewError(http.StatusUnprocessableEntity, err.Error())
	default:
		return huma.NewError(http.StatusInternalServerError, "internal error")
	}
}
