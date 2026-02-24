package category

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
)

const (
	postCountCacheTTL = 60 * time.Second
	postCountKeyFmt   = "cat:postcount:%s"
)

// Service implements category business logic.
type Service struct {
	repo     CategoryRepository
	cache    *cache.Client
	auditLog audit.Logger
}

// NewService creates a new category service.
func NewService(repo CategoryRepository, c *cache.Client, auditLog audit.Logger) *Service {
	return &Service{
		repo:     repo,
		cache:    c,
		auditLog: auditLog,
	}
}

// ListTree returns all categories as a tree with post counts.
func (s *Service) ListTree(ctx context.Context) ([]CategoryResp, error) {
	cats, err := s.repo.List(ctx)
	if err != nil {
		return nil, err
	}

	// Collect post counts for all categories.
	counts := make(map[string]int64, len(cats))
	for i := range cats {
		counts[cats[i].ID] = s.getPostCount(ctx, cats[i].ID)
	}

	return buildTree(cats, counts), nil
}

// GetCategory returns a single category with its children and post count.
func (s *Service) GetCategory(ctx context.Context, id string) (*CategoryResp, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	postCount := s.getPostCount(ctx, cat.ID)
	resp := ToCategoryResp(cat, postCount)

	// Attach direct children.
	children, err := s.repo.GetChildren(ctx, cat.ID)
	if err != nil {
		return nil, fmt.Errorf("get category children: %w", err)
	}
	if len(children) > 0 {
		resp.Children = make([]*CategoryResp, len(children))
		for i := range children {
			childCount := s.getPostCount(ctx, children[i].ID)
			cr := ToCategoryResp(&children[i], childCount)
			resp.Children[i] = &cr
		}
	}

	return &resp, nil
}

// CreateCategory creates a new category with slug uniqueness check and path computation.
func (s *Service) CreateCategory(ctx context.Context, req *CreateCategoryReq) (*CategoryResp, error) {
	// Validate slug uniqueness under the same parent.
	exists, err := s.repo.SlugExistsUnderParent(ctx, req.Slug, req.ParentID, "")
	if err != nil {
		return nil, fmt.Errorf("check slug uniqueness: %w", err)
	}
	if exists {
		return nil, apperror.Conflict("slug already exists under this parent", nil)
	}

	// Compute path.
	path, err := s.computePath(ctx, req.Slug, req.ParentID)
	if err != nil {
		return nil, err
	}

	cat := &model.Category{
		Name:        req.Name,
		Slug:        req.Slug,
		Path:        path,
		ParentID:    req.ParentID,
		Description: req.Description,
		SortOrder:   req.SortOrder,
	}

	if err := s.repo.Create(ctx, cat); err != nil {
		return nil, err
	}

	// Audit log.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionCreate,
		ResourceType: "category",
		ResourceID:   cat.ID,
	}); err != nil {
		slog.Error("audit log category create failed", "error", err)
	}

	resp := ToCategoryResp(cat, 0)
	return &resp, nil
}

// UpdateCategory updates a category, handling slug/parent changes with path cascading.
func (s *Service) UpdateCategory(ctx context.Context, id string, req *UpdateCategoryReq) (*CategoryResp, error) {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	oldPath := cat.Path

	if req.Name != nil {
		cat.Name = *req.Name
	}
	if req.Description != nil {
		cat.Description = *req.Description
	}
	if req.SortOrder != nil {
		cat.SortOrder = *req.SortOrder
	}

	// Handle parent change — detect cycles first.
	parentChanged := false
	if req.ParentID != nil {
		// Detect cycle: new parent must not be the category itself or any descendant.
		newParentID := *req.ParentID
		if newParentID != "" {
			if newParentID == id {
				return nil, apperror.Validation("category cannot be its own parent", nil)
			}
			if err := s.detectCycle(ctx, id, newParentID); err != nil {
				return nil, err
			}
			cat.ParentID = &newParentID
		} else {
			cat.ParentID = nil
		}
		parentChanged = true
	}

	// Handle slug change.
	slugChanged := false
	if req.Slug != nil && *req.Slug != cat.Slug {
		// Validate slug uniqueness under the (possibly new) parent.
		exists, err := s.repo.SlugExistsUnderParent(ctx, *req.Slug, cat.ParentID, id)
		if err != nil {
			return nil, fmt.Errorf("check slug uniqueness: %w", err)
		}
		if exists {
			return nil, apperror.Conflict("slug already exists under this parent", nil)
		}
		cat.Slug = *req.Slug
		slugChanged = true
	}

	// Recompute path if slug or parent changed.
	if slugChanged || parentChanged {
		newPath, err := s.computePath(ctx, cat.Slug, cat.ParentID)
		if err != nil {
			return nil, err
		}
		cat.Path = newPath

		// Cascade path changes to descendants.
		if oldPath != cat.Path {
			if _, err := s.repo.UpdatePathPrefix(ctx, oldPath, cat.Path); err != nil {
				return nil, fmt.Errorf("cascade path update: %w", err)
			}
		}
	}

	if err := s.repo.Update(ctx, cat); err != nil {
		return nil, err
	}

	// Audit log.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionUpdate,
		ResourceType: "category",
		ResourceID:   cat.ID,
	}); err != nil {
		slog.Error("audit log category update failed", "error", err)
	}

	postCount := s.getPostCount(ctx, cat.ID)
	resp := ToCategoryResp(cat, postCount)
	return &resp, nil
}

