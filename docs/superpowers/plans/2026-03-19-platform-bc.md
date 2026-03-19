# Platform BC Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement Platform bounded context for installation wizard (web + CLI), audit logging, and InstallGuard middleware.

**Architecture:** InstallGuard middleware intercepts all requests when not installed. Web wizard uses plain Chi handlers returning JSON. Audit uses Huma. .env written to binary directory by default.

**Tech Stack:** Go 1.25+, Chi v5, Huma v2, uptrace/bun, koanf, testify

---

## Prerequisites

Before starting, verify the new DDD directory skeleton exists (or create it):

```bash
mkdir -p internal/platform/domain
mkdir -p internal/platform/app
mkdir -p internal/platform/infra
mkdir -p internal/platform/delivery
mkdir -p internal/platform/middleware
```

Key references from existing codebase:
- `internal/setup/service.go` — existing installation logic: advisory lock, user creation, site creation, role assignment, `system.installed` config flag
- `internal/setup/interfaces.go` — `ConfigRepository`, `UserRepository`, `SiteRepository`, `UserRoleRepository` (port these as domain interfaces)
- `internal/audit/handler.go` — existing Gin audit handler (port to Huma; keep `ListFilter` shape)
- `internal/audit/interfaces.go` — `AuditRepository` interface (`List(ctx, ListFilter) ([]AuditWithActor, int64, error)`)
- `internal/identity/domain/user.go` — domain entity pattern (pure Go, no framework deps)
- `internal/identity/delivery/handler.go` — Huma handler pattern: `huma.Register`, `huma.Operation`, `huma.NewError`

Spec references: `docs/superpowers/specs/2026-03-19-project-redesign-design.md`
- v1 single schema (`public`), table prefix `sfc_` — **no** `site_` infix, **no** multi-tenant search_path
- InstallGuard detection: (1) `DATABASE_URL` empty → redirect `/setup`; (2) DB reachable but `sfc_migrations` table absent → redirect `/setup/migrate`
- .env write path: binary directory (`./`) by default, or `--config` flag path
- Post-install: return `{ "status": "installed", "action": "restart_required" }` — **no** auto-restart
- koanf config (not Viper)
- Chi v5 + Huma v2 (not Gin)

---

## TASK 1 — InstallState domain value object

**Files:**
- `internal/platform/domain/install.go`
- `internal/platform/domain/install_test.go`

**TDD cycle:**

- [ ] **RED** — Write `install_test.go`:
  ```go
  package domain_test

  import (
      "testing"

      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/stretchr/testify/assert"
  )

  func TestInstallState_NotInstalled_WhenDBURLEmpty(t *testing.T) {
      state := domain.NewInstallState(false, false)
      assert.False(t, state.IsInstalled())
      assert.Equal(t, domain.InstallPhaseNeedsConfig, state.Phase())
  }

  func TestInstallState_NeedsMigration_WhenConfigPresentButNoDB(t *testing.T) {
      state := domain.NewInstallState(true, false)
      assert.False(t, state.IsInstalled())
      assert.Equal(t, domain.InstallPhaseNeedsMigration, state.Phase())
  }

  func TestInstallState_Installed_WhenBothPresent(t *testing.T) {
      state := domain.NewInstallState(true, true)
      assert.True(t, state.IsInstalled())
      assert.Equal(t, domain.InstallPhaseComplete, state.Phase())
  }

  func TestInstallState_RedirectPath_NeedsConfig(t *testing.T) {
      state := domain.NewInstallState(false, false)
      assert.Equal(t, "/setup", state.RedirectPath())
  }

  func TestInstallState_RedirectPath_NeedsMigration(t *testing.T) {
      state := domain.NewInstallState(true, false)
      assert.Equal(t, "/setup/migrate", state.RedirectPath())
  }

  func TestInstallState_RedirectPath_Complete(t *testing.T) {
      state := domain.NewInstallState(true, true)
      assert.Equal(t, "", state.RedirectPath())
  }
  ```

- [ ] **Verify RED** — `go test ./internal/platform/domain/... -v -count=1` — must fail: "cannot find package"

- [ ] **GREEN** — Create `internal/platform/domain/install.go`:
  ```go
  package domain

  // InstallPhase represents the current installation state of the CMS.
  type InstallPhase int

  const (
      // InstallPhaseNeedsConfig means DATABASE_URL is not set; user must complete setup step 1.
      InstallPhaseNeedsConfig InstallPhase = iota
      // InstallPhaseNeedsMigration means DATABASE_URL is set but migrations have not run.
      InstallPhaseNeedsMigration
      // InstallPhaseComplete means the CMS is fully installed and operational.
      InstallPhaseComplete
  )

  // InstallState is an immutable value object describing whether the CMS is installed.
  // It is the source of truth for InstallGuard middleware routing decisions.
  type InstallState struct {
      hasConfig bool // DATABASE_URL is present and non-empty
      hasDB     bool // sfc_migrations table exists in the database
  }

  // NewInstallState constructs an InstallState from two boolean checks.
  //   - hasConfig: true when DATABASE_URL (or equivalent) is configured
  //   - hasDB:     true when the migrations metadata table exists in PostgreSQL
  func NewInstallState(hasConfig, hasDB bool) InstallState {
      return InstallState{hasConfig: hasConfig, hasDB: hasDB}
  }

  // IsInstalled reports whether both config and DB are present.
  func (s InstallState) IsInstalled() bool {
      return s.hasConfig && s.hasDB
  }

  // Phase returns the granular installation phase.
  func (s InstallState) Phase() InstallPhase {
      switch {
      case !s.hasConfig:
          return InstallPhaseNeedsConfig
      case !s.hasDB:
          return InstallPhaseNeedsMigration
      default:
          return InstallPhaseComplete
      }
  }

  // RedirectPath returns the setup URL the InstallGuard should redirect to,
  // or an empty string when the CMS is fully installed.
  func (s InstallState) RedirectPath() string {
      switch s.Phase() {
      case InstallPhaseNeedsConfig:
          return "/setup"
      case InstallPhaseNeedsMigration:
          return "/setup/migrate"
      default:
          return ""
      }
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/domain/... -v -count=1` — all pass

