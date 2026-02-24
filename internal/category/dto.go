package category

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// CreateCategoryReq is the request body for creating a category.
type CreateCategoryReq struct {
	Name        string  `json:"name" binding:"required,max=100"`
	Slug        string  `json:"slug" binding:"required,max=200"`
	ParentID    *string `json:"parent_id"`
	Description string  `json:"description"`
	SortOrder   int     `json:"sort_order"`
}

// UpdateCategoryReq is the request body for updating a category.
type UpdateCategoryReq struct {
	Name        *string `json:"name" binding:"omitempty,max=100"`
	Slug        *string `json:"slug" binding:"omitempty,max=200"`
	ParentID    *string `json:"parent_id"`
	Description *string `json:"description"`
	SortOrder   *int    `json:"sort_order"`
}

// ReorderReq is the request body for batch-reordering categories.
type ReorderReq struct {
	Orders []SortOrderItem `json:"orders" binding:"required,min=1"`
}

// CategoryResp is the API response representation of a category.
type CategoryResp struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Path        string          `json:"path"`
	ParentID    *string         `json:"parent_id,omitempty"`
	Description string          `json:"description,omitempty"`
	PostCount   int64           `json:"post_count"`
	SortOrder   int             `json:"sort_order"`
	Children    []*CategoryResp `json:"children,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

// ToCategoryResp converts a model.Category and post count to a CategoryResp.
func ToCategoryResp(c *model.Category, postCount int64) CategoryResp {
	return CategoryResp{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Path:        c.Path,
		ParentID:    c.ParentID,
		Description: c.Description,
		PostCount:   postCount,
		SortOrder:   c.SortOrder,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
}
