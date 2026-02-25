package comment

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/mail"
)

// Service handles comment business logic.
type Service struct {
	repo   CommentRepository
	audit  audit.Logger
	mailer mail.Sender
}

// NewService creates a new comment service.
func NewService(repo CommentRepository, audit audit.Logger, mailer mail.Sender) *Service {
	return &Service{repo: repo, audit: audit, mailer: mailer}
}

// List returns a paginated list of comments.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]CommentResp, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 || filter.PerPage > 100 {
		filter.PerPage = 20
	}

	rows, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, err
	}

	out := make([]CommentResp, len(rows))
	for i := range rows {
		out[i] = ToCommentResp(&rows[i])
	}
	return out, total, nil
}

// GetComment returns a comment with its direct children.
func (s *Service) GetComment(ctx context.Context, id string) (*CommentResp, error) {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return nil, err
	}
	comment.Children = children

	resp := ToCommentDetailResp(comment)
	return &resp, nil
}

// UpdateStatus changes the status of a comment.
func (s *Service) UpdateStatus(ctx context.Context, id string, req *UpdateStatusReq) error {
	// Verify comment exists
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	status := StringToCommentStatus(req.Status)
	if err := s.repo.UpdateStatus(ctx, id, status); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}

// TogglePin toggles the pinned status of a top-level comment.
func (s *Service) TogglePin(ctx context.Context, id string, req *TogglePinReq) error {
	comment, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Only top-level comments can be pinned
	if comment.ParentID != nil {
		return apperror.Validation("only top-level comments can be pinned", nil)
	}

	pinned := model.ToggleNo
	if req.Pinned {
		pinned = model.ToggleYes

		// Check max 3 pinned per post
		count, err := s.repo.CountPinnedByPost(ctx, comment.PostID)
		if err != nil {
			return err
		}
		// If already pinned, don't count it again
		if comment.Pinned != model.ToggleYes && count >= 3 {
			return apperror.Validation("maximum 3 pinned comments per post", nil)
		}
	}

	if err := s.repo.UpdatePinned(ctx, id, pinned); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}

// Reply creates an admin reply to a comment.
func (s *Service) Reply(ctx context.Context, id string, req *ReplyReq, userID, userName, userEmail string) (*CommentResp, error) {
	parent, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Check nesting depth (max 3 levels)
	depth, err := s.repo.GetParentChainDepth(ctx, id)
	if err != nil {
		return nil, err
	}
	if depth >= 2 { // parent is already at depth 2, reply would be depth 3
		return nil, apperror.Validation("maximum nesting depth reached", nil)
	}

	reply := &model.Comment{
		PostID:      parent.PostID,
		ParentID:    &id,
		UserID:      &userID,
		AuthorName:  userName,
		AuthorEmail: userEmail,
		Content:     req.Content,
		Status:      model.CommentStatusApproved,
	}

	if err := s.repo.Create(ctx, reply); err != nil {
		return nil, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "comment",
		ResourceID:   reply.ID,
	})

	// Async email notification to original comment author
	if parent.AuthorEmail != "" {
		go func() {
			msg := mail.Message{
				To:      parent.AuthorEmail,
				Subject: fmt.Sprintf("%s replied to your comment", userName),
				HTML:    fmt.Sprintf("<p>%s replied to your comment:</p><blockquote>%s</blockquote>", userName, req.Content),
			}
			if err := s.mailer.Send(context.Background(), msg); err != nil {
				slog.Error("failed to send comment reply notification", "error", err, "to", parent.AuthorEmail)
			}
		}()
	}

	resp := ToCommentDetailResp(reply)
	return &resp, nil
}

// BatchUpdateStatus bulk-updates the status of up to 100 comments.
func (s *Service) BatchUpdateStatus(ctx context.Context, req *BatchStatusReq) (int64, error) {
	status := StringToCommentStatus(req.Status)
	affected, err := s.repo.BatchUpdateStatus(ctx, req.IDs, status)
	if err != nil {
		return 0, err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "comment",
		ResourceID:   fmt.Sprintf("batch:%d", len(req.IDs)),
	})
	return affected, nil
}

// DeleteComment hard-deletes a comment (FK CASCADE removes children).
func (s *Service) DeleteComment(ctx context.Context, id string) error {
	if _, err := s.repo.GetByID(ctx, id); err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	_ = s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "comment",
		ResourceID:   id,
	})
	return nil
}
