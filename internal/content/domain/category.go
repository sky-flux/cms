package domain

import (
	"errors"
	"strings"
)

var (
	ErrEmptyCategoryName = errors.New("category name must not be empty")
	ErrEmptyCategorySlug = errors.New("category slug must not be empty")
	ErrCategoryNotFound  = errors.New("category not found")
	ErrCyclicCategory    = errors.New("would create a cycle in category tree")
)

// Category is a tree-structured taxonomy entity.
type Category struct {
	ID       string
	Name     string
	Slug     string
	ParentID string // empty = root
	Path     string // materialized path e.g. "/root-id/child-id"
	Sort     int
}

func NewCategory(name, slug, parentID string) (*Category, error) {
	if strings.TrimSpace(name) == "" {
		return nil, ErrEmptyCategoryName
	}
	if strings.TrimSpace(slug) == "" {
		return nil, ErrEmptyCategorySlug
	}
	return &Category{Name: name, Slug: slug, ParentID: parentID}, nil
}

func (c *Category) HasParent() bool { return c.ParentID != "" }

// WouldCreateCycle returns true if targetParentID is already an ancestor of the node.
// ancestors is the ordered list of ancestor IDs from root → direct parent.
func WouldCreateCycle(targetParentID string, ancestors []string) bool {
	for _, id := range ancestors {
		if id == targetParentID {
			return true
		}
	}
	return false
}
