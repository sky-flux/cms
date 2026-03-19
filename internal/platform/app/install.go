// Package app contains platform use cases: RunInstallWizard, RecordAudit, GetConfig.
package app

import (
	"context"
	"errors"
	"fmt"
)

// Sentinel errors exposed to callers (delivery layer maps these to HTTP status codes).
var (
	ErrDBConnectionFailed = errors.New("database connection failed")
	ErrMigrationFailed    = errors.New("database migration failed")
	ErrAdminCreateFailed  = errors.New("super-admin creation failed")
)

// DBProbe checks whether a PostgreSQL DSN is reachable.
type DBProbe interface {
	Ping(ctx context.Context, dsn string) error
}

// Migrator runs bun migrations against the configured database.
type Migrator interface {
	RunMigrations(ctx context.Context) error
}

// CreateAdminInput is the input DTO for super-admin creation.
type CreateAdminInput struct {
	Email    string
	Password string
	Name     string
}

// UserCreator creates the initial super-admin user during installation.
type UserCreator interface {
	CreateSuperAdmin(ctx context.Context, in CreateAdminInput) error
}

// EnvWriter persists key-value pairs to a .env file.
type EnvWriter interface {
	WriteEnvFile(path string, vals map[string]string) error
}

// InstallUseCase orchestrates the four-step web installation wizard.
type InstallUseCase struct {
	probe   DBProbe
	mig     Migrator
	creator UserCreator
	writer  EnvWriter
}

// NewInstallUseCase creates a new InstallUseCase with its dependencies.
func NewInstallUseCase(probe DBProbe, mig Migrator, creator UserCreator, writer EnvWriter) *InstallUseCase {
	return &InstallUseCase{probe: probe, mig: mig, creator: creator, writer: writer}
}

// TestDBConnection checks whether the given DSN is reachable.
// Returns ErrDBConnectionFailed (wrapped) on failure so callers can use errors.Is.
func (uc *InstallUseCase) TestDBConnection(ctx context.Context, dsn string) error {
	if err := uc.probe.Ping(ctx, dsn); err != nil {
		return fmt.Errorf("%w: %v", ErrDBConnectionFailed, err)
	}
	return nil
}

// RunMigrations executes all pending bun migrations.
func (uc *InstallUseCase) RunMigrations(ctx context.Context) error {
	if err := uc.mig.RunMigrations(ctx); err != nil {
		return fmt.Errorf("%w: %v", ErrMigrationFailed, err)
	}
	return nil
}

// CreateSuperAdmin creates the initial administrator account.
func (uc *InstallUseCase) CreateSuperAdmin(ctx context.Context, in CreateAdminInput) error {
	if err := uc.creator.CreateSuperAdmin(ctx, in); err != nil {
		return fmt.Errorf("%w: %v", ErrAdminCreateFailed, err)
	}
	return nil
}

// WriteEnvFile persists the provided environment variables to a .env file.
// path is usually the binary directory; vals must include at minimum DATABASE_URL.
func (uc *InstallUseCase) WriteEnvFile(path string, vals map[string]string) error {
	return uc.writer.WriteEnvFile(path, vals)
}