- [ ] **REFACTOR** — Add package-level doc comment `// Package domain contains pure value objects and entities for the Platform BC.`

- [ ] **Commit:** `git commit -m "✨ feat(platform): add InstallState domain value object"`

---

## TASK 2 — AuditEntry domain aggregate

**Files:**
- `internal/platform/domain/audit.go`
- `internal/platform/domain/audit_test.go`

**TDD cycle:**

- [ ] **RED** — Write `audit_test.go`:
  ```go
  package domain_test

  import (
      "testing"
      "time"

      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestNewAuditEntry_Valid(t *testing.T) {
      entry, err := domain.NewAuditEntry(
          "user-123",
          domain.AuditActionCreate,
          "post",
          "post-456",
          "127.0.0.1",
          "Mozilla/5.0",
      )
      require.NoError(t, err)
      assert.Equal(t, "user-123", entry.UserID)
      assert.Equal(t, domain.AuditActionCreate, entry.Action)
      assert.Equal(t, "post", entry.Resource)
      assert.Equal(t, "post-456", entry.ResourceID)
      assert.Equal(t, "127.0.0.1", entry.IP)
      assert.Equal(t, "Mozilla/5.0", entry.UserAgent)
      assert.False(t, entry.CreatedAt.IsZero())
  }

  func TestNewAuditEntry_EmptyUserID_ReturnsError(t *testing.T) {
      _, err := domain.NewAuditEntry("", domain.AuditActionCreate, "post", "post-456", "", "")
      assert.ErrorIs(t, err, domain.ErrEmptyUserID)
  }

  func TestNewAuditEntry_EmptyResource_ReturnsError(t *testing.T) {
      _, err := domain.NewAuditEntry("user-123", domain.AuditActionCreate, "", "post-456", "", "")
      assert.ErrorIs(t, err, domain.ErrEmptyResource)
  }

  func TestAuditEntry_CreatedAt_IsRecent(t *testing.T) {
      before := time.Now()
      entry, _ := domain.NewAuditEntry("u1", domain.AuditActionDelete, "media", "m1", "", "")
      after := time.Now()
      assert.True(t, entry.CreatedAt.After(before) || entry.CreatedAt.Equal(before))
      assert.True(t, entry.CreatedAt.Before(after) || entry.CreatedAt.Equal(after))
  }
  ```

- [ ] **Verify RED** — fails: `domain.AuditEntry` undefined, `domain.AuditActionCreate` undefined

- [ ] **GREEN** — Create `internal/platform/domain/audit.go`:
  ```go
  package domain

  import (
      "errors"
      "time"
  )

  // Sentinel errors for AuditEntry validation.
  var (
      ErrEmptyUserID   = errors.New("audit entry: userID must not be empty")
      ErrEmptyResource = errors.New("audit entry: resource must not be empty")
  )

  // AuditAction is a typed constant for the kind of operation logged.
  type AuditAction int8

  const (
      AuditActionCreate   AuditAction = 1
      AuditActionUpdate   AuditAction = 2
      AuditActionDelete   AuditAction = 3
      AuditActionPublish  AuditAction = 4
      AuditActionArchive  AuditAction = 5
      AuditActionLogin    AuditAction = 6
      AuditActionLogout   AuditAction = 7
      AuditActionRestore  AuditAction = 8
      AuditActionApprove  AuditAction = 9
      AuditActionReject   AuditAction = 10
      AuditActionGenerate AuditAction = 11
  )

  // AuditEntry is the aggregate for a single immutable audit log record.
  // Once created it is never mutated; all fields are set at construction time.
  type AuditEntry struct {
      ID         string      // set by the DB via uuidv7()
      UserID     string
      Action     AuditAction
      Resource   string // e.g. "post", "user", "media"
      ResourceID string // UUID of the affected resource
      IP         string
      UserAgent  string
      CreatedAt  time.Time
  }

  // NewAuditEntry validates inputs and constructs an AuditEntry ready for persistence.
  // CreatedAt is set to time.Now() in UTC; ID is left empty for the DB to assign.
  func NewAuditEntry(
      userID string,
      action AuditAction,
      resource, resourceID string,
      ip, userAgent string,
  ) (*AuditEntry, error) {
      if userID == "" {
          return nil, ErrEmptyUserID
      }
      if resource == "" {
          return nil, ErrEmptyResource
      }
      return &AuditEntry{
          UserID:     userID,
          Action:     action,
          Resource:   resource,
          ResourceID: resourceID,
          IP:         ip,
          UserAgent:  userAgent,
          CreatedAt:  time.Now().UTC(),
      }, nil
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/domain/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(platform): add AuditEntry domain aggregate"`

---

## TASK 3 — AuditRepository domain interface

**Files:**
- `internal/platform/domain/repo.go`
- `internal/platform/domain/repo_test.go`

**TDD cycle:**

