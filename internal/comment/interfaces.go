package comment

import (
	"context"

	"github.com/sky-flux/cms/internal/model"
)

// CommentRepository handles sfc_site_comments table operations.
type CommentRepository interface {
	List(ctx context.Context, filter ListFilter) ([]CommentRow, int64, error)
	GetByID(ctx context.Context, id string) (*model.Comment, error)
	GetChildren(ctx context.Context, parentID string) ([]*model.Comment, error)
	UpdateStatus(ctx context.Context, id string, status model.CommentStatus) error
	UpdatePinned(ctx context.Context, id string, pinned model.Toggle) error
	Create(ctx context.Context, comment *model.Comment) error
	BatchUpdateStatus(ctx context.Context, ids []string, status model.CommentStatus) (int64, error)
	Delete(ctx context.Context, id string) error
	CountPinnedByPost(ctx context.Context, postID string) (int64, error)
	GetParentChainDepth(ctx context.Context, commentID string) (int, error)
}

// ListFilter holds filtering/pagination options for comment listing.
type ListFilter struct {
	Page    int
	PerPage int
	PostID  string
	Status  string // "pending", "approved", "spam", "trash"
	Query   string // search content or author_name
	Sort    string
}

// CommentRow is a flattened row from the list query with post info joined.
type CommentRow struct {
	model.Comment
	PostTitle string `bun:"post_title" json:"post_title"`
	PostSlug  string `bun:"post_slug" json:"post_slug"`
}
