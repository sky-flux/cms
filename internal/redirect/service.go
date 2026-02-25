package redirect

import (
	"context"
	"encoding/csv"
	"fmt"
	"io"
	"log/slog"
	"strconv"
	"strings"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
)

// Service implements redirect business logic.
type Service struct {
	repo  RedirectRepository
	audit audit.Logger
	cache *cache.Client
}

// NewService creates a new redirect service.
func NewService(repo RedirectRepository, auditLogger audit.Logger, cacheClient *cache.Client) *Service {
	return &Service{
		repo:  repo,
		audit: auditLogger,
		cache: cacheClient,
	}
}

// List returns a paginated list of redirects.
func (s *Service) List(ctx context.Context, filter ListFilter) ([]RedirectResp, int64, error) {
	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.PerPage < 1 {
		filter.PerPage = 20
	}

	redirects, total, err := s.repo.List(ctx, filter)
	if err != nil {
		return nil, 0, fmt.Errorf("list redirects: %w", err)
	}

	resp := make([]RedirectResp, len(redirects))
	for i := range redirects {
		resp[i] = ToRedirectResp(&redirects[i])
	}
	return resp, total, nil
}

// GetRedirect returns a single redirect by ID.
func (s *Service) GetRedirect(ctx context.Context, id string) (*RedirectResp, error) {
	rd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}
	resp := ToRedirectResp(rd)
	return &resp, nil
}

// Create creates a new redirect.
func (s *Service) Create(ctx context.Context, siteSlug string, req *CreateRedirectReq, userID string) (*RedirectResp, error) {
	path, err := validateSourcePath(req.SourcePath)
	if err != nil {
		return nil, err
	}

	exists, err := s.repo.SourcePathExists(ctx, path, "")
	if err != nil {
		return nil, fmt.Errorf("create redirect check path: %w", err)
	}
	if exists {
		return nil, apperror.Conflict("source path already exists", nil)
	}

	statusCode := req.StatusCode
	if statusCode == 0 {
		statusCode = 301
	}

	rd := &model.Redirect{
		SourcePath: path,
		TargetURL:  req.TargetURL,
		StatusCode: statusCode,
		Status:     model.RedirectStatusActive,
		CreatedBy:  &userID,
	}

	if err := s.repo.Create(ctx, rd); err != nil {
		return nil, fmt.Errorf("create redirect insert: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionCreate,
		ResourceType:     "redirect",
		ResourceID:       rd.ID,
		ResourceSnapshot: rd,
	}); err != nil {
		slog.Error("audit log redirect create failed", "error", err)
	}

	s.invalidateCache(ctx, siteSlug)

	resp := ToRedirectResp(rd)
	return &resp, nil
}

// Update updates an existing redirect.
func (s *Service) Update(ctx context.Context, siteSlug string, id string, req *UpdateRedirectReq) (*RedirectResp, error) {
	rd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return nil, err
	}

	if req.SourcePath != nil {
		path, err := validateSourcePath(*req.SourcePath)
		if err != nil {
			return nil, err
		}
		if path != rd.SourcePath {
			exists, err := s.repo.SourcePathExists(ctx, path, rd.ID)
			if err != nil {
				return nil, fmt.Errorf("update redirect check path: %w", err)
			}
			if exists {
				return nil, apperror.Conflict("source path already exists", nil)
			}
			rd.SourcePath = path
		}
	}

	if req.TargetURL != nil {
		rd.TargetURL = *req.TargetURL
	}
	if req.StatusCode != nil {
		rd.StatusCode = *req.StatusCode
	}
	if req.IsActive != nil {
		if *req.IsActive {
			rd.Status = model.RedirectStatusActive
		} else {
			rd.Status = model.RedirectStatusDisabled
		}
	}

	if err := s.repo.Update(ctx, rd); err != nil {
		return nil, fmt.Errorf("update redirect: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionUpdate,
		ResourceType:     "redirect",
		ResourceID:       rd.ID,
		ResourceSnapshot: rd,
	}); err != nil {
		slog.Error("audit log redirect update failed", "error", err)
	}

	s.invalidateCache(ctx, siteSlug)

	resp := ToRedirectResp(rd)
	return &resp, nil
}

// Delete deletes a redirect by ID.
func (s *Service) Delete(ctx context.Context, siteSlug string, id string) error {
	rd, err := s.repo.GetByID(ctx, id)
	if err != nil {
		return err
	}

	if err := s.repo.Delete(ctx, id); err != nil {
		return fmt.Errorf("delete redirect: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:           model.LogActionDelete,
		ResourceType:     "redirect",
		ResourceID:       rd.ID,
		ResourceSnapshot: rd,
	}); err != nil {
		slog.Error("audit log redirect delete failed", "error", err)
	}

	s.invalidateCache(ctx, siteSlug)

	return nil
}

