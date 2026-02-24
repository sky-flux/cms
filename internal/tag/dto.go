package tag

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// CreateTagReq is the request body for creating a tag.
type CreateTagReq struct {
	Name string `json:"name" binding:"required,max=100"`
	Slug string `json:"slug" binding:"required,max=200"`
}

// UpdateTagReq is the request body for updating a tag.
type UpdateTagReq struct {
	Name *string `json:"name" binding:"omitempty,max=100"`
	Slug *string `json:"slug" binding:"omitempty,max=200"`
}

// --- Response DTOs ---

// TagResp is the API response representation of a tag.
type TagResp struct {
	ID        string    `json:"id"`
	Name      string    `json:"name"`
	Slug      string    `json:"slug"`
	PostCount int64     `json:"post_count"`
	CreatedAt time.Time `json:"created_at"`
}

// ToTagResp converts a model.Tag and post count to a TagResp.
func ToTagResp(t *model.Tag, postCount int64) TagResp {
	return TagResp{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		PostCount: postCount,
		CreatedAt: t.CreatedAt,
	}
}

// ToTagRespList converts a slice of model.Tag to a slice of TagResp
// with zero post counts (for list operations where counts are fetched separately).
func ToTagRespList(tags []model.Tag, countFn func(string) int64) []TagResp {
	out := make([]TagResp, len(tags))
	for i := range tags {
		out[i] = ToTagResp(&tags[i], countFn(tags[i].ID))
	}
	return out
}
