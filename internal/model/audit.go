package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

type Audit struct {
	bun.BaseModel `bun:"table:sfc_site_audits,alias:a"`

	ID               string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	ActorID          *string         `bun:"actor_id,type:uuid" json:"actor_id,omitempty"`
	ActorEmail       string          `bun:"actor_email" json:"actor_email,omitempty"`
	Action           LogAction       `bun:"action,notnull,type:smallint" json:"action"`
	ResourceType     string          `bun:"resource_type,notnull" json:"resource_type"`
	ResourceID       string          `bun:"resource_id" json:"resource_id,omitempty"`
	ResourceSnapshot json.RawMessage `bun:"resource_snapshot,type:jsonb" json:"resource_snapshot,omitempty"`
	IPAddress        string          `bun:"ip_address,type:inet" json:"ip_address,omitempty"`
	UserAgent        string          `bun:"user_agent" json:"user_agent,omitempty"`
	CreatedAt        time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}
