package app_test

import (
	"context"
	"testing"

	"github.com/sky-flux/cms/internal/platform/app"
	"github.com/sky-flux/cms/internal/platform/domain"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockAuditRepo struct {
	saved   *domain.AuditEntry
	saveErr error
}

func (m *mockAuditRepo) Save(ctx context.Context, e *domain.AuditEntry) error {
	m.saved = e
	return m.saveErr
}
func (m *mockAuditRepo) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
	return nil, 0, nil
}

func TestLogAuditUseCase_Execute_SavesEntry(t *testing.T) {
	repo := &mockAuditRepo{}
	uc := app.NewLogAuditUseCase(repo)

	err := uc.Execute(context.Background(), app.LogAuditInput{
		UserID:     "user-123",
		Action:     domain.AuditActionCreate,
		Resource:   "post",
		ResourceID: "post-456",
		IP:         "127.0.0.1",
		UserAgent:  "Go-Test",
	})
	require.NoError(t, err)
	require.NotNil(t, repo.saved)
	assert.Equal(t, "user-123", repo.saved.UserID)
	assert.Equal(t, domain.AuditActionCreate, repo.saved.Action)
	assert.Equal(t, "post", repo.saved.Resource)
}

func TestLogAuditUseCase_Execute_ValidationError(t *testing.T) {
	repo := &mockAuditRepo{}
	uc := app.NewLogAuditUseCase(repo)

	// Empty UserID must be rejected by the domain.
	err := uc.Execute(context.Background(), app.LogAuditInput{
		UserID:   "",
		Action:   domain.AuditActionCreate,
		Resource: "post",
	})
	assert.Error(t, err)
}

func TestLogAuditUseCase_Execute_EmptyResource_ReturnsError(t *testing.T) {
	repo := &mockAuditRepo{}
	uc := app.NewLogAuditUseCase(repo)

	err := uc.Execute(context.Background(), app.LogAuditInput{
		UserID:   "user-123",
		Action:   domain.AuditActionDelete,
		Resource: "", // invalid
	})
	assert.Error(t, err)
}
