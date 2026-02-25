package redirect

import (
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// CreateRedirectReq is the request body for POST /redirects.
type CreateRedirectReq struct {
	SourcePath string `json:"source_path" binding:"required,max=500"`
	TargetURL  string `json:"target_url" binding:"required"`
	StatusCode int    `json:"status_code" binding:"omitempty,oneof=301 302"`
}

// UpdateRedirectReq is the request body for PUT /redirects/:id.
type UpdateRedirectReq struct {
	SourcePath *string `json:"source_path" binding:"omitempty,max=500"`
	TargetURL  *string `json:"target_url"`
	StatusCode *int    `json:"status_code" binding:"omitempty,oneof=301 302"`
	IsActive   *bool   `json:"is_active"`
}

// BatchDeleteReq is the request body for DELETE /redirects/batch.
type BatchDeleteReq struct {
	IDs []string `json:"ids" binding:"required,min=1,max=100"`
}

// --- Response DTOs ---

// RedirectResp is the API response for a redirect.
type RedirectResp struct {
	ID         string     `json:"id"`
	SourcePath string     `json:"source_path"`
	TargetURL  string     `json:"target_url"`
	StatusCode int        `json:"status_code"`
	IsActive   bool       `json:"is_active"`
	HitCount   int64      `json:"hit_count"`
	LastHitAt  *time.Time `json:"last_hit_at,omitempty"`
	CreatedBy  *string    `json:"created_by,omitempty"`
	CreatedAt  time.Time  `json:"created_at"`
	UpdatedAt  time.Time  `json:"updated_at"`
}

// ImportResult is the response for CSV import.
type ImportResult struct {
	Imported int      `json:"imported"`
	Skipped  int      `json:"skipped"`
	Errors   []string `json:"errors,omitempty"`
}

// ToRedirectResp converts a model.Redirect to RedirectResp.
func ToRedirectResp(r *model.Redirect) RedirectResp {
	return RedirectResp{
		ID:         r.ID,
		SourcePath: r.SourcePath,
		TargetURL:  r.TargetURL,
		StatusCode: r.StatusCode,
		IsActive:   r.Status == model.RedirectStatusActive,
		HitCount:   r.HitCount,
		LastHitAt:  r.LastHitAt,
		CreatedBy:  r.CreatedBy,
		CreatedAt:  r.CreatedAt,
		UpdatedAt:  r.UpdatedAt,
	}
}
