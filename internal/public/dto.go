package public

import (
	"encoding/json"
	"time"
)

// PostListItem is the public response for a post in a list.
type PostListItem struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Excerpt     string           `json:"excerpt,omitempty"`
	Author      *AuthorBrief     `json:"author,omitempty"`
	CoverImage  *CoverImageBrief `json:"cover_image,omitempty"`
	Categories  []RefBrief       `json:"categories,omitempty"`
	Tags        []RefBrief       `json:"tags,omitempty"`
	ViewCount   int64            `json:"view_count"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
}

// PostDetail is the public response for a single post.
type PostDetail struct {
	ID          string           `json:"id"`
	Title       string           `json:"title"`
	Slug        string           `json:"slug"`
	Content     string           `json:"content,omitempty"`
	ContentJSON json.RawMessage  `json:"content_json,omitempty"`
	Excerpt     string           `json:"excerpt,omitempty"`
	Author      *AuthorBrief     `json:"author,omitempty"`
	CoverImage  *CoverImageBrief `json:"cover_image,omitempty"`
	Categories  []RefBrief       `json:"categories,omitempty"`
	Tags        []RefBrief       `json:"tags,omitempty"`
	SEO         *SEOFields       `json:"seo,omitempty"`
	ExtraFields json.RawMessage  `json:"extra_fields,omitempty"`
	ViewCount   int64            `json:"view_count"`
	PublishedAt *time.Time       `json:"published_at,omitempty"`
}

// AuthorBrief is the sanitized author info for public API.
type AuthorBrief struct {
	ID          string `json:"id"`
	DisplayName string `json:"display_name"`
	AvatarURL   string `json:"avatar_url,omitempty"`
}

// CoverImageBrief is the sanitized cover image for public API.
type CoverImageBrief struct {
	ID  string `json:"id"`
	URL string `json:"url"`
}

// RefBrief is a lightweight category/tag reference.
type RefBrief struct {
	ID   string `json:"id"`
	Name string `json:"name"`
	Slug string `json:"slug"`
}

// SEOFields groups SEO-related post fields.
type SEOFields struct {
	MetaTitle  string `json:"meta_title,omitempty"`
	MetaDesc   string `json:"meta_description,omitempty"`
	OGImageURL string `json:"og_image_url,omitempty"`
}

// CategoryNode is a public category with post count and children.
type CategoryNode struct {
	ID        string         `json:"id"`
	Name      string         `json:"name"`
	Slug      string         `json:"slug"`
	Path      string         `json:"path"`
	PostCount int64          `json:"post_count"`
	Children  []CategoryNode `json:"children"`
}

// TagItem is a public tag with post count.
type TagItem struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Slug      string `json:"slug"`
	PostCount int64  `json:"post_count"`
}

// PublicComment is a sanitized comment for public API (no email/IP/UA).
type PublicComment struct {
	ID         string          `json:"id"`
	ParentID   *string         `json:"parent_id,omitempty"`
	AuthorName string          `json:"author_name"`
	AuthorURL  string          `json:"author_url,omitempty"`
	Content    string          `json:"content"`
	IsPinned   bool            `json:"is_pinned"`
	CreatedAt  time.Time       `json:"created_at"`
	Replies    []PublicComment `json:"replies"`
}

// CommentListResult is the result of listing public comments.
type CommentListResult struct {
	CommentCount int64           `json:"comment_count"`
	Comments     []PublicComment `json:"comments"`
	Total        int64           `json:"-"`
	Page         int             `json:"-"`
	PerPage      int             `json:"-"`
}

// CreateCommentReq is the request body for public comment submission.
type CreateCommentReq struct {
	ParentID    *string `json:"parent_id"`
	AuthorName  string  `json:"author_name"`
	AuthorEmail string  `json:"author_email"`
	AuthorURL   string  `json:"author_url"`
	Content     string  `json:"content" binding:"required,min=1,max=10000"`
	Honeypot    string  `json:"honeypot"`
}

// CreateCommentResp is the response after submitting a comment.
type CreateCommentResp struct {
	ID      string `json:"id"`
	Status  string `json:"status"`
	Message string `json:"message"`
}

// PublicMenuItem is a sanitized menu item for public API.
type PublicMenuItem struct {
	ID       string           `json:"id"`
	Label    string           `json:"label"`
	URL      string           `json:"url,omitempty"`
	Target   string           `json:"target"`
	Icon     string           `json:"icon,omitempty"`
	CSSClass string           `json:"css_class,omitempty"`
	Children []PublicMenuItem  `json:"children"`
}

// PublicMenu is the public menu response.
type PublicMenu struct {
	ID       string           `json:"id"`
	Name     string           `json:"name"`
	Slug     string           `json:"slug"`
	Location string           `json:"location,omitempty"`
	Items    []PublicMenuItem  `json:"items"`
}

// PreviewResp is the response for a preview token consumption.
type PreviewResp struct {
	PostDetail
	IsPreview        bool       `json:"is_preview"`
	PreviewExpiresAt *time.Time `json:"preview_expires_at,omitempty"`
}

// SearchResultItem is a single search result.
type SearchResultItem struct {
	ID          string       `json:"id"`
	Title       string       `json:"title"`
	Slug        string       `json:"slug"`
	Excerpt     string       `json:"excerpt,omitempty"`
	Author      *AuthorBrief `json:"author,omitempty"`
	Categories  []RefBrief   `json:"categories,omitempty"`
	Tags        []RefBrief   `json:"tags,omitempty"`
	PublishedAt *time.Time   `json:"published_at,omitempty"`
}