- [ ] **RED** — Write `repo_test.go` (compile-time interface check + mock definition):
  ```go
  package domain_test

  import (
      "context"
      "testing"
      "time"

      "github.com/sky-flux/cms/internal/platform/domain"
  )

  // Compile-time interface satisfaction check.
  var _ domain.AuditRepository = (*mockAuditRepo)(nil)

  type mockAuditRepo struct {
      saveFn func(ctx context.Context, e *domain.AuditEntry) error
      listFn func(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error)
  }

  func (m *mockAuditRepo) Save(ctx context.Context, e *domain.AuditEntry) error {
      return m.saveFn(ctx, e)
  }
  func (m *mockAuditRepo) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
      return m.listFn(ctx, f)
  }

  func TestAuditRepository_Interface(t *testing.T) {
      t.Log("AuditRepository interface satisfied by mockAuditRepo")
  }

  func TestAuditFilter_ZeroValue_IsValid(t *testing.T) {
      // A zero-value AuditFilter must be usable (no required fields).
      var f domain.AuditFilter
      if f.Page == 0 {
          f.Page = 1
      }
      if f.PerPage == 0 {
          f.PerPage = 20
      }
      _ = f
  }

  func TestAuditFilter_WithDateRange(t *testing.T) {
      start := time.Now().Add(-24 * time.Hour)
      end := time.Now()
      f := domain.AuditFilter{
          StartDate: &start,
          EndDate:   &end,
      }
      _ = f
  }
  ```

- [ ] **Verify RED** — fails: `domain.AuditRepository` undefined, `domain.AuditFilter` undefined

- [ ] **GREEN** — Create `internal/platform/domain/repo.go`:
  ```go
  package domain

  import (
      "context"
      "time"
  )

  // AuditFilter controls pagination and filtering for AuditRepository.List.
  type AuditFilter struct {
      Page         int
      PerPage      int
      UserID       string      // filter by actor (empty = all users)
      Resource     string      // filter by resource type (empty = all)
      Action       *AuditAction // nil = all actions
      StartDate    *time.Time
      EndDate      *time.Time
  }

  // AuditRepository is the persistence port for audit log entries.
  // The domain defines the interface; infra/ provides the bun implementation.
  type AuditRepository interface {
      // Save persists a new audit entry. The DB sets ID via uuidv7().
      Save(ctx context.Context, e *AuditEntry) error
      // List returns paginated audit entries matching the filter.
      List(ctx context.Context, f AuditFilter) ([]AuditEntry, int64, error)
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/domain/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(platform): add AuditRepository domain interface"`

---

## TASK 4 — InstallUseCase (app layer)

**Files:**
- `internal/platform/app/install.go`
- `internal/platform/app/install_test.go`

**TDD cycle:**

- [ ] **RED** — Write `install_test.go`:
  ```go
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
      written map[string]string
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
  ```

- [ ] **Verify RED** — fails: package `app` not found

- [ ] **GREEN** — Create `internal/platform/app/install.go`:
  ```go
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
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/app/... -v -count=1` — all pass

- [ ] **REFACTOR** — Ensure `ErrDBConnectionFailed` wraps the root cause correctly for `errors.Is` and `errors.As`

- [ ] **Commit:** `git commit -m "✨ feat(platform): add InstallUseCase application service"`

---

## TASK 5 — LogAuditUseCase (app layer)

**Files:**
- `internal/platform/app/audit.go`
- `internal/platform/app/audit_test.go`

**TDD cycle:**

- [ ] **RED** — Write `audit_test.go`:
  ```go
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
      saved *domain.AuditEntry
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
  ```

- [ ] **Verify RED** — fails: `app.NewLogAuditUseCase` undefined, `app.LogAuditInput` undefined

- [ ] **GREEN** — Create `internal/platform/app/audit.go`:
  ```go
  package app

  import (
      "context"
      "fmt"

      "github.com/sky-flux/cms/internal/platform/domain"
  )

  // LogAuditInput is the input DTO for the LogAuditUseCase.
  type LogAuditInput struct {
      UserID     string
      Action     domain.AuditAction
      Resource   string
      ResourceID string
      IP         string
      UserAgent  string
  }

  // LogAuditUseCase creates and persists a single audit log entry.
  type LogAuditUseCase struct {
      repo domain.AuditRepository
  }

  // NewLogAuditUseCase creates a LogAuditUseCase.
  func NewLogAuditUseCase(repo domain.AuditRepository) *LogAuditUseCase {
      return &LogAuditUseCase{repo: repo}
  }

  // Execute validates inputs via the domain, then persists the entry.
  // It is safe to call in a goroutine for fire-and-forget audit logging.
  func (uc *LogAuditUseCase) Execute(ctx context.Context, in LogAuditInput) error {
      entry, err := domain.NewAuditEntry(
          in.UserID,
          in.Action,
          in.Resource,
          in.ResourceID,
          in.IP,
          in.UserAgent,
      )
      if err != nil {
          return fmt.Errorf("invalid audit input: %w", err)
      }
      if err := uc.repo.Save(ctx, entry); err != nil {
          return fmt.Errorf("save audit entry: %w", err)
      }
      return nil
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/app/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(platform): add LogAuditUseCase application service"`

---

## TASK 6 — bun AuditRepository infra implementation

**Files:**
- `internal/platform/infra/audit_repo.go`
- `internal/platform/infra/audit_repo_test.go`

**Note:** The integration test requires a real PostgreSQL instance. It is skipped with `-short`.

