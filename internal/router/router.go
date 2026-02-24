package router

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/apikey"
	"github.com/sky-flux/cms/internal/audit"
	"github.com/sky-flux/cms/internal/auth"
	"github.com/sky-flux/cms/internal/category"
	"github.com/sky-flux/cms/internal/config"
	"github.com/sky-flux/cms/internal/media"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/model"
	pkgaudit "github.com/sky-flux/cms/internal/pkg/audit"
	"github.com/sky-flux/cms/internal/pkg/cache"
	"github.com/sky-flux/cms/internal/pkg/imaging"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/sky-flux/cms/internal/pkg/mail"
	"github.com/sky-flux/cms/internal/pkg/search"
	"github.com/sky-flux/cms/internal/pkg/storage"
	"github.com/sky-flux/cms/internal/posttype"
	"github.com/sky-flux/cms/internal/rbac"
	"github.com/sky-flux/cms/internal/setup"
	"github.com/sky-flux/cms/internal/site"
	"github.com/sky-flux/cms/internal/system"
	"github.com/sky-flux/cms/internal/tag"
	"github.com/sky-flux/cms/internal/user"
)

// siteLookupAdapter implements middleware.SiteLookup using direct DB queries.
type siteLookupAdapter struct {
	db *bun.DB
}

func (a *siteLookupAdapter) GetIDBySlug(ctx context.Context, slug string) (string, error) {
	var id string
	err := a.db.NewSelect().
		Model((*model.Site)(nil)).
		Column("id").
		Where("slug = ?", slug).
		Where("status = ?", model.SiteStatusActive).
		Scan(ctx, &id)
	if err != nil {
		return "", err
	}
	return id, nil
}

func (a *siteLookupAdapter) GetSlugByDomain(ctx context.Context, domain string) (string, string, error) {
	var site model.Site
	err := a.db.NewSelect().
		Model(&site).
		Column("id", "slug").
		Where("domain = ?", domain).
		Where("status = ?", model.SiteStatusActive).
		Scan(ctx)
	if err != nil {
		return "", "", err
	}
	return site.Slug, site.ID, nil
}

