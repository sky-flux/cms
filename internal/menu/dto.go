package menu

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// CreateMenuReq is the request body for POST /menus.
type CreateMenuReq struct {
	Name        string `json:"name" binding:"required,max=100"`
	Slug        string `json:"slug" binding:"required,max=100"`
	Location    string `json:"location" binding:"omitempty,max=50"`
	Description string `json:"description" binding:"omitempty"`
}

// UpdateMenuReq is the request body for PUT /menus/:id.
type UpdateMenuReq struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Slug        *string `json:"slug" binding:"omitempty,max=100"`
	Location    *string `json:"location" binding:"omitempty,max=50"`
	Description *string `json:"description"`
}

// CreateMenuItemReq is the request body for POST /menus/:id/items.
type CreateMenuItemReq struct {
	ParentID    *string `json:"parent_id"`
	Label       string  `json:"label" binding:"required,max=200"`
	URL         string  `json:"url" binding:"omitempty"`
	Target      string  `json:"target" binding:"omitempty,oneof=_self _blank"`
	Type        string  `json:"type" binding:"required,oneof=custom post category tag page"`
	ReferenceID *string `json:"reference_id"`
	SortOrder   int     `json:"sort_order"`
	Icon        string  `json:"icon" binding:"omitempty,max=50"`
	CSSClass    string  `json:"css_class" binding:"omitempty,max=100"`
}

// UpdateMenuItemReq is the request body for PUT /menus/:id/items/:item_id.
type UpdateMenuItemReq struct {
	ParentID    *string `json:"parent_id"`
	Label       *string `json:"label" binding:"omitempty,max=200"`
	URL         *string `json:"url"`
	Target      *string `json:"target" binding:"omitempty,oneof=_self _blank"`
	Type        *string `json:"type" binding:"omitempty,oneof=custom post category tag page"`
	ReferenceID *string `json:"reference_id"`
	SortOrder   *int    `json:"sort_order"`
	Icon        *string `json:"icon" binding:"omitempty,max=50"`
	CSSClass    *string `json:"css_class" binding:"omitempty,max=100"`
	Status      *string `json:"status" binding:"omitempty,oneof=active hidden"`
}

// ReorderReq is the request body for PUT /menus/:id/items/reorder.
type ReorderReq struct {
	Items []ReorderItem `json:"items" binding:"required,min=1"`
}

// --- Response DTOs ---

// MenuResp is the API response for a menu (list view).
type MenuResp struct {
	ID          string    `json:"id"`
	Name        string    `json:"name"`
	Slug        string    `json:"slug"`
	Location    string    `json:"location,omitempty"`
	Description string    `json:"description,omitempty"`
	ItemCount   int64     `json:"item_count"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// MenuDetailResp is the API response for a menu with nested items.
type MenuDetailResp struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Location    string          `json:"location,omitempty"`
	Description string          `json:"description,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Items       []*MenuItemResp `json:"items"`
}

// MenuItemResp is the API response for a menu item.
type MenuItemResp struct {
	ID          string          `json:"id"`
	MenuID      string          `json:"menu_id"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Label       string          `json:"label"`
	URL         string          `json:"url,omitempty"`
	Target      string          `json:"target"`
	Type        string          `json:"type"`
	ReferenceID *string         `json:"reference_id,omitempty"`
	SortOrder   int             `json:"sort_order"`
	Status      string          `json:"status"`
	Icon        string          `json:"icon,omitempty"`
	CSSClass    string          `json:"css_class,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
	Children    []*MenuItemResp `json:"children,omitempty"`
}

// MenuItemTypeToString converts MenuItemType enum to string.
func MenuItemTypeToString(t model.MenuItemType) string {
	switch t {
	case model.MenuItemTypeCustom:
		return "custom"
	case model.MenuItemTypePost:
		return "post"
	case model.MenuItemTypeCategory:
		return "category"
	case model.MenuItemTypeTag:
		return "tag"
	case model.MenuItemTypePage:
		return "page"
	default:
		return "custom"
	}
}

// StringToMenuItemType converts string to MenuItemType enum.
func StringToMenuItemType(s string) model.MenuItemType {
	switch s {
	case "post":
		return model.MenuItemTypePost
	case "category":
		return model.MenuItemTypeCategory
	case "tag":
		return model.MenuItemTypeTag
	case "page":
		return model.MenuItemTypePage
	default:
		return model.MenuItemTypeCustom
	}
}

// MenuItemStatusToString converts MenuItemStatus enum to string.
func MenuItemStatusToString(s model.MenuItemStatus) string {
	if s == model.MenuItemStatusHidden {
		return "hidden"
	}
	return "active"
}

// StringToMenuItemStatus converts string to MenuItemStatus enum.
func StringToMenuItemStatus(s string) model.MenuItemStatus {
	if s == "hidden" {
		return model.MenuItemStatusHidden
	}
	return model.MenuItemStatusActive
}

// ToMenuResp converts a SiteMenu + item count to MenuResp.
func ToMenuResp(m *model.SiteMenu, itemCount int64) MenuResp {
	return MenuResp{
		ID:          m.ID,
		Name:        m.Name,
		Slug:        m.Slug,
		Location:    m.Location,
		Description: m.Description,
		ItemCount:   itemCount,
		CreatedAt:   m.CreatedAt,
		UpdatedAt:   m.UpdatedAt,
	}
}

// ToMenuItemResp converts a SiteMenuItem to MenuItemResp (recursive for children).
func ToMenuItemResp(item *model.SiteMenuItem) *MenuItemResp {
	resp := &MenuItemResp{
		ID:          item.ID,
		MenuID:      item.MenuID,
		ParentID:    item.ParentID,
		Label:       item.Label,
		URL:         item.URL,
		Target:      item.Target,
		Type:        MenuItemTypeToString(item.Type),
		ReferenceID: item.ReferenceID,
		SortOrder:   item.SortOrder,
		Status:      MenuItemStatusToString(item.Status),
		Icon:        item.Icon,
		CSSClass:    item.CSSClass,
		CreatedAt:   item.CreatedAt,
		UpdatedAt:   item.UpdatedAt,
	}
	if item.Children != nil {
		resp.Children = make([]*MenuItemResp, len(item.Children))
		for i, child := range item.Children {
			resp.Children[i] = ToMenuItemResp(child)
		}
	}
	return resp
}
