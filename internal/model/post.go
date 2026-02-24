package model

import (
	"encoding/json"
	"time"

	"github.com/uptrace/bun"
)

// Post maps to sfc_site_posts in site_{slug} schema.
type Post struct {
	bun.BaseModel `bun:"table:sfc_site_posts,alias:p"`

	ID             string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	AuthorID       string          `bun:"author_id,notnull,type:uuid" json:"author_id"`
	CoverImageID   *string         `bun:"cover_image_id,type:uuid" json:"cover_image_id,omitempty"`
	PostType       string          `bun:"post_type,notnull,default:'article'" json:"post_type"`
	Status         PostStatus      `bun:"status,notnull,type:smallint,default:1" json:"status"`
	Title          string          `bun:"title,notnull" json:"title"`
	Slug           string          `bun:"slug,notnull" json:"slug"`
	Excerpt        string          `bun:"excerpt" json:"excerpt,omitempty"`
	Content        string          `bun:"content" json:"content,omitempty"`
	ContentJSON    json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	MetaTitle      string          `bun:"meta_title" json:"meta_title,omitempty"`
	MetaDesc       string          `bun:"meta_description" json:"meta_description,omitempty"`
	OGImageURL     string          `bun:"og_image_url" json:"og_image_url,omitempty"`
	ExtraFields    json.RawMessage `bun:"extra_fields,type:jsonb,default:'{}'" json:"extra_fields,omitempty"`
	ViewCount      int64           `bun:"view_count,notnull,default:0" json:"view_count"`
	Version        int             `bun:"version,notnull,default:1" json:"version"`
	PublishedAt    *time.Time      `bun:"published_at" json:"published_at,omitempty"`
	ScheduledAt    *time.Time      `bun:"scheduled_at" json:"scheduled_at,omitempty"`
	CreatedAt      time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt      time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
	DeletedAt      *time.Time      `bun:"deleted_at,soft_delete,nullzero" json:"-"`

	Author *User `bun:"rel:belongs-to,join:author_id=id" json:"author,omitempty"`
}

// PostTranslation maps to sfc_site_post_translations.
type PostTranslation struct {
	bun.BaseModel `bun:"table:sfc_site_post_translations,alias:pt"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string          `bun:"post_id,notnull,type:uuid" json:"post_id"`
	Locale      string          `bun:"locale,notnull" json:"locale"`
	Title       string          `bun:"title" json:"title,omitempty"`
	Excerpt     string          `bun:"excerpt" json:"excerpt,omitempty"`
	Content     string          `bun:"content" json:"content,omitempty"`
	ContentJSON json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	MetaTitle   string          `bun:"meta_title" json:"meta_title,omitempty"`
	MetaDesc    string          `bun:"meta_description" json:"meta_description,omitempty"`
	OGImageURL  string          `bun:"og_image_url" json:"og_image_url,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
	UpdatedAt   time.Time       `bun:"updated_at,notnull,default:current_timestamp" json:"updated_at"`
}

// PostRevision maps to sfc_site_post_revisions.
type PostRevision struct {
	bun.BaseModel `bun:"table:sfc_site_post_revisions,alias:pr"`

	ID          string          `bun:"id,pk,type:uuid,default:uuidv7()" json:"id"`
	PostID      string          `bun:"post_id,notnull,type:uuid" json:"post_id"`
	EditorID    string          `bun:"editor_id,notnull,type:uuid" json:"editor_id"`
	Version     int             `bun:"version,notnull" json:"version"`
	Title       string          `bun:"title" json:"title,omitempty"`
	Content     string          `bun:"content" json:"content,omitempty"`
	ContentJSON json.RawMessage `bun:"content_json,type:jsonb" json:"content_json,omitempty"`
	DiffSummary string          `bun:"diff_summary" json:"diff_summary,omitempty"`
	CreatedAt   time.Time       `bun:"created_at,notnull,default:current_timestamp" json:"created_at"`
}

// PostCategoryMap maps to sfc_site_post_category_map (many-to-many).
type PostCategoryMap struct {
	bun.BaseModel `bun:"table:sfc_site_post_category_map"`

	PostID     string `bun:"post_id,pk,type:uuid" json:"post_id"`
	CategoryID string `bun:"category_id,pk,type:uuid" json:"category_id"`
	IsPrimary  bool   `bun:"is_primary,notnull,default:false" json:"is_primary"`
}

// PostTagMap maps to sfc_site_post_tag_map (many-to-many).
type PostTagMap struct {
	bun.BaseModel `bun:"table:sfc_site_post_tag_map"`

	PostID string `bun:"post_id,pk,type:uuid" json:"post_id"`
	TagID  string `bun:"tag_id,pk,type:uuid" json:"tag_id"`
}
