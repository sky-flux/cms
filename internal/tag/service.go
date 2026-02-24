package tag

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/pkg/search"
)

// Service implements tag business logic.
type Service struct {
	repo   TagRepository
	search *search.Client
	cache  *cache.Client
	audit  audit.Logger
}

// NewService creates a new tag service.
func NewService(repo TagRepository, searchClient *search.Client, cacheClient *cache.Client, auditLogger audit.Logger) *Service {
	return &Service{
		repo:   repo,
		search: searchClient,
		cache:  cacheClient,
		audit:  auditLogger,
	}
}

// List returns a paginated list of tags with post counts.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]TagResp, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	tags, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list tags: %w", err)
	}

	resp := ToTagRespList(tags, func(tagID string) int64 {
		count, err := s.repo.CountPosts(ctx, tagID)
		if err != nil {
			slog.Error("count posts for tag failed", "error", err, "tag_id", tagID)
			return 0
		}
		return count
	})

	return resp, total, nil
}

// GetTag returns a single tag by ID.
func (s *Service) GetTag(ctx context.Context, id string) (*TagResp, error) {
	tag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	count, err := s.repo.CountPosts(ctx, tag.ID)
	if err != nil {
		slog.Error("count posts for tag failed", "error", err, "tag_id", tag.ID)
	}

	resp := ToTagResp(tag, count)
	return &resp, nil
}

// CreateTag creates a new tag.
func (s *Service) CreateTag(ctx context.Context, siteSlug string, req *CreateTagReq) (*TagResp, error) {
	// Check name uniqueness.
	nameExists, err := s.repo.NameExists(ctx, req.Name, "")
	if err != nil {
		return nil, fmt.Errorf("create tag check name: %w", err)
	}
	if nameExists {
		return nil, apperror.Conflict("tag name already exists", nil)
	}

	// Check slug uniqueness.
	slugExists, err := s.repo.SlugExists(ctx, req.Slug, "")
	if err != nil {
		return nil, fmt.Errorf("create tag check slug: %w", err)
	}
	if slugExists {
		return nil, apperror.Conflict("tag slug already exists", nil)
	}

	tag := &model.Tag{
		Name: req.Name,
		Slug: req.Slug,
	}

	if err := s.repo.Create(ctx, tag); err != nil {
		return nil, fmt.Errorf("create tag insert: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionCreate,
		ResourceType:     "tag",
		ResourceID:       tag.ID,
		ResourceSnapshot: tag,
	}); err != nil {
		slog.Error("audit log tag create failed", "error", err)
	}

	s.syncToSearch(siteSlug, tag)

	resp := ToTagResp(tag, 0)
	return &resp, nil
}

// UpdateTag updates an existing tag.
func (s *Service) UpdateTag(ctx context.Context, siteSlug string, id string, req *UpdateTagReq) (*TagResp, error) {
	tag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.Name != nil && *req.Name != tag.Name {
		nameExists, err := s.repo.NameExists(ctx, *req.Name, tag.ID)
		if err != nil {
			return nil, fmt.Errorf("update tag check name: %w", err)
		}
		if nameExists {
			return nil, apperror.Conflict("tag name already exists", nil)
		}
		tag.Name = *req.Name
	}

	if req.Slug != nil && *req.Slug != tag.Slug {
		slugExists, err := s.repo.SlugExists(ctx, *req.Slug, tag.ID)
		if err != nil {
			return nil, fmt.Errorf("update tag check slug: %w", err)
		}
		if slugExists {
			return nil, apperror.Conflict("tag slug already exists", nil)
		}
		tag.Slug = *req.Slug
	}

	if err := s.repo.Update(ctx, tag); err != nil {
		return nil, fmt.Errorf("update tag: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionUpdate,
		ResourceType:     "tag",
		ResourceID:       tag.ID,
		ResourceSnapshot: tag,
	}); err != nil {
		slog.Error("audit log tag update failed", "error", err)
	}

	s.syncToSearch(siteSlug, tag)

	count, err := s.repo.CountPosts(ctx, tag.ID)
	if err != nil {
		slog.Error("count posts for tag failed", "error", err, "tag_id", tag.ID)
	}

	resp := ToTagResp(tag, count)
	return &resp, nil
}

// DeleteTag deletes a tag by ID.
func (s *Service) DeleteTag(ctx context.Context, siteSlug string, id string) error {
	tag, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete tag: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionDelete,
		ResourceType:     "tag",
		ResourceID:       tag.ID,
		ResourceSnapshot: tag,
	}); err != nil {
		slog.Error("audit log tag delete failed", "error", err)
	}

	s.removeFromSearch(siteSlug, tag.ID)

	return nil
}

// Suggest performs an autocomplete search for tags via Meilisearch.
func (s *Service) Suggest(ctx context.Context, siteSlug string, query string) ([]TagResp, error) {
	if s.search == nil {
		return []TagResp{}, nil
	}

	result, err := s.search.Search(ctx, "tags-"+siteSlug, query, &search.SearchOpts{Limit: 10})
	if err != nil {
		return nil, fmt.Errorf("suggest tags: %w", err)
	}

	tags := make([]TagResp, 0, len(result.Hits))
	for _, hit := range result.Hits {
		resp := TagResp{}
		if id, ok := hit["id"].(string); ok {
			resp.ID = id
		}
		if name, ok := hit["name"].(string); ok {
			resp.Name = name
		}
		if slug, ok := hit["slug"].(string); ok {
			resp.Slug = slug
		}
		tags = append(tags, resp)
	}

	return tags, nil
}

// syncToSearch asynchronously upserts a tag document to Meilisearch.
func (s *Service) syncToSearch(siteSlug string, tag *model.Tag) {
	if s.search == nil {
		return
	}
	go func() {
		doc := map[string]any{"id": tag.ID, "name": tag.Name, "slug": tag.Slug}
		if err := s.search.UpsertDocuments(context.Background(), "tags-"+siteSlug, []map[string]any{doc}); err != nil {
			slog.Error("meilisearch tag sync failed", "error", err, "tag_id", tag.ID)
		}
	}()
}

// removeFromSearch asynchronously removes a tag document from Meilisearch.
func (s *Service) removeFromSearch(siteSlug string, tagID string) {
	if s.search == nil {
		return
	}
	go func() {
		if err := s.search.DeleteDocuments(context.Background(), "tags-"+siteSlug, []string{tagID}); err != nil {
			slog.Error("meilisearch tag delete failed", "error", err, "tag_id", tagID)
		}
	}()
}
