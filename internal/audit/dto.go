package audit

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Filters ---

type ListFilter struct {
	Page         int
	PerPage      int
	ActorID      string
	Action       *model.LogAction
	ResourceType string
	StartDate    *time.Time
	EndDate      *time.Time
}

// --- Joined model ---

// AuditWithActor extends Audit with the actor's display name from sfc_users.
type AuditWithActor struct {
	model.Audit
	ActorDisplayName string `bun:"actor_display_name"`
}

// --- Response DTOs ---

type AuditResp struct {
	ID               string          `json:"id"`
	ActorID          *string         `json:"actor_id,omitempty"`
	ActorEmail       string          `json:"actor_email,omitempty"`
	ActorDisplayName string          `json:"actor_display_name,omitempty"`
	Action           model.LogAction `json:"action"`
	ResourceType     string          `json:"resource_type"`
	ResourceID       string          `json:"resource_id,omitempty"`
	ResourceSnapshot json.RawMessage `json:"resource_snapshot,omitempty"`
	IPAddress        string          `json:"ip_address,omitempty"`
	UserAgent        string          `json:"user_agent,omitempty"`
	CreatedAt        time.Time       `json:"created_at"`
}

func ToAuditResp(a *AuditWithActor) AuditResp {
	return AuditResp{
		ID:               a.ID,
		ActorID:          a.ActorID,
		ActorEmail:       a.ActorEmail,
		ActorDisplayName: a.ActorDisplayName,
		Action:           a.Action,
		ResourceType:     a.ResourceType,
		ResourceID:       a.ResourceID,
		ResourceSnapshot: a.ResourceSnapshot,
		IPAddress:        a.IPAddress,
		UserAgent:        a.UserAgent,
		CreatedAt:        a.CreatedAt,
	}
}

func ToAuditRespList(items []AuditWithActor) []AuditResp {
	out := make([]AuditResp, len(items))
	for i := range items {
		out[i] = ToAuditResp(&items[i])
	}
	return out
}
