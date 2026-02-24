package router

import (
	"context"
	"net/http"
	"time"

	"github.com/aws/aws-sdk-go-v2/service/s3"
	"github.com/gin-gonic/gin"
	"github.com/meilisearch/meilisearch-go"
	"github.com/redis/go-redis/v9"
	"github.com/uptrace/bun"

	"github.com/sky-flux/cms/internal/auth"
	"github.com/sky-flux/cms/internal/config"
	"github.com/sky-flux/cms/internal/middleware"
	"github.com/sky-flux/cms/internal/pkg/jwt"
	"github.com/sky-flux/cms/internal/rbac"
	"github.com/sky-flux/cms/internal/setup"
)

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
	rbacSvc := rbac.NewService(rbacUserRoleRepo, rbacRoleAPIRepo, rbacMenuRepo, rdb)

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