// BatchDelete deletes multiple redirects by IDs.
func (s *Service) BatchDelete(ctx context.Context, siteSlug string, ids []string) (int64, error) {
	count, err := s.repo.BatchDelete(ctx, ids)
	if err != nil {
		return 0, fmt.Errorf("batch delete redirects: %w", err)
	}

	if err := s.audit.Log(ctx, audit.Entry{
		Action:       model.LogActionDelete,
		ResourceType: "redirect",
		ResourceID:   fmt.Sprintf("batch:%d", count),
	}); err != nil {
		slog.Error("audit log redirect batch delete failed", "error", err)
	}

	s.invalidateCache(ctx, siteSlug)

	return count, nil
}

// Import parses a CSV reader and bulk-inserts redirects. Skips duplicates.
func (s *Service) Import(ctx context.Context, siteSlug string, r io.Reader, userID string) (*ImportResult, error) {
	reader := csv.NewReader(r)

	header, err := reader.Read()
	if err != nil {
		return nil, apperror.Validation("cannot read CSV header", err)
	}
	if !validCSVHeader(header) {
		return nil, apperror.Validation("invalid CSV header: expected source_path,target_url[,status_code]", nil)
	}

	hasStatusCode := len(header) >= 3 && strings.TrimSpace(strings.ToLower(header[2])) == "status_code"

	var toInsert []*model.Redirect
	result := &ImportResult{}
	lineNum := 1

	for {
		record, err := reader.Read()
		if err == io.EOF {
			break
		}
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %v", lineNum+1, err))
			lineNum++
			continue
		}

		if lineNum >= 1000 {
			result.Errors = append(result.Errors, "maximum 1000 rows exceeded, remaining rows skipped")
			break
		}

		if len(record) < 2 {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: insufficient columns", lineNum+1))
			lineNum++
			continue
		}

		sourcePath := strings.TrimSpace(record[0])
		targetURL := strings.TrimSpace(record[1])

		path, err := validateSourcePath(sourcePath)
		if err != nil {
			result.Errors = append(result.Errors, fmt.Sprintf("line %d: %s", lineNum+1, err.Error()))
			lineNum++
			continue
		}

		statusCode := 301
		if hasStatusCode && len(record) >= 3 && strings.TrimSpace(record[2]) != "" {
			sc, err := strconv.Atoi(strings.TrimSpace(record[2]))
			if err != nil || (sc != 301 && sc != 302) {
				result.Errors = append(result.Errors, fmt.Sprintf("line %d: invalid status_code (must be 301 or 302)", lineNum+1))
				lineNum++
				continue
			}
			statusCode = sc
		}

		exists, err := s.repo.SourcePathExists(ctx, path, "")
		if err != nil {
			return nil, fmt.Errorf("import check path: %w", err)
		}
		if exists {
			result.Skipped++
			lineNum++
			continue
		}

		toInsert = append(toInsert, &model.Redirect{
			SourcePath: path,
			TargetURL:  targetURL,
			StatusCode: statusCode,
			Status:     model.RedirectStatusActive,
			CreatedBy:  &userID,
		})
		lineNum++
	}

	if len(toInsert) > 0 {
		inserted, err := s.repo.BulkInsert(ctx, toInsert)
		if err != nil {
			return nil, fmt.Errorf("import bulk insert: %w", err)
		}
		result.Imported = int(inserted)

		if err := s.audit.Log(ctx, audit.Entry{
			Action:       model.LogActionCreate,
			ResourceType: "redirect",
			ResourceID:   fmt.Sprintf("import:%d", inserted),
		}); err != nil {
			slog.Error("audit log redirect import failed", "error", err)
		}

		s.invalidateCache(ctx, siteSlug)
	}

	return result, nil
}

// Export returns all redirects for CSV export.
func (s *Service) Export(ctx context.Context) ([]model.Redirect, error) {
	return s.repo.ListAll(ctx)
}

func (s *Service) invalidateCache(ctx context.Context, siteSlug string) {
	if s.cache == nil {
		return
	}
	key := fmt.Sprintf("site:%s:redirects:map", siteSlug)
	if err := s.cache.Del(ctx, key); err != nil {
		slog.Error("cache invalidate redirects failed", "error", err, "site", siteSlug)
	}
}

func validateSourcePath(path string) (string, error) {
	if !strings.HasPrefix(path, "/") {
		return "", apperror.Validation("source_path must start with /", nil)
	}
	if strings.Contains(path, "?") {
		return "", apperror.Validation("source_path must not contain query string", nil)
	}
	// Strip trailing slash (except for root "/").
	if len(path) > 1 {
		path = strings.TrimRight(path, "/")
	}
	return path, nil
}

func validCSVHeader(header []string) bool {
	if len(header) < 2 {
		return false
	}
	h0 := strings.TrimSpace(strings.ToLower(header[0]))
	h1 := strings.TrimSpace(strings.ToLower(header[1]))
	return h0 == "source_path" && h1 == "target_url"
}
