package rbac

import (
	"context"
	"fmt"
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/sky-flux/cms/internal/model"
)

// APIMeta holds metadata for route registration.
type APIMeta struct {
	Name        string
	Description string
	Group       string
}

// Registry scans Gin routes and syncs to sfc_apis.
type Registry struct {
	apiRepo APIRepository
}

// NewRegistry creates a Registry backed by the given API repository.
func NewRegistry(apiRepo APIRepository) *Registry {
	return &Registry{apiRepo: apiRepo}
}

// SyncRoutes reads all registered Gin routes and upserts to sfc_apis.
// Called once at application startup after all routes are registered.
// Routes without metadata in metaMap are skipped (public, health, etc.).
func (r *Registry) SyncRoutes(ctx context.Context, engine *gin.Engine, metaMap map[string]APIMeta) error {
	routes := engine.Routes()

	var endpoints []model.APIEndpoint
	var activeKeys []string

	for _, route := range routes {
		key := route.Method + ":" + route.Path
		meta, ok := metaMap[key]
		if !ok {
			continue
		}

		endpoints = append(endpoints, model.APIEndpoint{
			Method:      route.Method,
			Path:        route.Path,
			Name:        meta.Name,
			Description: meta.Description,
			Group:       meta.Group,
			Status:      model.APIStatusActive,
		})
		activeKeys = append(activeKeys, key)
	}

	if len(endpoints) > 0 {
		if err := r.apiRepo.UpsertBatch(ctx, endpoints); err != nil {
			return fmt.Errorf("upsert api endpoints: %w", err)
		}
	}

	if err := r.apiRepo.DisableStale(ctx, activeKeys); err != nil {
		return fmt.Errorf("disable stale endpoints: %w", err)
	}

	slog.Info("api registry synced", "total", len(endpoints))
	return nil
}
