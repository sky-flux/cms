package delivery

import (
	"context"
	"net/http"
	"time"

	"github.com/danielgtaylor/huma/v2"
	"github.com/sky-flux/cms/internal/platform/domain"
)

// AuditLister is the minimal port the audit handler needs.
type AuditLister interface {
	List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error)
}

// AuditHandler handles audit log endpoints.
type AuditHandler struct {
	lister AuditLister
}

// NewAuditHandler creates an AuditHandler.
func NewAuditHandler(lister AuditLister) *AuditHandler {
	return &AuditHandler{lister: lister}
}

// RegisterAuditRoutes wires audit endpoints onto the Huma API.
func RegisterAuditRoutes(api huma.API, h *AuditHandler) {
	huma.Register(api, huma.Operation{
		OperationID: "admin-list-audit",
		Method:      http.MethodGet,
		Path:        "/api/v1/admin/audit",
		Summary:     "List audit log entries",
		Tags:        []string{"Audit"},
	}, h.ListAudit)
}

// --- Request / Response types ---

// ListAuditInput defines query parameters for listing audit entries.
type ListAuditInput struct {
	Page     int    `query:"page" default:"1" minimum:"1"`
	PerPage  int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
	UserID   string `query:"user_id"`
	Resource string `query:"resource"`
	Action   int    `query:"action"` // 0 = all actions; positive values filter by AuditAction
	Start    string `query:"start_date"` // RFC3339
	End      string `query:"end_date"`   // RFC3339
}

// AuditEntryResp is the JSON representation of an audit log entry.
type AuditEntryResp struct {
	ID         string    `json:"id"`
	UserID     string    `json:"user_id"`
	Action     int8      `json:"action"`
	Resource   string    `json:"resource"`
	ResourceID string    `json:"resource_id"`
	IP         string    `json:"ip"`
	UserAgent  string    `json:"user_agent"`
	CreatedAt  time.Time `json:"created_at"`
}

// ListAuditOutput is the response body for listing audit entries.
type ListAuditOutput struct {
	Body struct {
		Items []AuditEntryResp `json:"items"`
		Total int64            `json:"total"`
		Page  int              `json:"page"`
	}
}

// ListAudit handles GET /api/v1/admin/audit.
func (h *AuditHandler) ListAudit(ctx context.Context, in *ListAuditInput) (*ListAuditOutput, error) {
	f := domain.AuditFilter{
		Page:     in.Page,
		PerPage:  in.PerPage,
		UserID:   in.UserID,
		Resource: in.Resource,
	}
	if in.Action > 0 {
		a := domain.AuditAction(in.Action)
		f.Action = &a
	}
	if in.Start != "" {
		t, err := time.Parse(time.RFC3339, in.Start)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "invalid start_date: use RFC3339 format")
		}
		f.StartDate = &t
	}
	if in.End != "" {
		t, err := time.Parse(time.RFC3339, in.End)
		if err != nil {
			return nil, huma.NewError(http.StatusBadRequest, "invalid end_date: use RFC3339 format")
		}
		f.EndDate = &t
	}

	entries, total, err := h.lister.List(ctx, f)
	if err != nil {
		return nil, huma.NewError(http.StatusInternalServerError, "failed to list audit entries")
	}

	out := &ListAuditOutput{}
	out.Body.Total = total
	out.Body.Page = in.Page
	for _, e := range entries {
		out.Body.Items = append(out.Body.Items, AuditEntryResp{
			ID:         e.ID,
			UserID:     e.UserID,
			Action:     int8(e.Action),
			Resource:   e.Resource,
			ResourceID: e.ResourceID,
			IP:         e.IP,
			UserAgent:  e.UserAgent,
			CreatedAt:  e.CreatedAt,
		})
	}
	if out.Body.Items == nil {
		out.Body.Items = []AuditEntryResp{}
	}
	return out, nil
}
