package database

import (
	"fmt"

	"github.com/meilisearch/meilisearch-go"

	"github.com/sky-flux/cms/internal/config"
)

func NewMeilisearch(cfg *config.Config) (meilisearch.ServiceManager, error) {
	client := meilisearch.New(cfg.Meili.URL, meilisearch.WithAPIKey(cfg.Meili.MasterKey))

	if !client.IsHealthy() {
		return nil, fmt.Errorf("meilisearch health check failed")
	}

	return client, nil
}
