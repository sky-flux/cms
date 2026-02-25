package menu

import (
	"context"
	"regexp"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9]([a-z0-9-]*[a-z0-9])?$`)

// Service handles menu business logic.
type Service struct {
	menuRepo MenuRepository
	itemRepo MenuItemRepository
	audit    audit.Logger
}

// NewService creates a new menu service.
func NewService(menuRepo MenuRepository, itemRepo MenuItemRepository, audit audit.Logger) *Service {
	return &Service{menuRepo: menuRepo, itemRepo: itemRepo, audit: audit}
}

// ListMenus returns all menus with item counts, optionally filtered by location.
func (s *Service) ListMenus(ctx context.Context, location string) ([]MenuResp, error) {
	menus, err := s.menuRepo.List(ctx, location)
	if err != nil {
		return nil, err
	}

	out := make([]MenuResp, len(menus))
	for i := range menus {
		count, _ := s.menuRepo.CountItems(ctx, menus[i].ID)
		out[i] = ToMenuResp(&menus[i], count)
	}
	return out, nil
}

// GetMenu returns a menu with its nested item tree.
func (s *Service) GetMenu(ctx context.Context, id string) (*MenuDetailResp, error) {
	menu, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	items, err := s.itemRepo.ListByMenuID(ctx, id)
	if err != nil {
		return nil, err
	}

	tree := BuildMenuTree(items)
	itemResps := make([]*MenuItemResp, len(tree))
	for i, item := range tree {
		itemResps[i] = ToMenuItemResp(item)
	}

	return &MenuDetailResp{
		ID:          menu.ID,
		Name:        menu.Name,
		Slug:        menu.Slug,
		Location:    menu.Location,
		Description: menu.Description,
		CreatedAt:   menu.CreatedAt,
		UpdatedAt:   menu.UpdatedAt,
		Items:       itemResps,
	}, nil
}

// CreateMenu creates a new navigation menu.
func (s *Service) CreateMenu(ctx context.Context, req *CreateMenuReq) (*MenuResp, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, apperror.Validation("invalid slug format", nil)
	}

	exists, err := s.menuRepo.SlugExists(ctx, req.Slug, "")
	if err != nil {
		return nil, err
	}
	if exists {
		return nil, apperror.Conflict("menu slug already exists", nil)
	}

	menu := &model.SiteMenu{
		Name:        req.Name,
		Slug:        req.Slug,
		Location:    req.Location,
		Description: req.Description,
	}

	if err := s.menuRepo.Create(ctx, menu); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "menu",
		ResourceID:   menu.ID,
	})

	resp := ToMenuResp(menu, 0)
	return &resp, nil
}

// UpdateMenu updates a menu's metadata.
func (s *Service) UpdateMenu(ctx context.Context, id string, req *UpdateMenuReq) (*MenuResp, error) {
	menu, err := s.menuRepo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil {
		menu.Name = *req.Name
	}
	if req.Slug != nil {
		if !slugRegex.MatchString(*req.Slug) {
			return nil, apperror.Validation("invalid slug format", nil)
		}
		exists, err := s.menuRepo.SlugExists(ctx, *req.Slug, id)
		if err != nil {
			return nil, err
		}
		if exists {
			return nil, apperror.Conflict("menu slug already exists", nil)
		}
		menu.Slug = *req.Slug
	}
	if req.Location != nil {
		menu.Location = *req.Location
	}
	if req.Description != nil {
		menu.Description = *req.Description
	}

	if err := s.menuRepo.Update(ctx, menu); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "menu",
		ResourceID:   id,
	})

	count, _ := s.menuRepo.CountItems(ctx, id)
	resp := ToMenuResp(menu, count)
	return &resp, nil
}

// DeleteMenu deletes a menu (FK CASCADE removes items).
func (s *Service) DeleteMenu(ctx context.Context, id string) error {
	if _, err := s.menuRepo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.menuRepo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "menu",
		ResourceID:   id,
	})
	return nil
}

// AddItem adds an item to a menu.
func (s *Service) AddItem(ctx context.Context, menuID string, req *CreateMenuItemReq) (*MenuItemResp, error) {
	// Verify menu exists
	if _, err := s.menuRepo.GetByID(ctx, menuID); err != nil {
		return nil, err
	}

	// Type-based validation
	itemType := StringToMenuItemType(req.Type)
	if itemType == model.MenuItemTypeCustom && req.URL == "" {
		return nil, apperror.Validation("url required for custom menu item", nil)
	}
	if itemType != model.MenuItemTypeCustom && req.ReferenceID == nil {
		return nil, apperror.Validation("reference_id required for non-custom menu item", nil)
	}

	// Validate parent belongs to same menu and check depth
	if req.ParentID != nil {
		belongs, err := s.itemRepo.BelongsToMenu(ctx, *req.ParentID, menuID)
		if err != nil {
			return nil, err
		}
		if !belongs {
			return nil, apperror.Validation("parent item does not belong to this menu", nil)
		}
		depth, err := s.itemRepo.GetDepth(ctx, *req.ParentID)
		if err != nil {
			return nil, err
		}
		if depth >= 2 {
			return nil, apperror.Validation("maximum 3-level nesting exceeded", nil)
		}
	}

	target := req.Target
	if target == "" {
		target = "_self"
	}

	item := &model.SiteMenuItem{
		MenuID:      menuID,
		ParentID:    req.ParentID,
		Label:       req.Label,
		URL:         req.URL,
		Target:      target,
		Type:        itemType,
		ReferenceID: req.ReferenceID,
		SortOrder:   req.SortOrder,
		Icon:        req.Icon,
		CSSClass:    req.CSSClass,
	}

	if err := s.itemRepo.Create(ctx, item); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "menu_item",
		ResourceID:   item.ID,
	})

	return ToMenuItemResp(item), nil
}

// UpdateItem updates a menu item.
func (s *Service) UpdateItem(ctx context.Context, menuID, itemID string, req *UpdateMenuItemReq) (*MenuItemResp, error) {
	belongs, err := s.itemRepo.BelongsToMenu(ctx, itemID, menuID)
	if err != nil {
		return nil, err
	}
	if !belongs {
		return nil, apperror.NotFound("menu item not found in this menu", nil)
	}

	item, err := s.itemRepo.GetByID(ctx, itemID)
	if err != nil {
		return nil, err
	}

	if req.Label != nil {
		item.Label = *req.Label
	}
	if req.URL != nil {
		item.URL = *req.URL
	}
	if req.Target != nil {
		item.Target = *req.Target
	}
	if req.Type != nil {
		item.Type = StringToMenuItemType(*req.Type)
	}
	if req.ReferenceID != nil {
		item.ReferenceID = req.ReferenceID
	}
	if req.SortOrder != nil {
		item.SortOrder = *req.SortOrder
	}
	if req.Icon != nil {
		item.Icon = *req.Icon
	}
	if req.CSSClass != nil {
		item.CSSClass = *req.CSSClass
	}
	if req.Status != nil {
		item.Status = StringToMenuItemStatus(*req.Status)
	}
	if req.ParentID != nil {
		item.ParentID = req.ParentID
	}

	if err := s.itemRepo.Update(ctx, item); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "menu_item",
		ResourceID:   itemID,
	})

	return ToMenuItemResp(item), nil
}

// DeleteItem deletes a menu item (FK CASCADE removes children).
func (s *Service) DeleteItem(ctx context.Context, menuID, itemID string) error {
	belongs, err := s.itemRepo.BelongsToMenu(ctx, itemID, menuID)
	if err != nil {
		return err
	}
	if !belongs {
		return apperror.NotFound("menu item not found in this menu", nil)
	}

	if err := s.itemRepo.Delete(ctx, itemID); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "menu_item",
		ResourceID:   itemID,
	})
	return nil
}

// ReorderItems batch-updates item positions within a menu.
func (s *Service) ReorderItems(ctx context.Context, menuID string, req *ReorderReq) error {
	// Validate all items belong to this menu
	for _, item := range req.Items {
		belongs, err := s.itemRepo.BelongsToMenu(ctx, item.ID, menuID)
		if err != nil {
			return err
		}
		if !belongs {
			return apperror.Validation("item does not belong to this menu: "+item.ID, nil)
		}
	}

	if err := s.itemRepo.BatchUpdateOrder(ctx, req.Items); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "menu",
		ResourceID:   menuID,
	})
	return nil
}
