package posttype

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type CreatePostTypeReq struct {
	Name        string          `json:"name" binding:"required,max=100"`
	Slug        string          `json:"slug" binding:"required,max=100"`
	Description string          `json:"description"`
	Fields      json.RawMessage `json:"fields"`
}

type UpdatePostTypeReq struct {
	Name        *string          `json:"name" binding:"omitempty,max=100"`
	Slug        *string          `json:"slug" binding:"omitempty,max=100"`
	Description *string          `json:"description"`
	Fields      *json.RawMessage `json:"fields"`
}

// --- Response DTOs ---

type PostTypeResp struct {
	ID          string          `json:"id"`
	Name        string          `json:"name"`
	Slug        string          `json:"slug"`
	Description string          `json:"description,omitempty"`
	Fields      json.RawMessage `json:"fields"`
	FieldCount  int             `json:"field_count"`
	BuiltIn     model.Toggle    `json:"built_in"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

func ToPostTypeResp(pt *model.PostType) PostTypeResp {
	return PostTypeResp{
		ID:          pt.ID,
		Name:        pt.Name,
		Slug:        pt.Slug,
		Description: pt.Description,
		Fields:      pt.Fields,
		FieldCount:  countFields(pt.Fields),
		BuiltIn:     pt.BuiltIn,
		CreatedAt:   pt.CreatedAt,
		UpdatedAt:   pt.UpdatedAt,
	}
}

func ToPostTypeRespList(pts []model.PostType) []PostTypeResp {
	out := make([]PostTypeResp, len(pts))
	for i := range pts {
		out[i] = ToPostTypeResp(&pts[i])
	}
	return out
}

// countFields returns the number of elements in a JSON array, or 0 if invalid.
func countFields(raw json.RawMessage) int {
	if len(raw) == 0 {
		return 0
	}
	var arr []json.RawMessage
	if err := json.Unmarshal(raw, &arr); err != nil {
		return 0
	}
	return len(arr)
}
