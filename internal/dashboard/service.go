package dashboard

import "context"

// Service provides dashboard statistics.
type Service struct {
	repo StatsReader
}

func NewService(repo StatsReader) *Service {
	return &Service{repo: repo}
}

func (s *Service) GetStats(ctx context.Context, siteSchema string) (*DashboardStats, error) {
	return s.repo.GetStats(ctx, siteSchema)
}
