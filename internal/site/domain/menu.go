package domain

import (
	"errors"
	"strings"
	"time"
)

var (
	ErrEmptyMenuName      = errors.New("menu name must not be empty")
	ErrEmptyMenuSlug      = errors.New("menu slug must not be empty")
	ErrEmptyMenuItemLabel = errors.New("menu item label must not be empty")
	ErrEmptyMenuItemURL   = errors.New("menu item URL must not be empty")
)

// Menu is an aggregate root representing a navigation menu (e.g., "Main Nav", "Footer").
type Menu struct {
	ID        string
	Name      string
	Slug      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMenu validates inputs and constructs a Menu.
func NewMenu(name, slug string) (*Menu, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyMenuName
	}
	if strings.TrimSpace(slug) == "" {
		return nil, ErrEmptyMenuSlug
	}
	return &Menu{
		Name: strings.TrimSpace(name),
		Slug: strings.TrimSpace(slug),
	}, nil
}

// MenuItem is a value object: one link within a Menu.
// ParentID is empty for top-level items; max nesting depth is 3.
type MenuItem struct {
	ID        string
	MenuID    string
	ParentID  string // empty = top level
	Label     string
	URL       string
	Order     int
	CreatedAt time.Time
	UpdatedAt time.Time
}

// NewMenuItem validates inputs and constructs a MenuItem.
func NewMenuItem(menuID, label, url string, order int, parentID string) (*MenuItem, error) {
	if strings.TrimSpace(label) == "" {
		return nil, ErrEmptyMenuItemLabel
	}
	if strings.TrimSpace(url) == "" {
		return nil, ErrEmptyMenuItemURL
	}
	return &MenuItem{
		MenuID:   menuID,
		Label:    strings.TrimSpace(label),
		URL:      strings.TrimSpace(url),
		Order:    order,
		ParentID: parentID,
	}, nil
}

// IsTopLevel returns true when the item has no parent.
func (mi *MenuItem) IsTopLevel() bool {
	return mi.ParentID == ""
}
