package setup

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sync/atomic"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sky-flux/cms/internal/model"
	"github.com/sky-flux/cms/internal/pkg/apperror"
	"github.com/sky-flux/cms/internal/pkg/crypto"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/sky-flux/cms/internal/schema"
	"github.com/uptrace/bun"
)

var slugRegex = regexp.MustCompile(`^[a-z0-9_]{3,50}$`)

type Service struct {
	db           *bun.DB
	rdb          *redis.Client
	jwtMgr       *jwt.Manager
	configRepo   ConfigRepository
	userRepo     UserRepository
	siteRepo     SiteRepository
	userRoleRepo UserRoleRepository
	accessExpiry time.Duration
	installed    atomic.Int32
}

func NewService(
	db *bun.DB,
	rdb *redis.Client,
	jwtMgr *jwt.Manager,
	configRepo ConfigRepository,
	userRepo UserRepository,
	siteRepo SiteRepository,
	userRoleRepo UserRoleRepository,
	accessExpiry time.Duration,
) *Service {
	return &Service{
		db:           db,
		rdb:          rdb,
		jwtMgr:       jwtMgr,
		configRepo:   configRepo,
		userRepo:     userRepo,
		siteRepo:     siteRepo,
		userRoleRepo: userRoleRepo,
		accessExpiry: accessExpiry,
	}
}

func (s *Service) IsInstalled(ctx context.Context) bool {
	if s.installed.Load() == 1 {
		return true
	}
	val, err := s.rdb.Get(ctx, "system:installed").Result()
	if err == nil && val == "true" {
		s.installed.Store(1)
		return true
	}
	raw, err := s.configRepo.GetValue(ctx, "system.installed")
	if err != nil {
		return false
	}
	var installed bool
	if json.Unmarshal(raw, &installed) == nil && installed {
		s.rdb.Set(ctx, "system:installed", "true", 0)
		s.installed.Store(1)
		return true
	}
	return false
}

func (s *Service) MarkInstalled() {
	s.installed.Store(1)
}

func (s *Service) Check(ctx context.Context) bool {
	return s.IsInstalled(ctx)
}

func (s *Service) Initialize(ctx context.Context, req *InitializeReq) (*InitializeResp, error) {
	if s.IsInstalled(ctx) {
		return nil, apperror.Conflict("system already installed", nil)
	}
	if !slugRegex.MatchString(req.SiteSlug) {
		return nil, apperror.Validation("invalid site slug: must match ^[a-z0-9_]{3,50}$", nil)
	}
	passwordHash, err := crypto.HashPassword(req.AdminPassword)
	if err != nil {
		return nil, apperror.Internal("hash password failed", err)
	}
	locale := req.Locale
	if locale == "" {
		locale = "zh-CN"
	}
	var user model.User
	var site model.Site
	err = s.db.RunInTx(ctx, nil, func(ctx context.Context, tx bun.Tx) error {
		var locked bool
		if err := tx.QueryRowContext(ctx, "SELECT pg_try_advisory_xact_lock(1)").Scan(&locked); err != nil {
			return fmt.Errorf("advisory lock: %w", err)
		}
		if !locked {
			return apperror.Conflict("initialization already in progress", nil)
		}
		var cfg model.Config
		err := tx.NewSelect().Model(&cfg).Where("key = ?", "system.installed").Scan(ctx)
		if err == nil {
			var installed bool
			if json.Unmarshal(cfg.Value, &installed) == nil && installed {
				return apperror.Conflict("system already installed", nil)
			}
		}
		user = model.User{
			Email:        req.AdminEmail,
			PasswordHash: passwordHash,
			DisplayName:  req.AdminDisplayName,
			Status:       model.UserStatusActive,
		}
		if _, err := tx.NewInsert().Model(&user).Exec(ctx); err != nil {
			return fmt.Errorf("create admin user: %w", err)
		}
		site = model.Site{
			Name:          req.SiteName,
			Slug:          req.SiteSlug,
			Domain:        req.SiteURL,
			DefaultLocale: locale,
			Status:        model.SiteStatusActive,
		}
		if _, err := tx.NewInsert().Model(&site).Exec(ctx); err != nil {
			return fmt.Errorf("create site: %w", err)
		}
		var role model.Role
		if err := tx.NewSelect().Model(&role).Where("slug = ?", "super").Scan(ctx); err != nil {
			return fmt.Errorf("find super role: %w", err)
		}
		ur := &model.UserRole{UserID: user.ID, RoleID: role.ID}
		if _, err := tx.NewInsert().Model(ur).Exec(ctx); err != nil {
			return fmt.Errorf("assign super role: %w", err)
		}
		installedValue, _ := json.Marshal(true)
		_, err = tx.NewUpdate().Model(&model.Config{}).
			Set("value = ?", json.RawMessage(installedValue)).
			Where("key = ?", "system.installed").
			Exec(ctx)
		if err != nil {
			return fmt.Errorf("set installed flag: %w", err)
		}
		if err := schema.CreateSiteSchema(ctx, tx, req.SiteSlug); err != nil {
			return fmt.Errorf("create site schema: %w", err)
		}
		return nil
	})
	if err != nil {
		return nil, err
	}
	s.rdb.Set(ctx, "system:installed", "true", 0)
	s.installed.Store(1)
	token, err := s.jwtMgr.SignAccessToken(user.ID)
	if err != nil {
		return nil, apperror.Internal("sign token failed", err)
	}
	return &InitializeResp{
		User: UserResp{
			ID:          user.ID,
			Email:       user.Email,
			DisplayName: user.DisplayName,
		},
		Site: SiteResp{
			ID:   site.ID,
			Name: site.Name,
			Slug: site.Slug,
		},
		AccessToken: token,
		TokenType:   "Bearer",
		ExpiresIn:   int(s.accessExpiry.Seconds()),
	}, nil
}
