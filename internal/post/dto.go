package post

import (
	"encoding/json"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

type CreatePostReq struct {
	Title             string          `json:"title" binding:"required,max=500"`
	Slug              string          `json:"slug" binding:"omitempty,max=200"`
	Content           string          `json:"content"`
	ContentJSON       json.RawMessage `json:"content_json"`
	Excerpt           string          `json:"excerpt"`
	Status            string          `json:"status" binding:"omitempty,oneof=draft published scheduled"`
	ScheduledAt       *time.Time      `json:"scheduled_at"`
	CoverImageID      *string         `json:"cover_image_id"`
	CategoryIDs       []string        `json:"category_ids"`
	PrimaryCategoryID string          `json:"primary_category_id"`
	TagIDs            []string        `json:"tag_ids"`
	MetaTitle         string          `json:"meta_title" binding:"max=200"`
	MetaDescription   string          `json:"meta_description" binding:"max=500"`
	OGImageURL        string          `json:"og_image_url"`
	ExtraFields       json.RawMessage `json:"extra_fields"`
}

type UpdatePostReq struct {
	Title             *string         `json:"title" binding:"omitempty,max=500"`
	Slug              *string         `json:"slug" binding:"omitempty,max=200"`
	Content           *string         `json:"content"`
	ContentJSON       json.RawMessage `json:"content_json"`
	Excerpt           *string         `json:"excerpt"`
	Status            *string         `json:"status" binding:"omitempty,oneof=draft published scheduled archived"`
	ScheduledAt       *time.Time      `json:"scheduled_at"`
	CoverImageID      *string         `json:"cover_image_id"`
	CategoryIDs       []string        `json:"category_ids"`
	PrimaryCategoryID *string         `json:"primary_category_id"`
	TagIDs            []string        `json:"tag_ids"`
	MetaTitle         *string         `json:"meta_title" binding:"omitempty,max=200"`
	MetaDescription   *string         `json:"meta_description" binding:"omitempty,max=500"`
	OGImageURL        *string         `json:"og_image_url"`
	ExtraFields       json.RawMessage `json:"extra_fields"`
	Version           int             `json:"version" binding:"required,min=1"`
}

type UpsertTranslationReq struct {
	Title           string          `json:"title" binding:"max=500"`
	Excerpt         string          `json:"excerpt"`
	Content         string          `json:"content"`
	ContentJSON     json.RawMessage `json:"content_json"`
	MetaTitle       string          `json:"meta_title" binding:"max=200"`
	MetaDescription string          `json:"meta_description" binding:"max=500"`
	OGImageURL      string          `json:"og_image_url"`
}

// --- Response DTOs ---

type PostResp struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Status      string          `json:"status"`
	Excerpt     string          `json:"excerpt,omitempty"`
	Content     string          `json:"content,omitempty"`
	ContentJSON json.RawMessage `json:"content_json,omitempty"`
	Author      *AuthorResp     `json:"author,omitempty"`
	CoverImage  *CoverImageResp `json:"cover_image,omitempty"`
	Categories  []CategoryBrief `json:"categories,omitempty"`
	Tags        []TagBrief      `json:"tags,omitempty"`
	SEO         *SEOResp        `json:"seo,omitempty"`
	ExtraFields json.RawMessage `json:"extra_fields,omitempty"`
	ViewCount   int64           `json:"view_count"`
	Version     int             `json:"version"`
	ScheduledAt *time.Time      `json:"scheduled_at,omitempty"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type PostListItem struct {
	ID          string          `json:"id"`
	Title       string          `json:"title"`
	Slug        string          `json:"slug"`
	Status      string          `json:"status"`
	Author      *AuthorResp     `json:"author,omitempty"`
	CoverImage  *CoverImageResp `json:"cover_image,omitempty"`
	Categories  []CategoryBrief `json:"categories,omitempty"`
	Tags        []TagBrief      `json:"tags,omitempty"`
	ViewCount   int64           `json:"view_count"`
	PublishedAt *time.Time      `json:"published_at,omitempty"`
	CreatedAt   time.Time       `json:"created_at"`
	UpdatedAt   time.Time       `json:"updated_at"`
}

type AuthorResp struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

type CoverImageResp struct {
	ID            string          `json:"id"`
	URL           string          `json:"url"`
	WebpURL       string          `json:"webp_url,omitempty"`
	ThumbnailURLs json.RawMessage `json:"thumbnail_urls,omitempty"`
}

type CategoryBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type TagBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

type SEOResp struct {
	MetaTitle       string `json:"meta_title,omitempty"`
	MetaDescription string `json:"meta_description,omitempty"`
	OGImageURL      string `json:"og_image_url,omitempty"`
}

type RevisionResp struct {
	ID          string      `json:"id"`
	Version     int         `json:"version"`
	Editor      *AuthorResp `json:"editor,omitempty"`
	DiffSummary string      `json:"diff_summary,omitempty"`
	CreatedAt   time.Time   `json:"created_at"`
}

type TranslationResp struct {
	Locale          string          `json:"locale"`
	Title           string          `json:"title,omitempty"`
	Excerpt         string          `json:"excerpt,omitempty"`
	Content         string          `json:"content,omitempty"`
	ContentJSON     json.RawMessage `json:"content_json,omitempty"`
	MetaTitle       string          `json:"meta_title,omitempty"`
	MetaDescription string          `json:"meta_description,omitempty"`
	OGImageURL      string          `json:"og_image_url,omitempty"`
	CreatedAt       time.Time       `json:"created_at"`
	UpdatedAt       time.Time       `json:"updated_at"`
}

type TranslationListItem struct {
	Locale    string    `json:"locale"`
	Title     string    `json:"title,omitempty"`
	UpdatedAt time.Time `json:"updated_at"`
}

type PreviewTokenResp struct {
	PreviewURL  string    `json:"preview_url,omitempty"`
	Token       string    `json:"token,omitempty"`
	ID          string    `json:"id,omitempty"`
	ExpiresAt   time.Time `json:"expires_at"`
	CreatedAt   time.Time `json:"created_at"`
	ActiveCount int       `json:"active_count,omitempty"`
}

// --- Status helpers ---

var statusMap = map[model.PostStatus]string{
	model.PostStatusDraft:     "draft",
	model.PostStatusScheduled: "scheduled",
	model.PostStatusPublished: "published",
	model.PostStatusArchived:  "archived",
}

var statusReverseMap = map[string]model.PostStatus{
	"draft":     model.PostStatusDraft,
	"scheduled": model.PostStatusScheduled,
	"published": model.PostStatusPublished,
	"archived":  model.PostStatusArchived,
}

func statusString(s model.PostStatus) string {
	if v, ok := statusMap[s]; ok {
		return v
	}
	return "unknown"
}

func parseStatus(s string) (model.PostStatus, bool) {
	v, ok := statusReverseMap[s]
	return v, ok
}

// --- Converters ---

func ToPostResp(p *model.Post) PostResp {
	resp := PostResp{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Status:      statusString(p.Status),
		Excerpt:     p.Excerpt,
		Content:     p.Content,
		ContentJSON: p.ContentJSON,
		ExtraFields: p.ExtraFields,
		ViewCount:   p.ViewCount,
		Version:     p.Version,
		ScheduledAt: p.ScheduledAt,
		PublishedAt:  p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.MetaTitle != "" || p.MetaDesc != "" || p.OGImageURL != "" {
		resp.SEO = &SEOResp{
			MetaTitle:       p.MetaTitle,
			MetaDescription: p.MetaDesc,
			OGImageURL:      p.OGImageURL,
		}
	}

	if p.Author != nil {
		resp.Author = &AuthorResp{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}

	return resp
}

func ToPostListItem(p *model.Post) PostListItem {
	item := PostListItem{
		ID:          p.ID,
		Title:       p.Title,
		Slug:        p.Slug,
		Status:      statusString(p.Status),
		ViewCount:   p.ViewCount,
		PublishedAt: p.PublishedAt,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}

	if p.Author != nil {
		item.Author = &AuthorResp{
			ID:          p.Author.ID,
			DisplayName: p.Author.DisplayName,
			AvatarURL:   p.Author.AvatarURL,
		}
	}

	return item
}

func ToRevisionResp(r *model.PostRevision) RevisionResp {
	return RevisionResp{
		ID:          r.ID,
		Version:     r.Version,
		DiffSummary: r.DiffSummary,
		CreatedAt:   r.CreatedAt,
	}
}

func ToTranslationResp(t *model.PostTranslation) TranslationResp {
	return TranslationResp{
		Locale:          t.Locale,
		Title:           t.Title,
		Excerpt:         t.Excerpt,
		Content:         t.Content,
		ContentJSON:     t.ContentJSON,
		MetaTitle:       t.MetaTitle,
		MetaDescription: t.MetaDesc,
		OGImageURL:      t.OGImageURL,
		CreatedAt:       t.CreatedAt,
		UpdatedAt:       t.UpdatedAt,
	}
}

func ToTranslationListItem(t *model.PostTranslation) TranslationListItem {
	return TranslationListItem{
		Locale:    t.Locale,
		Title:     t.Title,
		UpdatedAt: t.UpdatedAt,
	}
}
