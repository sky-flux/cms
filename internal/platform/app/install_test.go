package app_test

import (
	"context"
	"errors"
	"testing"

	"github.com/sky-flux/cms/internal/platform/app"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- hand-written mocks ---

type mockDBProbe struct{ alive bool }

func (m *mockDBProbe) Ping(ctx context.Context, dsn string) error {
	if !m.alive {
		return errors.New("connection refused")
	}
	return nil
}

type mockMigrator struct{ runErr error }

func (m *mockMigrator) RunMigrations(ctx context.Context) error { return m.runErr }

type mockUserCreator struct {
	createdEmail string
	returnErr    error
}

func (m *mockUserCreator) CreateSuperAdmin(ctx context.Context, in app.CreateAdminInput) error {
	m.createdEmail = in.Email
	return m.returnErr
}

type mockEnvWriter struct {
	written  map[string]string
	writeErr error
}

func (m *mockEnvWriter) WriteEnvFile(path string, vals map[string]string) error {
	m.written = vals
	return m.writeErr
}

// --- tests ---

func TestInstallUseCase_TestDBConnection_Success(t *testing.T) {
	uc := app.NewInstallUseCase(
		&mockDBProbe{alive: true},
		&mockMigrator{},
		&mockUserCreator{},
		&mockEnvWriter{},
	)
	err := uc.TestDBConnection(context.Background(), "postgres://localhost/cms")
	require.NoError(t, err)
}

func TestInstallUseCase_TestDBConnection_Failure(t *testing.T) {
	uc := app.NewInstallUseCase(
		&mockDBProbe{alive: false},
		&mockMigrator{},
		&mockUserCreator{},
		&mockEnvWriter{},
	)
	err := uc.TestDBConnection(context.Background(), "postgres://bad")
	assert.Error(t, err)
	assert.ErrorIs(t, err, app.ErrDBConnectionFailed)
}

func TestInstallUseCase_RunMigrations_Success(t *testing.T) {
	uc := app.NewInstallUseCase(
		&mockDBProbe{},
		&mockMigrator{runErr: nil},
		&mockUserCreator{},
		&mockEnvWriter{},
	)
	err := uc.RunMigrations(context.Background())
	require.NoError(t, err)
}

func TestInstallUseCase_RunMigrations_Failure(t *testing.T) {
	uc := app.NewInstallUseCase(
		&mockDBProbe{},
		&mockMigrator{runErr: errors.New("migration failed")},
		&mockUserCreator{},
		&mockEnvWriter{},
	)
	err := uc.RunMigrations(context.Background())
	assert.Error(t, err)
}

func TestInstallUseCase_CreateSuperAdmin_Success(t *testing.T) {
	creator := &mockUserCreator{}
	uc := app.NewInstallUseCase(&mockDBProbe{}, &mockMigrator{}, creator, &mockEnvWriter{})
	err := uc.CreateSuperAdmin(context.Background(), app.CreateAdminInput{
		Email:    "admin@example.com",
		Password: "secret123",
		Name:     "Admin",
	})
	require.NoError(t, err)
	assert.Equal(t, "admin@example.com", creator.createdEmail)
}

func TestInstallUseCase_WriteEnvFile_IncludesDBURL(t *testing.T) {
	writer := &mockEnvWriter{}
	uc := app.NewInstallUseCase(&mockDBProbe{}, &mockMigrator{}, &mockUserCreator{}, writer)
	err := uc.WriteEnvFile("./.env", map[string]string{
		"DATABASE_URL": "postgres://localhost/cms",
		"JWT_SECRET":   "changeme",
	})
	require.NoError(t, err)
	assert.Equal(t, "postgres://localhost/cms", writer.written["DATABASE_URL"])
	assert.Equal(t, "changeme", writer.written["JWT_SECRET"])
}