func Setup(engine *gin.Engine, db *bun.DB, rdb *redis.Client, meili meilisearch.ServiceManager, s3Client *s3.Client, cfg *config.Config) {
	// Global middleware chain
	engine.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.CORS(cfg.Server.FrontendURL),
	)

	// Health check
	engine.GET("/health", healthHandler(db, rdb, meili, s3Client))

	// ── JWT Manager ──────────────────────────────────────────────
	jwtMgr := jwt.NewManager(cfg.JWT.Secret, cfg.JWT.AccessExpiry, 5*time.Minute, rdb)

	// ── Setup module (repos → service → handler) ─────────────────
	setupConfigRepo := setup.NewConfigRepo(db)
	setupSiteRepo := setup.NewSiteRepo(db)
	setupUserRepo := setup.NewUserRepo(db)
	setupUserRoleRepo := setup.NewUserRoleRepo(db)
	setupSvc := setup.NewService(db, rdb, jwtMgr, setupConfigRepo, setupUserRepo, setupSiteRepo, setupUserRoleRepo, cfg.JWT.AccessExpiry)
	setupHandler := setup.NewHandler(setupSvc)

	// ── Installation guard (setupSvc implements InstallChecker) ──
	engine.Use(middleware.InstallationGuard(setupSvc, "/health", "/api/v1/setup/"))

	// ── Auth module (repos → service → handler) ──────────────────
	authUserRepo := auth.NewUserRepo(db)
	authTokenRepo := auth.NewTokenRepo(db)
	authTOTPRepo := auth.NewTOTPRepo(db)
	authRoleLoader := auth.NewRoleLoader(db)
	authSiteLoader := auth.NewSiteLoader(db)
	authSvc := auth.NewService(authUserRepo, authTokenRepo, authTOTPRepo, authRoleLoader, authSiteLoader, jwtMgr, rdb, auth.ServiceConfig{
		TOTPEncryptionKey: cfg.TOTP.EncryptionKey,
		AccessExpiry:      cfg.JWT.AccessExpiry,
		RefreshExpiry:     cfg.JWT.RefreshExpiry,
	})
	authHandler := auth.NewHandler(authSvc, cfg.JWT.RefreshExpiry)

	// ── RBAC module ──────────────────────────────────────────────
	rbacUserRoleRepo := rbac.NewUserRoleRepo(db)
	rbacRoleAPIRepo := rbac.NewRoleAPIRepo(db)
	rbacMenuRepo := rbac.NewMenuRepo(db)
	rbacRoleRepo := rbac.NewRoleRepo(db)
	rbacAPIRepo := rbac.NewAPIRepo(db)
	rbacTemplateRepo := rbac.NewTemplateRepo(db)
	rbacSvc := rbac.NewService(rbacUserRoleRepo, rbacRoleAPIRepo, rbacMenuRepo, rdb)
	rbacHandler := rbac.NewHandler(rbacSvc, rbacRoleRepo, rbacAPIRepo, rbacRoleAPIRepo, rbacMenuRepo, rbacTemplateRepo, rbacUserRoleRepo)

	// ── Site module (repos → service → handler) ──────────────────
	siteRepo := site.NewSiteRepo(db)
	siteUserRoleRepo := site.NewUserRoleRepo(db)
	siteRoleResolver := site.NewRoleResolver(db)
	siteSchemaMgr := site.NewSchemaManager(db)
	siteSvc := site.NewService(siteRepo, siteUserRoleRepo, siteRoleResolver, rbacSvc, siteSchemaMgr)
	siteHandler := site.NewHandler(siteSvc)

	// ── API v1 routes ────────────────────────────────────────────
	v1 := engine.Group("/api/v1")

	// Setup routes (no auth, exempt from installation guard)
	setupGroup := v1.Group("/setup")
	setupGroup.POST("/check", setupHandler.Check)
	setupGroup.POST("/initialize", setupHandler.Initialize)

	// Auth public routes (no JWT required)
	authPublic := v1.Group("/auth")
	authPublic.POST("/login", authHandler.Login)
	authPublic.POST("/refresh", authHandler.Refresh)
	authPublic.POST("/forgot-password", authHandler.ForgotPassword)
	authPublic.POST("/reset-password", authHandler.ResetPassword)
	authPublic.POST("/2fa/validate", authHandler.Validate2FA)

	// Auth protected routes (JWT required)
	authProtected := v1.Group("/auth")
	authProtected.Use(middleware.Auth(jwtMgr))
	authProtected.POST("/logout", authHandler.Logout)
	authProtected.GET("/me", authHandler.Me)
	authProtected.PUT("/password", authHandler.ChangePassword)
	authProtected.POST("/2fa/setup", authHandler.Setup2FA)
	authProtected.POST("/2fa/verify", authHandler.Verify2FA)
	authProtected.POST("/2fa/disable", authHandler.Disable2FA)
	authProtected.POST("/2fa/backup-codes", authHandler.RegenerateBackupCodes)
	authProtected.GET("/2fa/status", authHandler.Get2FAStatus)

	// Auth admin routes (JWT + RBAC: super role required)
	authAdmin := v1.Group("/auth")
	authAdmin.Use(middleware.Auth(jwtMgr))
	authAdmin.Use(middleware.RBAC(rbacSvc))
	authAdmin.DELETE("/2fa/users/:user_id", authHandler.ForceDisable2FA)

	// Sites management (JWT + RBAC)
	sites := v1.Group("/sites")
	sites.Use(middleware.Auth(jwtMgr))
	sites.Use(middleware.RBAC(rbacSvc))
	sites.GET("", siteHandler.ListSites)
	sites.POST("", siteHandler.CreateSite)
	sites.GET("/:slug", siteHandler.GetSite)
	sites.PUT("/:slug", siteHandler.UpdateSite)
	sites.DELETE("/:slug", siteHandler.DeleteSite)
	sites.GET("/:slug/users", siteHandler.ListSiteUsers)
	sites.PUT("/:slug/users/:user_id/role", siteHandler.AssignSiteRole)
	sites.DELETE("/:slug/users/:user_id/role", siteHandler.RemoveSiteRole)

	// RBAC management (JWT + RBAC)
	rbacGroup := v1.Group("/rbac")
	rbacGroup.Use(middleware.Auth(jwtMgr))
	rbacGroup.Use(middleware.RBAC(rbacSvc))
	rbacGroup.GET("/roles", rbacHandler.ListRoles)
	rbacGroup.POST("/roles", rbacHandler.CreateRole)
	rbacGroup.GET("/roles/:id", rbacHandler.GetRole)
	rbacGroup.PUT("/roles/:id", rbacHandler.UpdateRole)
	rbacGroup.DELETE("/roles/:id", rbacHandler.DeleteRole)
	rbacGroup.GET("/roles/:id/apis", rbacHandler.GetRoleAPIs)
	rbacGroup.PUT("/roles/:id/apis", rbacHandler.SetRoleAPIs)
	rbacGroup.GET("/roles/:id/menus", rbacHandler.GetRoleMenus)
	rbacGroup.PUT("/roles/:id/menus", rbacHandler.SetRoleMenus)
	rbacGroup.POST("/roles/:id/apply-template", rbacHandler.ApplyTemplate)
	rbacGroup.GET("/users/:id/roles", rbacHandler.GetUserRoles)
	rbacGroup.POST("/users/:id/roles", rbacHandler.SetUserRoles)
	rbacGroup.GET("/menus", rbacHandler.ListMenus)
	rbacGroup.POST("/menus", rbacHandler.CreateMenu)
	rbacGroup.PUT("/menus/:id", rbacHandler.UpdateMenu)
	rbacGroup.DELETE("/menus/:id", rbacHandler.DeleteMenu)
	rbacGroup.GET("/apis", rbacHandler.ListAPIs)
	rbacGroup.GET("/templates", rbacHandler.ListTemplates)
	rbacGroup.POST("/templates", rbacHandler.CreateTemplate)
	rbacGroup.GET("/templates/:id", rbacHandler.GetTemplate)
	rbacGroup.PUT("/templates/:id", rbacHandler.UpdateTemplate)
	rbacGroup.DELETE("/templates/:id", rbacHandler.DeleteTemplate)

	// My menus (JWT only — every authenticated user can see their own menus)
	rbacMe := v1.Group("/rbac")
	rbacMe.Use(middleware.Auth(jwtMgr))
	rbacMe.GET("/me/menus", rbacHandler.GetMyMenus)

	// ── Shared services ──────────────────────────────────────────
	auditSvc := pkgaudit.NewService(db)
	var mailer mail.Sender
	if cfg.Resend.APIKey != "" {
		mailer = mail.NewResendSender(cfg.Resend.APIKey, cfg.Resend.FromName, cfg.Resend.FromEmail)
	} else {
		mailer = &mail.NoopSender{}
	}

	// ── User module (global scope — manages all CMS users) ──────
	userRepo := user.NewUserRepo(db)
	userRoleRepo := user.NewRoleRepo(db)
	userURRepo := user.NewUserRoleRepo(db)
	userTokenRevoker := user.NewTokenRevoker(db)
	userSvc := user.NewService(userRepo, userRoleRepo, userURRepo, userTokenRevoker, auditSvc, mailer, cfg.Resend.FromName)
	userHandler := user.NewHandler(userSvc)

	// Users management (JWT + AuditContext + RBAC)
	users := v1.Group("/users")
	users.Use(middleware.Auth(jwtMgr))
	users.Use(middleware.AuditContext())
	users.Use(middleware.RBAC(rbacSvc))
	users.GET("", userHandler.List)
	users.POST("", userHandler.Create)
	users.GET("/:id", userHandler.Get)
	users.PUT("/:id", userHandler.Update)
	users.DELETE("/:id", userHandler.Delete)

	// ── Site-scoped modules ──────────────────────────────────────
	siteLookup := &siteLookupAdapter{db: db}
	siteScoped := v1.Group("/site")
	siteScoped.Use(middleware.SiteResolver(siteLookup))
	siteScoped.Use(middleware.Schema(db))
	siteScoped.Use(middleware.AuditContext())
	siteScoped.Use(middleware.Auth(jwtMgr))
	siteScoped.Use(middleware.RBAC(rbacSvc))

	// Settings
	settingsRepo := system.NewConfigRepo(db)
	settingsSvc := system.NewService(settingsRepo, auditSvc)
	settingsHandler := system.NewHandler(settingsSvc)
	siteScoped.GET("/settings", settingsHandler.ListSettings)
	siteScoped.PUT("/settings", settingsHandler.UpdateSetting)

	// API Keys
	apikeyRepo := apikey.NewRepo(db)
	apikeySvc := apikey.NewService(apikeyRepo, auditSvc)
	apikeyHandler := apikey.NewHandler(apikeySvc)
	siteScoped.GET("/api-keys", apikeyHandler.ListAPIKeys)
	siteScoped.POST("/api-keys", apikeyHandler.CreateAPIKey)
	siteScoped.DELETE("/api-keys/:id", apikeyHandler.RevokeAPIKey)

	// Post Types
	posttypeRepo := posttype.NewRepo(db)
	posttypeSvc := posttype.NewService(posttypeRepo, auditSvc)
	posttypeHandler := posttype.NewHandler(posttypeSvc)
	siteScoped.GET("/post-types", posttypeHandler.List)
	siteScoped.POST("/post-types", posttypeHandler.Create)
	siteScoped.PUT("/post-types/:id", posttypeHandler.Update)
	siteScoped.DELETE("/post-types/:id", posttypeHandler.Delete)

	// Audit Logs
	auditRepo := audit.NewAuditRepo(db)
	auditHandler := audit.NewHandler(auditRepo)
	siteScoped.GET("/audit-logs", auditHandler.ListAuditLogs)

	// ── Shared infrastructure clients ───────────────────────────
	cacheClient := cache.NewClient(rdb)
	searchClient := search.NewClient(meili)
	storageClient := storage.NewClient(s3Client, cfg.RustFS.Bucket, cfg.RustFS.Endpoint+"/"+cfg.RustFS.Bucket)
	imgProcessor := imaging.NewProcessor()

	// Categories
	catRepo := category.NewRepo(db)
	catSvc := category.NewService(catRepo, cacheClient, auditSvc)
	catHandler := category.NewHandler(catSvc)
	siteScoped.GET("/categories", catHandler.List)
	siteScoped.PUT("/categories/reorder", catHandler.Reorder)
	siteScoped.GET("/categories/:id", catHandler.Get)
	siteScoped.POST("/categories", catHandler.Create)
	siteScoped.PUT("/categories/:id", catHandler.Update)
	siteScoped.DELETE("/categories/:id", catHandler.Delete)

	// Tags
	tagRepo := tag.NewRepo(db)
	tagSvc := tag.NewService(tagRepo, searchClient, cacheClient, auditSvc)
	tagHandler := tag.NewHandler(tagSvc)
	siteScoped.GET("/tags", tagHandler.List)
	siteScoped.GET("/tags/suggest", tagHandler.Suggest)
	siteScoped.GET("/tags/:id", tagHandler.Get)
	siteScoped.POST("/tags", tagHandler.Create)
	siteScoped.PUT("/tags/:id", tagHandler.Update)
	siteScoped.DELETE("/tags/:id", tagHandler.Delete)

	// Media
	mediaRepo := media.NewRepo(db)
	mediaSvc := media.NewService(mediaRepo, storageClient, imgProcessor, auditSvc)
	mediaHandler := media.NewHandler(mediaSvc)
	siteScoped.GET("/media", mediaHandler.List)
	siteScoped.DELETE("/media/batch", mediaHandler.BatchDelete)
	siteScoped.POST("/media", mediaHandler.Upload)
	siteScoped.GET("/media/:id", mediaHandler.Get)
	siteScoped.PUT("/media/:id", mediaHandler.Update)
	siteScoped.DELETE("/media/:id", mediaHandler.Delete)

	// ── API Registry — sync routes to sfc_apis ──────────────────
	registry := rbac.NewRegistry(rbacAPIRepo)
	metaMap := BuildAPIMetaMap()
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := registry.SyncRoutes(ctx, engine, metaMap); err != nil {
			slog.Error("api registry sync failed", "error", err)
		}
	}()
}

