package comment

import (
	"crypto/md5"
	"fmt"
	"strings"
	"time"

	"github.com/sky-flux/cms/internal/model"
)

// --- Request DTOs ---

// UpdateStatusReq is the request body for PUT /comments/:id/status.
type UpdateStatusReq struct {
	Status string `json:"status" binding:"required,oneof=pending approved spam trash"`
}

// TogglePinReq is the request body for PUT /comments/:id/pin.
type TogglePinReq struct {
	Pinned bool `json:"is_pinned"`
}

// ReplyReq is the request body for POST /comments/:id/reply.
type ReplyReq struct {
	Content string `json:"content" binding:"required,max=5000"`
}

// BatchStatusReq is the request body for PUT /comments/batch-status.
type BatchStatusReq struct {
	IDs    []string `json:"ids" binding:"required,min=1,max=100"`
	Status string   `json:"status" binding:"required,oneof=approved spam trash"`
}

// --- Response DTOs ---

// CommentResp is the API response for a comment.
type CommentResp struct {
	ID          string         `json:"id"`
	PostID      string         `json:"post_id"`
	PostTitle   string         `json:"post_title,omitempty"`
	PostSlug    string         `json:"post_slug,omitempty"`
	ParentID    *string        `json:"parent_id,omitempty"`
	UserID      *string        `json:"user_id,omitempty"`
	AuthorName  string         `json:"author_name"`
	AuthorEmail string         `json:"author_email,omitempty"`
	AuthorURL   string         `json:"author_url,omitempty"`
	AuthorIP    string         `json:"author_ip,omitempty"`
	UserAgent   string         `json:"user_agent,omitempty"`
	GravatarURL string         `json:"gravatar_url"`
	Content     string         `json:"content"`
	Status      string         `json:"status"`
	IsPinned    bool           `json:"is_pinned"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	Children    []*CommentResp `json:"children,omitempty"`
}

// GravatarURL computes a gravatar URL from an email address.
func GravatarURL(email string) string {
	trimmed := strings.TrimSpace(strings.ToLower(email))
	hash := md5.Sum([]byte(trimmed))
	return fmt.Sprintf("https://www.gravatar.com/avatar/%x?d=mp", hash)
}

// CommentStatusToString converts CommentStatus enum to string.
func CommentStatusToString(s model.CommentStatus) string {
	switch s {
	case model.CommentStatusPending:
		return "pending"
	case model.CommentStatusApproved:
		return "approved"
	case model.CommentStatusSpam:
		return "spam"
	case model.CommentStatusTrash:
		return "trash"
	default:
		return "unknown"
	}
}

// StringToCommentStatus converts string to CommentStatus enum.
func StringToCommentStatus(s string) model.CommentStatus {
	switch s {
	case "pending":
		return model.CommentStatusPending
	case "approved":
		return model.CommentStatusApproved
	case "spam":
		return model.CommentStatusSpam
	case "trash":
		return model.CommentStatusTrash
	default:
		return model.CommentStatusPending
	}
}

// ToCommentResp converts a CommentRow to CommentResp (for list).
func ToCommentResp(row *CommentRow) CommentResp {
	return CommentResp{
		ID:          row.ID,
		PostID:      row.PostID,
		PostTitle:   row.PostTitle,
		PostSlug:    row.PostSlug,
		ParentID:    row.ParentID,
		UserID:      row.UserID,
		AuthorName:  row.AuthorName,
		AuthorEmail: row.AuthorEmail,
		AuthorURL:   row.AuthorURL,
		AuthorIP:    row.AuthorIP,
		UserAgent:   row.UserAgent,
		GravatarURL: GravatarURL(row.AuthorEmail),
		Content:     row.Content,
		Status:      CommentStatusToString(row.Status),
		IsPinned:    row.Pinned == model.ToggleYes,
		CreatedAt:   row.CreatedAt,
		UpdatedAt:   row.UpdatedAt,
	}
}

// ToCommentDetailResp converts a model.Comment to CommentResp (for detail/reply).
func ToCommentDetailResp(c *model.Comment) CommentResp {
	resp := CommentResp{
		ID:          c.ID,
		PostID:      c.PostID,
		ParentID:    c.ParentID,
		UserID:      c.UserID,
		AuthorName:  c.AuthorName,
		AuthorEmail: c.AuthorEmail,
		AuthorURL:   c.AuthorURL,
		AuthorIP:    c.AuthorIP,
		UserAgent:   c.UserAgent,
		GravatarURL: GravatarURL(c.AuthorEmail),
		Content:     c.Content,
		Status:      CommentStatusToString(c.Status),
		IsPinned:    c.Pinned == model.ToggleYes,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}
	if c.Children != nil {
		resp.Children = make([]*CommentResp, len(c.Children))
		for i, child := range c.Children {
			r := ToCommentDetailResp(child)
			resp.Children[i] = &r
		}
	}
	return resp
}
