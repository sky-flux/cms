package site

import (
	"context"
	"fmt"
	"log/slog"
	"regexp"

	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

type Service struct {
	siteRepo     SiteRepository
	userRoleRepo UserRoleRepository
	roleResolver RoleResolver
	rbac         RBACInvalidator
	schemaMgr    SchemaManager
}

func NewService(
	siteRepo SiteRepository,
	userRoleRepo UserRoleRepository,
	roleResolver RoleResolver,
	rbac RBACInvalidator,
	schemaMgr SchemaManager,
) *Service {
	return &Service{
		siteRepo:     siteRepo,
		userRoleRepo: userRoleRepo,
		roleResolver: roleResolver,
		rbac:         rbac,
		schemaMgr:    schemaMgr,
	}
}

func (s *Service) ListSites(ctx context.Context, f ListFilter) ([]model.Site, int64, error) {
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	return s.siteRepo.List(ctx, f)
}

func (s *Service) GetSite(ctx context.Context, slug string) (*model.Site, error) {
	return s.siteRepo.GetBySlug(ctx, slug)
}

func (s *Service) CreateSite(ctx context.Context, req *CreateSiteReq) (*model.Site, error) {
	if !slugRegex.MatchString(req.Slug) {
		return nil, apperror.Validation("invalid slug: must match ^[a-z0-9_]{3,50}$", nil)
	}

	exists, err := s.siteRepo.SlugExists(ctx, req.Slug)
	if err != nil {
		return nil, fmt.Errorf("create site check slug: %w", err)
	}
	if exists {
		return nil, apperror.Conflict("site slug already exists", nil)
	}

	if req.Domain != "" {
		domainExists, err := s.siteRepo.DomainExists(ctx, req.Domain, "")
		if err != nil {
			return nil, fmt.Errorf("create site check domain: %w", err)
		}
		if domainExists {
			return nil, apperror.Conflict("domain already exists", nil)
		}
	}

	locale := req.DefaultLocale
	if locale == "" {
		locale = "zh-CN"
	}
	tz := req.Timezone
	if tz == "" {
		tz = "Asia/Shanghai"
	}

	site := &model.Site{
		Name:          req.Name,
		Slug:          req.Slug,
		Domain:        req.Domain,
		Description:   req.Description,
		DefaultLocale: locale,
		Timezone:      tz,
		Status:        model.SiteStatusActive,
	}

	if err := s.siteRepo.Create(ctx, site); err != nil {
		return nil, fmt.Errorf("create site insert: %w", err)
	}

	if err := s.schemaMgr.Create(ctx, req.Slug); err != nil {
		slog.Error("create site schema failed", "error", err, "slug", req.Slug)
		return nil, apperror.Internal("site schema creation failed", err)
	}

	return site, nil
}

func (s *Service) UpdateSite(ctx context.Context, slug string, req *UpdateSiteReq) (*model.Site, error) {
	site, err := s.siteRepo.GetBySlug(ctx, slug)
	if err != nil {
		return nil, err
	}

	if req.Domain != nil && *req.Domain != site.Domain {
		if *req.Domain != "" {
			domainExists, err := s.siteRepo.DomainExists(ctx, *req.Domain, site.ID)
			if err != nil {
				return nil, fmt.Errorf("update site check domain: %w", err)
			}
			if domainExists {
				return nil, apperror.Conflict("domain already exists", nil)
			}
		}
		site.Domain = *req.Domain
	}

	if req.Name != nil {
		site.Name = *req.Name
	}
	if req.Description != nil {
		site.Description = *req.Description
	}
	if req.LogoURL != nil {
		site.LogoURL = *req.LogoURL
	}
	if req.DefaultLocale != nil {
		site.DefaultLocale = *req.DefaultLocale
	}
	if req.Timezone != nil {
		site.Timezone = *req.Timezone
	}
	if req.Status != nil {
		site.Status = *req.Status
	}
	if req.Settings != nil {
		site.Settings = *req.Settings
	}

	if err := s.siteRepo.Update(ctx, site); err != nil {
		return nil, fmt.Errorf("update site: %w", err)
	}
	return site, nil
}

func (s *Service) DeleteSite(ctx context.Context, slug string, confirmSlug string) error {
	if slug != confirmSlug {
		return apperror.Validation("confirm_slug does not match", nil)
	}

	site, err := s.siteRepo.GetBySlug(ctx, slug)
	if err != nil {
		return err
	}

	count, err := s.siteRepo.CountActive(ctx)
	if err != nil {
		return fmt.Errorf("delete site count: %w", err)
	}
	if count <= 1 {
		return apperror.Validation("cannot delete the last site", nil)
	}

	if err := s.siteRepo.Delete(ctx, site.ID); err != nil {
		return fmt.Errorf("delete site: %w", err)
	}

	if err := s.schemaMgr.Drop(ctx, slug); err != nil {
		slog.Error("drop site schema failed", "error", err, "slug", slug)
	}

	return nil
}

func (s *Service) ListSiteUsers(ctx context.Context, slug string, f UserFilter) ([]UserWithRole, int64, error) {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return nil, 0, err
	}
	if f.Page < 1 {
		f.Page = 1
	}
	if f.PerPage < 1 || f.PerPage > 100 {
		f.PerPage = 20
	}
	return s.userRoleRepo.ListUsersWithRoles(ctx, f)
}

func (s *Service) AssignSiteRole(ctx context.Context, slug, userID string, roleSlug string) error {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return err
	}

	exists, err := s.userRoleRepo.UserExists(ctx, userID)
	if err != nil {
		return fmt.Errorf("assign role check user: %w", err)
	}
	if !exists {
		return apperror.NotFound("user not found", nil)
	}

	role, err := s.roleResolver.GetBySlug(ctx, roleSlug)
	if err != nil {
		return err
	}

	if err := s.userRoleRepo.AssignRole(ctx, userID, role.ID); err != nil {
		return fmt.Errorf("assign role: %w", err)
	}

	if err := s.rbac.InvalidateUserCache(ctx, userID); err != nil {
		slog.Error("invalidate user cache after role assign", "error", err, "user_id", userID)
	}

	return nil
}

func (s *Service) RemoveSiteRole(ctx context.Context, slug, userID string) error {
	if _, err := s.siteRepo.GetBySlug(ctx, slug); err != nil {
		return err
	}

	if err := s.userRoleRepo.RemoveRole(ctx, userID); err != nil {
		return fmt.Errorf("remove role: %w", err)
	}

	if err := s.rbac.InvalidateUserCache(ctx, userID); err != nil {
		slog.Error("invalidate user cache after role remove", "error", err, "user_id", userID)
	}

	return nil
}