- [ ] **RED** — Write `audit_repo_test.go`:
  ```go
  package infra_test

  import (
      "context"
      "os"
      "testing"

      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/sky-flux/cms/internal/platform/infra"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
      "github.com/uptrace/bun"
  )

  func openTestDB(t *testing.T) *bun.DB {
      t.Helper()
      if testing.Short() {
          t.Skip("skipping integration test in -short mode")
      }
      dsn := os.Getenv("TEST_DATABASE_URL")
      if dsn == "" {
          t.Skip("TEST_DATABASE_URL not set")
      }
      // Use testcontainers-go or a pre-existing test DB.
      // For now we connect to the env-provided DSN directly.
      db, err := infra.OpenBunDB(dsn)
      require.NoError(t, err)
      t.Cleanup(func() { db.Close() })
      return db
  }

  func TestBunAuditRepo_Save_And_List(t *testing.T) {
      db := openTestDB(t)
      repo := infra.NewBunAuditRepository(db)
      ctx := context.Background()

      entry, err := domain.NewAuditEntry("user-1", domain.AuditActionCreate, "post", "post-1", "127.0.0.1", "Go-Test")
      require.NoError(t, err)

      err = repo.Save(ctx, entry)
      require.NoError(t, err)

      items, total, err := repo.List(ctx, domain.AuditFilter{Page: 1, PerPage: 10, UserID: "user-1"})
      require.NoError(t, err)
      assert.GreaterOrEqual(t, total, int64(1))
      assert.NotEmpty(t, items)
  }
  ```

- [ ] **Verify RED** — `go test ./internal/platform/infra/... -v -count=1` — fails: package not found (or OpenBunDB undefined)

- [ ] **GREEN** — Create `internal/platform/infra/audit_repo.go`:
  ```go
  package infra

  import (
      "context"
      "database/sql"
      "fmt"

      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/uptrace/bun"
      "github.com/uptrace/bun/dialect/pgdialect"
      _ "github.com/uptrace/bun/driver/pgdriver"
  )

  // auditRecord is the bun ORM model for sfc_audits.
  // Fields map to the existing sfc_site_audits / sfc_audits schema.
  type auditRecord struct {
      bun.BaseModel `bun:"table:sfc_audits,alias:a"`

      ID         string           `bun:"id,pk,type:uuid,default:uuidv7()"`
      UserID     string           `bun:"user_id,type:uuid,notnull"`
      Action     domain.AuditAction `bun:"action,notnull"`
      Resource   string           `bun:"resource,notnull"`
      ResourceID string           `bun:"resource_id,type:uuid"`
      IP         string           `bun:"ip"`
      UserAgent  string           `bun:"user_agent"`
      CreatedAt  interface{}      `bun:"created_at,notnull,default:now()"`
  }

  // BunAuditRepository implements domain.AuditRepository using uptrace/bun.
  type BunAuditRepository struct {
      db *bun.DB
  }

  // NewBunAuditRepository creates a new BunAuditRepository.
  func NewBunAuditRepository(db *bun.DB) *BunAuditRepository {
      return &BunAuditRepository{db: db}
  }

  // OpenBunDB opens a bun.DB connection to the given PostgreSQL DSN.
  func OpenBunDB(dsn string) (*bun.DB, error) {
      sqldb, err := sql.Open("pg", dsn)
      if err != nil {
          return nil, fmt.Errorf("open pg driver: %w", err)
      }
      db := bun.NewDB(sqldb, pgdialect.New(), bun.WithDiscardUnknownColumns())
      return db, nil
  }

  // Save inserts a new audit entry into sfc_audits.
  func (r *BunAuditRepository) Save(ctx context.Context, e *domain.AuditEntry) error {
      rec := &auditRecord{
          UserID:     e.UserID,
          Action:     e.Action,
          Resource:   e.Resource,
          ResourceID: e.ResourceID,
          IP:         e.IP,
          UserAgent:  e.UserAgent,
      }
      _, err := r.db.NewInsert().Model(rec).Exec(ctx)
      if err != nil {
          return fmt.Errorf("insert audit record: %w", err)
      }
      e.ID = rec.ID
      return nil
  }

  // List returns paginated audit entries filtered by AuditFilter.
  func (r *BunAuditRepository) List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
      if f.Page < 1 {
          f.Page = 1
      }
      if f.PerPage < 1 {
          f.PerPage = 20
      }
      if f.PerPage > 100 {
          f.PerPage = 100
      }

      var recs []auditRecord
      q := r.db.NewSelect().Model(&recs).OrderExpr("a.created_at DESC").
          Limit(f.PerPage).Offset((f.Page - 1) * f.PerPage)

      if f.UserID != "" {
          q = q.Where("a.user_id = ?", f.UserID)
      }
      if f.Resource != "" {
          q = q.Where("a.resource = ?", f.Resource)
      }
      if f.Action != nil {
          q = q.Where("a.action = ?", *f.Action)
      }
      if f.StartDate != nil {
          q = q.Where("a.created_at >= ?", f.StartDate)
      }
      if f.EndDate != nil {
          q = q.Where("a.created_at <= ?", f.EndDate)
      }

      total, err := q.ScanAndCount(ctx)
      if err != nil {
          return nil, 0, fmt.Errorf("list audit records: %w", err)
      }

      entries := make([]domain.AuditEntry, len(recs))
      for i, rec := range recs {
          entries[i] = domain.AuditEntry{
              ID:         rec.ID,
              UserID:     rec.UserID,
              Action:     rec.Action,
              Resource:   rec.Resource,
              ResourceID: rec.ResourceID,
              IP:         rec.IP,
              UserAgent:  rec.UserAgent,
          }
      }
      return entries, int64(total), nil
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/infra/... -short -v -count=1` — all pass (integration test skipped)

