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

	"github.com/sky-flux/cms/internal/middleware"
)

func Setup(engine *gin.Engine, db *bun.DB, rdb *redis.Client, meili meilisearch.ServiceManager, s3Client *s3.Client) {
	// Global middleware chain
	engine.Use(
		middleware.Recovery(),
		middleware.RequestID(),
		middleware.Logger(),
		middleware.CORS("http://localhost:3000"),
	)

	// Health check
	engine.GET("/health", healthHandler(db, rdb, meili, s3Client))

	// API v1 routes
	// v1 := engine.Group("/api/v1")
	// TODO: register module routes
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
