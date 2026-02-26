package dashboard

import "context"

// StatsReader retrieves dashboard statistics for a site schema.
type StatsReader interface {
	GetStats(ctx context.Context, siteSchema string) (*DashboardStats, error)
}