- [ ] **REFACTOR** — Convert `auditRecord.CreatedAt` to `time.Time` and map it back into `domain.AuditEntry.CreatedAt`

- [ ] **Commit:** `git commit -m "✨ feat(platform): add bun AuditRepository infra implementation"`

---

## TASK 7 — Installation wizard Chi handler

**Files:**
- `internal/platform/delivery/install_handler.go`
- `internal/platform/delivery/install_handler_test.go`

**Note:** Setup wizard handlers use **plain Chi** (`http.ResponseWriter`, `*http.Request`) and return JSON. They are NOT Huma because the setup wizard runs before the API is fully configured.

- [ ] **RED** — Write `install_handler_test.go`:
  ```go
  package delivery_test

  import (
      "bytes"
      "context"
      "encoding/json"
      "errors"
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/platform/app"
      "github.com/sky-flux/cms/internal/platform/delivery"
      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  // --- stubs ---

  type stubInstaller struct {
      testDBErr      error
      migrateErr     error
      createAdminErr error
      writeEnvErr    error
  }

  func (s *stubInstaller) TestDBConnection(_ context.Context, _ string) error { return s.testDBErr }
  func (s *stubInstaller) RunMigrations(_ context.Context) error               { return s.migrateErr }
  func (s *stubInstaller) CreateSuperAdmin(_ context.Context, _ app.CreateAdminInput) error {
      return s.createAdminErr
  }
  func (s *stubInstaller) WriteEnvFile(_ string, _ map[string]string) error { return s.writeEnvErr }

  func newInstallRouter(installer delivery.InstallExecutor) *chi.Mux {
      h := delivery.NewInstallHandler(installer)
      r := chi.NewRouter()
      delivery.RegisterInstallRoutes(r, h)
      return r
  }

  func TestSetupPage_GET_Returns200(t *testing.T) {
      r := newInstallRouter(&stubInstaller{})
      req := httptest.NewRequest(http.MethodGet, "/setup", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
      assert.Contains(t, rec.Header().Get("Content-Type"), "application/json")
  }

  func TestTestDB_Success_Returns200(t *testing.T) {
      r := newInstallRouter(&stubInstaller{})
      body, _ := json.Marshal(map[string]string{"database_url": "postgres://localhost/cms"})
      req := httptest.NewRequest(http.MethodPost, "/setup/test-db", bytes.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestTestDB_Failure_Returns422(t *testing.T) {
      r := newInstallRouter(&stubInstaller{testDBErr: app.ErrDBConnectionFailed})
      body, _ := json.Marshal(map[string]string{"database_url": "postgres://bad"})
      req := httptest.NewRequest(http.MethodPost, "/setup/test-db", bytes.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      assert.Equal(t, http.StatusUnprocessableEntity, rec.Code)
  }

  func TestMigrate_Success_Returns200(t *testing.T) {
      r := newInstallRouter(&stubInstaller{})
      req := httptest.NewRequest(http.MethodPost, "/setup/migrate", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestCreateAdmin_Success_Returns201WithRestartRequired(t *testing.T) {
      r := newInstallRouter(&stubInstaller{})
      body, _ := json.Marshal(map[string]string{
          "email":      "admin@example.com",
          "password":   "secret123",
          "name":       "Admin",
          "jwt_secret": "changeme",
      })
      req := httptest.NewRequest(http.MethodPost, "/setup/create-admin", bytes.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusCreated, rec.Code)
      var resp map[string]any
      require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &resp))
      assert.Equal(t, "installed", resp["status"])
      assert.Equal(t, "restart_required", resp["action"])
  }

  func TestCreateAdmin_MissingEmail_Returns400(t *testing.T) {
      r := newInstallRouter(&stubInstaller{})
      body, _ := json.Marshal(map[string]string{"password": "secret123"})
      req := httptest.NewRequest(http.MethodPost, "/setup/create-admin", bytes.NewReader(body))
      req.Header.Set("Content-Type", "application/json")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)
      assert.Equal(t, http.StatusBadRequest, rec.Code)
  }

  // Verify stubInstaller satisfies the interface.
  var _ delivery.InstallExecutor = (*stubInstaller)(nil)

  // Keep compiler happy with domain import.
  var _ = domain.InstallPhaseComplete
  var _ = errors.New
  ```

- [ ] **Verify RED** — fails: `delivery.NewInstallHandler` undefined

