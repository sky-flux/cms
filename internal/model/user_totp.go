package model

import (
	"context"
	"time"

	"github.com/uptrace/bun"
)

type UserTOTP struct {
	bun.BaseModel `bun:"table:sfc_user_totp,alias:totp"`

	ID              string     `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	UserID          string     `bun:"user_id,notnull,unique,type:uuid" json:"user_id"`
	SecretEncrypted string     `bun:"secret_encrypted,notnull" json:"-"`
	BackupCodesHash []string   `bun:"backup_codes_hash,type:text[],array" json:"-"`
	Enabled         Toggle     `bun:"enabled,notnull,type:smallint,default:1" json:"enabled"`
	VerifiedAt      *time.Time `bun:"verified_at" json:"verified_at,omitempty"`
	CreatedAt       time.Time  `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt       time.Time  `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// BeforeAppendModel implements bun.BeforeAppendModelHook.
func (ut *UserTOTP) BeforeAppendModel(ctx context.Context, query bun.Query) error {
	SetTimestamps(&ut.CreatedAt, &ut.UpdatedAt, query)
	return nil
}
