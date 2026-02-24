package posttype

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9_-]{1,100}$`)

type Service struct {
	repo  PostTypeRepository
	audit audit.Logger
}

func NewService(repo PostTypeRepository, auditLogger audit.Logger) *Service {
	return &Service{repo: repo, audit: auditLogger}
}

func (s *Service) List(ctx context.Context) ([]model.PostType, error) {
	return s.repo.List(ctx)
}

func (s *Service) Create(ctx context.Context, req *CreatePostTypeReq) (*model.PostType, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, apperror.Validation("invalid slug: must match ^[a-z0-9_-]{1,100}$", nil)
	}

	if err := validateFieldsJSON(req.Fields); err != nil {
		return nil, apperror.Validation("invalid fields: "+err.Error(), err)
	}

	existing, err := s.repo.GetBySlug(ctx, req.Slug)
	if err != nil && !isNotFound(err) {
		return nil, fmt.Errorf("create posttype check slug: %w", err)
	}
	if existing != nil {
		return nil, apperror.Conflict("post type slug already exists", nil)
	}

	fields := req.Fields
	if len(fields) == 0 {
		fields = json.RawMessage("[]")
	}

	pt := &model.PostType{
		Name:        req.Name,
		Slug:        req.Slug,
		Description: req.Description,
		Fields:      fields,
		BuiltIn:     model.ToggleNo,
	}

	if err := s.repo.Create(ctx, pt); err != nil {
		return nil, fmt.Errorf("create posttype insert: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionCreate,
		ResourceType:     "post_type",
		ResourceID:       pt.ID,
		ResourceSnapshot: pt,
	}); err != nil {
		slog.Error("audit log posttype create failed", "error", err)
	}

	return pt, nil
}

func (s *Service) Update(ctx context.Context, id string, req *UpdatePostTypeReq) (*model.PostType, error) {
	pt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	// Built-in types cannot have their slug modified.
	if pt.BuiltIn == model.ToggleYes && req.Slug != nil && *req.Slug != pt.Slug {
		return nil, apperror.Validation("cannot modify slug of a built-in post type", nil)
	}

	if req.Slug != nil && *req.Slug != pt.Slug {
		if !slugRegex.MatchString(*req.Slug) {
			return nil, apperror.Validation("invalid slug: must match ^[a-z0-9_-]{1,100}$", nil)
		}
		existing, err := s.repo.GetBySlug(ctx, *req.Slug)
		if err != nil && !isNotFound(err) {
			return nil, fmt.Errorf("update posttype check slug: %w", err)
		}
		if existing != nil {
			return nil, apperror.Conflict("post type slug already exists", nil)
		}
		pt.Slug = *req.Slug
	}

	if req.Name != nil {
		pt.Name = *req.Name
	}
	if req.Description != nil {
		pt.Description = *req.Description
	}
	if req.Fields != nil {
		if err := validateFieldsJSON(*req.Fields); err != nil {
			return nil, apperror.Validation("invalid fields: "+err.Error(), err)
		}
		pt.Fields = *req.Fields
	}

	if err := s.repo.Update(ctx, pt); err != nil {
		return nil, fmt.Errorf("update posttype: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionUpdate,
		ResourceType:     "post_type",
		ResourceID:       pt.ID,
		ResourceSnapshot: pt,
	}); err != nil {
		slog.Error("audit log posttype update failed", "error", err)
	}

	return pt, nil
}

func (s *Service) Delete(ctx context.Context, id string) error {
	pt, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if pt.BuiltIn == model.ToggleYes {
		return apperror.Validation("cannot delete a built-in post type", nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete posttype: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionDelete,
		ResourceType:     "post_type",
		ResourceID:       pt.ID,
		ResourceSnapshot: pt,
	}); err != nil {
		slog.Error("audit log posttype delete failed", "error", err)
	}

	return nil
}

// validateFieldsJSON checks that the raw JSON is a valid JSON array.
func validateFieldsJSON(raw json.RawMessage) error {
	if len(raw) == 0 {
		return nil
	}
	var arr []json.RawMessage
	return json.Unmarshal(raw, &arr)
}

// isNotFound returns true if err wraps apperror.ErrNotFound.
func isNotFound(err error) bool {
	return err != nil && fmt.Sprintf("%v", err) != "" && apperror.HTTPStatusCode(err) == 404
}