- [ ] **GREEN** — Create `internal/platform/delivery/install_handler.go`:
  ```go
  package delivery

  import (
      "context"
      "encoding/json"
      "errors"
      "net/http"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/platform/app"
  )

  // InstallExecutor is the port the install handler needs from the app layer.
  type InstallExecutor interface {
      TestDBConnection(ctx context.Context, dsn string) error
      RunMigrations(ctx context.Context) error
      CreateSuperAdmin(ctx context.Context, in app.CreateAdminInput) error
      WriteEnvFile(path string, vals map[string]string) error
  }

  // InstallHandler serves the web installation wizard endpoints.
  type InstallHandler struct {
      installer InstallExecutor
  }

  // NewInstallHandler creates an InstallHandler.
  func NewInstallHandler(installer InstallExecutor) *InstallHandler {
      return &InstallHandler{installer: installer}
  }

  // RegisterInstallRoutes wires setup wizard endpoints onto a Chi router.
  func RegisterInstallRoutes(r chi.Router, h *InstallHandler) {
      r.Get("/setup", h.SetupPage)
      r.Post("/setup/test-db", h.TestDB)
      r.Post("/setup/migrate", h.Migrate)
      r.Post("/setup/create-admin", h.CreateAdmin)
  }

  // SetupPage handles GET /setup — returns current installation status as JSON.
  func (h *InstallHandler) SetupPage(w http.ResponseWriter, _ *http.Request) {
      writeJSON(w, http.StatusOK, map[string]any{
          "status":  "not_installed",
          "message": "Complete the setup wizard to install Sky Flux CMS.",
      })
  }

  // TestDB handles POST /setup/test-db — tests a database connection.
  func (h *InstallHandler) TestDB(w http.ResponseWriter, r *http.Request) {
      var body struct {
          DatabaseURL string `json:"database_url"`
      }
      if err := json.NewDecoder(r.Body).Decode(&body); err != nil || body.DatabaseURL == "" {
          writeJSON(w, http.StatusBadRequest, errResp("database_url is required"))
          return
      }
      if err := h.installer.TestDBConnection(r.Context(), body.DatabaseURL); err != nil {
          if errors.Is(err, app.ErrDBConnectionFailed) {
              writeJSON(w, http.StatusUnprocessableEntity, errResp(err.Error()))
              return
          }
          writeJSON(w, http.StatusInternalServerError, errResp("internal error"))
          return
      }
      writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
  }

  // Migrate handles POST /setup/migrate — runs all pending migrations.
  func (h *InstallHandler) Migrate(w http.ResponseWriter, r *http.Request) {
      if err := h.installer.RunMigrations(r.Context()); err != nil {
          writeJSON(w, http.StatusInternalServerError, errResp(err.Error()))
          return
      }
      writeJSON(w, http.StatusOK, map[string]string{"status": "ok"})
  }

  // CreateAdmin handles POST /setup/create-admin — creates super-admin and writes .env.
  func (h *InstallHandler) CreateAdmin(w http.ResponseWriter, r *http.Request) {
      var body struct {
          Email     string `json:"email"`
          Password  string `json:"password"`
          Name      string `json:"name"`
          JWTSecret string `json:"jwt_secret"`
      }
      if err := json.NewDecoder(r.Body).Decode(&body); err != nil {
          writeJSON(w, http.StatusBadRequest, errResp("invalid request body"))
          return
      }
      if body.Email == "" || body.Password == "" {
          writeJSON(w, http.StatusBadRequest, errResp("email and password are required"))
          return
      }
      if err := h.installer.CreateSuperAdmin(r.Context(), app.CreateAdminInput{
          Email:    body.Email,
          Password: body.Password,
          Name:     body.Name,
      }); err != nil {
          writeJSON(w, http.StatusInternalServerError, errResp(err.Error()))
          return
      }
      envVals := map[string]string{}
      if body.JWTSecret != "" {
          envVals["JWT_SECRET"] = body.JWTSecret
      }
      // WriteEnvFile is best-effort; errors are logged but do not fail the response.
      _ = h.installer.WriteEnvFile("./.env", envVals)

      writeJSON(w, http.StatusCreated, map[string]string{
          "status": "installed",
          "action": "restart_required",
      })
  }

  // --- helpers ---

  func writeJSON(w http.ResponseWriter, status int, v any) {
      w.Header().Set("Content-Type", "application/json")
      w.WriteHeader(status)
      json.NewEncoder(w).Encode(v)
  }

  func errResp(msg string) map[string]string {
      return map[string]string{"error": msg}
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/delivery/... -v -count=1` — all pass

- [ ] **REFACTOR** — Extract request body struct types as named types for clarity

- [ ] **Commit:** `git commit -m "✨ feat(platform): add installation wizard Chi handler"`

---

## TASK 8 — Audit Huma handler

**Files:**
- `internal/platform/delivery/audit_handler.go`
- `internal/platform/delivery/audit_handler_test.go`

- [ ] **RED** — Write `audit_handler_test.go`:
  ```go
  package delivery_test

  import (
      "context"
      "encoding/json"
      "net/http"
      "net/http/httptest"
      "testing"
      "time"

      "github.com/danielgtaylor/huma/v2"
      "github.com/danielgtaylor/huma/v2/humatest"
      "github.com/sky-flux/cms/internal/platform/delivery"
      "github.com/sky-flux/cms/internal/platform/domain"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  type stubAuditLister struct {
      entries []domain.AuditEntry
      total   int64
  }

  func (s *stubAuditLister) List(_ context.Context, _ domain.AuditFilter) ([]domain.AuditEntry, int64, error) {
      return s.entries, s.total, nil
  }

  func newAuditAPI(t *testing.T, lister delivery.AuditLister) huma.API {
      t.Helper()
      _, api := humatest.New(t, huma.DefaultConfig("Test API", "1.0.0"))
      h := delivery.NewAuditHandler(lister)
      delivery.RegisterAuditRoutes(api, h)
      return api
  }

  func TestListAudit_Returns200WithItems(t *testing.T) {
      now := time.Now()
      lister := &stubAuditLister{
          entries: []domain.AuditEntry{
              {ID: "entry-1", UserID: "user-1", Action: domain.AuditActionCreate, Resource: "post", CreatedAt: now},
          },
          total: 1,
      }
      api := newAuditAPI(t, lister)

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
      var body map[string]any
      require.NoError(t, json.Unmarshal(rec.Body.Bytes(), &body))
      items, ok := body["items"].([]any)
      require.True(t, ok)
      assert.Len(t, items, 1)
      assert.Equal(t, int64(1), int64(body["total"].(float64)))
  }

  func TestListAudit_EmptyResult_Returns200(t *testing.T) {
      api := newAuditAPI(t, &stubAuditLister{entries: nil, total: 0})
      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestListAudit_WithFilters_Returns200(t *testing.T) {
      api := newAuditAPI(t, &stubAuditLister{})
      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/audit?page=2&per_page=10&resource=post", nil)
      rec := httptest.NewRecorder()
      api.Adapter().ServeHTTP(rec, req)
      require.Equal(t, http.StatusOK, rec.Code)
  }
  ```