// DeleteCategory deletes a leaf category. Returns conflict if it has children.
func (s *Service) DeleteCategory(ctx context.Context, id string) error {
	cat, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	// Check for children.
	children, err := s.repo.GetChildren(ctx, id)
	if err != nil {
		return fmt.Errorf("check children: %w", err)
	}
	if len(children) > 0 {
		return apperror.Conflict("cannot delete category with children", nil)
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return err
	}

	// Invalidate cache.
	_ = s.cache.Del(ctx, fmt.Sprintf(postCountKeyFmt, id))

	// Audit log.
	if err := s.auditLog.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "category",
		ResourceID:   cat.ID,
	}); err != nil {
		slog.Error("audit log category delete failed", "error", err)
	}

	return nil
}

// Reorder batch-updates sort_order for a list of categories.
func (s *Service) Reorder(ctx context.Context, orders []SortOrderItem) error {
	return s.repo.BatchUpdateSortOrder(ctx, orders)
}

// computePath builds a materialized path string for a category.
// Root: "/{slug}/", child: "{parentPath}{slug}/"
func (s *Service) computePath(ctx context.Context, slug string, parentID *string) (string, error) {
	if parentID == nil || *parentID == "" {
		return "/" + slug + "/", nil
	}

	parent, err := s.repo.GetByID(ctx, *parentID)
	if err != nil {
		return "", fmt.Errorf("get parent for path: %w", err)
	}
	return parent.Path + slug + "/", nil
}

// detectCycle walks the parent chain from newParentID and returns an error
// if the category with id is found (would create a cycle).
func (s *Service) detectCycle(ctx context.Context, id, newParentID string) error {
	visited := map[string]bool{id: true}
	current := newParentID

	for current != "" {
		if visited[current] {
			return apperror.Validation("moving category here would create a cycle", nil)
		}
		visited[current] = true

		parent, err := s.repo.GetByID(ctx, current)
		if err != nil {
			return fmt.Errorf("detect cycle: %w", err)
		}
		if parent.ParentID == nil {
			break
		}
		current = *parent.ParentID
	}
	return nil
}

// getPostCount returns the cached post count for a category.
func (s *Service) getPostCount(ctx context.Context, categoryID string) int64 {
	key := fmt.Sprintf(postCountKeyFmt, categoryID)
	var count int64
	err := s.cache.GetOrSet(ctx, key, &count, postCountCacheTTL, func() (any, error) {
		return s.repo.CountPosts(ctx, categoryID)
	})
	if err != nil {
		slog.Error("get post count failed", "error", err, "category_id", categoryID)
		return 0
	}
	return count
}

// buildTree assembles a flat list of categories into a tree of CategoryResp.
func buildTree(cats []model.Category, counts map[string]int64) []CategoryResp {
	// Build index by ID (pointers so children are linked before collecting roots).
	respMap := make(map[string]*CategoryResp, len(cats))
	for i := range cats {
		r := ToCategoryResp(&cats[i], counts[cats[i].ID])
		respMap[cats[i].ID] = &r
	}

	// Link children to their parents first.
	for i := range cats {
		if cats[i].ParentID != nil {
			parent, ok := respMap[*cats[i].ParentID]
			if ok {
				child := respMap[cats[i].ID]
				parent.Children = append(parent.Children, child)
			}
		}
	}

	// Collect root nodes (after all children have been linked).
	var roots []CategoryResp
	for i := range cats {
		if cats[i].ParentID == nil {
			roots = append(roots, *respMap[cats[i].ID])
		}
	}

	return roots
}
