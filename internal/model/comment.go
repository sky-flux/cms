package model

import (
	"time"

	"github.com/uptrace/bun"
)

type Comment struct {
	bun.BaseModel `bun:"table:sfc_site_comments,alias:cm"`

	ID          string        `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string        `bun:"post_id,notnull,type:uuid" json:"post_id"`
	ParentID    *string       `bun:"parent_id,type:uuid" json:"parent_id,omitempty"`
	UserID      *string       `bun:"user_id,type:uuid" json:"user_id,omitempty"`
	AuthorName  string        `bun:"author_name" json:"author_name,omitempty"`
	AuthorEmail string        `bun:"author_email" json:"author_email,omitempty"`
	AuthorURL   string        `bun:"author_url" json:"author_url,omitempty"`
	AuthorIP    string        `bun:"author_ip,type:inet" json:"-"`
	UserAgent   string        `bun:"user_agent" json:"-"`
	Content     string        `bun:"content,notnull" json:"content"`
	Status      CommentStatus `bun:"status,notnull,type:smallint,default:1" json:"status"`
	IsPinned    bool          `bun:"is_pinned,notnull,default:false" json:"is_pinned"`
	CreatedAt   time.Time     `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time     `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt   *time.Time    `bun:"deleted_at,soft_delete,nullzero" json:"-"`

	Children []*Comment `bun:"rel:has-many,join:id=parent_id" json:"children,omitempty"`
}