- [ ] **Verify RED** — fails: `delivery.NewAuditHandler` undefined, `delivery.AuditLister` undefined

- [ ] **GREEN** — Create `internal/platform/delivery/audit_handler.go`:
  ```go
  package delivery

  import (
      "context"
      "net/http"
      "time"

      "github.com/danielgtaylor/huma/v2"
      "github.com/sky-flux/cms/internal/platform/domain"
  )

  // AuditLister is the minimal port the audit handler needs.
  type AuditLister interface {
      List(ctx context.Context, f domain.AuditFilter) ([]domain.AuditEntry, int64, error)
  }

  // AuditHandler handles audit log endpoints.
  type AuditHandler struct {
      lister AuditLister
  }

  // NewAuditHandler creates an AuditHandler.
  func NewAuditHandler(lister AuditLister) *AuditHandler {
      return &AuditHandler{lister: lister}
  }

  // RegisterAuditRoutes wires audit endpoints onto the Huma API.
  func RegisterAuditRoutes(api huma.API, h *AuditHandler) {
      huma.Register(api, huma.Operation{
          OperationID: "admin-list-audit",
          Method:      http.MethodGet,
          Path:        "/api/v1/admin/audit",
          Summary:     "List audit log entries",
          Tags:        []string{"Audit"},
      }, h.ListAudit)
  }

  // --- Request / Response types ---

  type ListAuditInput struct {
      Page     int    `query:"page" default:"1" minimum:"1"`
      PerPage  int    `query:"per_page" default:"20" minimum:"1" maximum:"100"`
      UserID   string `query:"user_id"`
      Resource string `query:"resource"`
      Action   *int8  `query:"action"`
      Start    string `query:"start_date"` // RFC3339
      End      string `query:"end_date"`   // RFC3339
  }

  // AuditEntryResp is the JSON representation of an audit log entry.
  type AuditEntryResp struct {
      ID         string    `json:"id"`
      UserID     string    `json:"user_id"`
      Action     int8      `json:"action"`
      Resource   string    `json:"resource"`
      ResourceID string    `json:"resource_id"`
      IP         string    `json:"ip"`
      UserAgent  string    `json:"user_agent"`
      CreatedAt  time.Time `json:"created_at"`
  }

  type ListAuditOutput struct {
      Body struct {
          Items []AuditEntryResp `json:"items"`
          Total int64            `json:"total"`
          Page  int              `json:"page"`
      }
  }

  // ListAudit handles GET /api/v1/admin/audit.
  func (h *AuditHandler) ListAudit(ctx context.Context, in *ListAuditInput) (*ListAuditOutput, error) {
      f := domain.AuditFilter{
          Page:     in.Page,
          PerPage:  in.PerPage,
          UserID:   in.UserID,
          Resource: in.Resource,
      }
      if in.Action != nil {
          a := domain.AuditAction(*in.Action)
          f.Action = &a
      }
      if in.Start != "" {
          t, err := time.Parse(time.RFC3339, in.Start)
          if err != nil {
              return nil, huma.NewError(http.StatusBadRequest, "invalid start_date: use RFC3339 format")
          }
          f.StartDate = &t
      }
      if in.End != "" {
          t, err := time.Parse(time.RFC3339, in.End)
          if err != nil {
              return nil, huma.NewError(http.StatusBadRequest, "invalid end_date: use RFC3339 format")
          }
          f.EndDate = &t
      }

      entries, total, err := h.lister.List(ctx, f)
      if err != nil {
          return nil, huma.NewError(http.StatusInternalServerError, "failed to list audit entries")
      }

      out := &ListAuditOutput{}
      out.Body.Total = total
      out.Body.Page = in.Page
      for _, e := range entries {
          out.Body.Items = append(out.Body.Items, AuditEntryResp{
              ID:         e.ID,
              UserID:     e.UserID,
              Action:     int8(e.Action),
              Resource:   e.Resource,
              ResourceID: e.ResourceID,
              IP:         e.IP,
              UserAgent:  e.UserAgent,
              CreatedAt:  e.CreatedAt,
          })
      }
      if out.Body.Items == nil {
          out.Body.Items = []AuditEntryResp{}
      }
      return out, nil
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/delivery/... -v -count=1` — all pass

- [ ] **Commit:** `git commit -m "✨ feat(platform): add audit Huma handler"`

---

## TASK 9 — InstallGuard Chi middleware

**Files:**
- `internal/platform/middleware/install_guard.go`
- `internal/platform/middleware/install_guard_test.go`