func healthHandler(db *bun.DB, rdb *redis.Client, meili meilisearch.ServiceManager, s3Client *s3.Client) gin.HandlerFunc {
	return func(c *gin.Context) {
		ctx, cancel := context.WithTimeout(c.Request.Context(), 3*time.Second)
		defer cancel()

		dbStatus := "connected"
		if err := db.PingContext(ctx); err != nil {
			dbStatus = "disconnected"
		}

		redisStatus := "connected"
		if err := rdb.Ping(ctx).Err(); err != nil {
			redisStatus = "disconnected"
		}

		meiliStatus := "connected"
		if !meili.IsHealthy() {
			meiliStatus = "disconnected"
		}

		rustfsStatus := "connected"
		if s3Client == nil {
			rustfsStatus = "disconnected"
		} else if _, err := s3Client.ListBuckets(ctx, &s3.ListBucketsInput{}); err != nil {
			rustfsStatus = "disconnected"
		}

		status := http.StatusOK
		overall := "ok"
		if dbStatus == "disconnected" || redisStatus == "disconnected" || meiliStatus == "disconnected" || rustfsStatus == "disconnected" {
			status = http.StatusServiceUnavailable
			overall = "degraded"
		}

		c.JSON(status, gin.H{
			"status":      overall,
			"db":          dbStatus,
			"redis":       redisStatus,
			"meilisearch": meiliStatus,
			"rustfs":      rustfsStatus,
		})
	}
}