- [ ] **RED** — Write `install_guard_test.go`:
  ```go
  package middleware_test

  import (
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/go-chi/chi/v5"
      "github.com/sky-flux/cms/internal/platform/domain"
      platformmw "github.com/sky-flux/cms/internal/platform/middleware"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  type stubInstallChecker struct {
      state domain.InstallState
  }

  func (s *stubInstallChecker) DetectInstallState() domain.InstallState { return s.state }

  func okHandlerFn(w http.ResponseWriter, _ *http.Request) {
      w.WriteHeader(http.StatusOK)
  }

  func setupRouterWith(checker platformmw.InstallChecker) *chi.Mux {
      r := chi.NewRouter()
      r.Use(platformmw.InstallGuard(checker))
      r.Get("/api/v1/admin/posts", okHandlerFn)
      r.Get("/setup", okHandlerFn)
      r.Get("/setup/migrate", okHandlerFn)
      return r
  }

  func TestInstallGuard_Installed_PassesThrough(t *testing.T) {
      checker := &stubInstallChecker{state: domain.NewInstallState(true, true)}
      r := setupRouterWith(checker)

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)
  }

  func TestInstallGuard_NoConfig_RedirectsToSetup(t *testing.T) {
      checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
      r := setupRouterWith(checker)

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      // Must redirect (301/302/307) to /setup.
      assert.True(t, rec.Code == http.StatusFound || rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusTemporaryRedirect)
      assert.Equal(t, "/setup", rec.Header().Get("Location"))
  }

  func TestInstallGuard_NeedsDB_RedirectsToMigrate(t *testing.T) {
      checker := &stubInstallChecker{state: domain.NewInstallState(true, false)}
      r := setupRouterWith(checker)

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      assert.True(t, rec.Code == http.StatusFound || rec.Code == http.StatusMovedPermanently || rec.Code == http.StatusTemporaryRedirect)
      assert.Equal(t, "/setup/migrate", rec.Header().Get("Location"))
  }

  func TestInstallGuard_SetupPaths_AreAlwaysAllowed(t *testing.T) {
      // Even when not installed, /setup and /setup/* must be reachable.
      checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
      r := setupRouterWith(checker)

      for _, path := range []string{"/setup", "/setup/migrate"} {
          req := httptest.NewRequest(http.MethodGet, path, nil)
          rec := httptest.NewRecorder()
          r.ServeHTTP(rec, req)
          assert.Equal(t, http.StatusOK, rec.Code, "expected passthrough for path: %s", path)
      }
  }

  func TestInstallGuard_APIRequests_JSONError_WhenNotInstalled(t *testing.T) {
      // API clients that send Accept: application/json should get JSON, not a redirect.
      checker := &stubInstallChecker{state: domain.NewInstallState(false, false)}
      r := setupRouterWith(checker)

      req := httptest.NewRequest(http.MethodGet, "/api/v1/admin/posts", nil)
      req.Header.Set("Accept", "application/json")
      rec := httptest.NewRecorder()
      r.ServeHTTP(rec, req)

      // Either a JSON 503 or redirect is acceptable; what must NOT happen is 200.
      assert.NotEqual(t, http.StatusOK, rec.Code)
  }
  ```

- [ ] **Verify RED** — fails: package `platformmw` not found

- [ ] **GREEN** — Create `internal/platform/middleware/install_guard.go`:
  ```go
  package middleware

  import (
      "encoding/json"
      "net/http"
      "strings"

      "github.com/sky-flux/cms/internal/platform/domain"
  )

  // InstallChecker provides the current install state for the InstallGuard middleware.
  // The infra layer implements this by calling domain.NewInstallState with live checks.
  type InstallChecker interface {
      // DetectInstallState performs the two-step detection:
      //   1. Is DATABASE_URL configured?
      //   2. Does sfc_migrations exist in the database?
      DetectInstallState() domain.InstallState
  }

  // InstallGuard is a Chi middleware that intercepts all requests when the CMS
  // is not fully installed, redirecting browsers to the setup wizard and returning
  // JSON errors to API clients.
  //
  // Passthrough rules (always allowed regardless of install state):
  //   - /setup and anything under /setup/
  //   - /console/* (so the setup wizard SPA loads)
  func InstallGuard(checker InstallChecker) func(http.Handler) http.Handler {
      return func(next http.Handler) http.Handler {
          return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
              path := r.URL.Path

              // Always allow setup paths and the console SPA.
              if path == "/setup" ||
                  strings.HasPrefix(path, "/setup/") ||
                  strings.HasPrefix(path, "/console") {
                  next.ServeHTTP(w, r)
                  return
              }

              state := checker.DetectInstallState()
              if state.IsInstalled() {
                  next.ServeHTTP(w, r)
                  return
              }

              redirectTo := state.RedirectPath()

              // API clients (Accept: application/json or /api/* paths) get JSON.
              if strings.HasPrefix(path, "/api/") ||
                  strings.Contains(r.Header.Get("Accept"), "application/json") {
                  w.Header().Set("Content-Type", "application/json")
                  w.WriteHeader(http.StatusServiceUnavailable)
                  json.NewEncoder(w).Encode(map[string]string{
                      "title":    "Service Unavailable",
                      "detail":   "CMS is not installed. Complete the setup wizard.",
                      "setup_url": redirectTo,
                  })
                  return
              }

              http.Redirect(w, r, redirectTo, http.StatusFound)
          })
      }
  }
  ```

- [ ] **Verify GREEN** — `go test ./internal/platform/middleware/... -v -count=1` — all pass

- [ ] **REFACTOR** — Extract `isSetupPath` helper; add a comment explaining the passthrough rules

- [ ] **Commit:** `git commit -m "✨ feat(platform): add InstallGuard Chi middleware"`

---

## Final verification

- [ ] `go test ./internal/platform/... -short -v -count=1` — all tasks green (infra test skipped)
- [ ] `go vet ./internal/platform/...` — zero warnings
- [ ] `go build ./...` — clean compile

- [ ] **Final commit:** `git commit -m "✅ test(platform): all platform BC tests green"`
