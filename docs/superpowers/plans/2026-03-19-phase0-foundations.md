# Phase 0: Foundations Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the project's DDD directory structure, replace Gin/Viper with Chi+Huma+koanf, and clean up migrations for single-site v1.

**Architecture:** Single root go.mod, root-level cmd/internal structure, go:embed for console/dist and web/static.

**Tech Stack:** Go 1.25+, Chi v5, Huma v2, koanf v2, uptrace/bun, testify

---

## Overview

Phase 0 lays the non-negotiable groundwork that every subsequent phase depends on. No domain logic is implemented here — only infrastructure. Each task is self-contained and ends with a passing test suite and a commit.

**Execution order is strict**: Tasks 1 → 2 → 3 → 4 → 5 → 6. Each task's tests must be green before starting the next.

**Time estimate:** ~30–45 minutes total (5–8 min per task).

---

## Task 1: go.mod cleanup

**Remove:** `github.com/gin-gonic/gin`, `github.com/spf13/viper`
**Add:** `github.com/go-chi/chi/v5`, `github.com/danielgtaylor/huma/v2`, `github.com/knadh/koanf/v2`, `github.com/knadh/koanf/providers/env`, `github.com/knadh/koanf/providers/file`, `github.com/knadh/koanf/parsers/dotenv`

### Steps

- [ ] **RED** — Write a compile-check test confirming chi and huma packages are importable:

  Create `/Users/martinadamsdev/workspace/sky-flux-cms/internal/config/deps_test.go`:
  ```go
  package config_test

  import (
      "testing"
      _ "github.com/go-chi/chi/v5"
      _ "github.com/danielgtaylor/huma/v2"
      _ "github.com/knadh/koanf/v2"
  )

  func TestDepsImportable(t *testing.T) {
      // This test passes if the file compiles.
      // If go.mod is missing any dependency, `go test` will fail with a
      // "cannot find module" error — which is our expected RED state.
      t.Log("chi, huma, koanf are importable")
  }
  ```

- [ ] **Verify RED** — Run the test (it will fail because the packages are not yet in go.mod):
  ```bash
  go test ./internal/config/... -v -count=1
  # Expected: build failed: no required module provides github.com/go-chi/chi/v5
  ```

- [ ] **GREEN** — Add new dependencies and remove old ones:
  ```bash
  # Add new packages
  go get github.com/go-chi/chi/v5@latest
  go get github.com/danielgtaylor/huma/v2@latest
  go get github.com/knadh/koanf/v2@latest
  go get github.com/knadh/koanf/providers/env@latest
  go get github.com/knadh/koanf/providers/file@latest
  go get github.com/knadh/koanf/parsers/dotenv@latest

  # Remove old packages (keep cobra — still used for CLI)
  go get github.com/gin-gonic/gin@none
  go get github.com/spf13/viper@none

  # Tidy
  go mod tidy
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./internal/config/... -v -count=1
  # Expected: PASS — TestDepsImportable and all existing config tests pass
  ```

  > **Note:** The existing `config_test.go` imports `github.com/spf13/viper` directly. Those tests will fail after viper removal in Task 3. That is acceptable — Task 3 will rewrite both `config.go` and `config_test.go`. For now, check that `TestDepsImportable` passes and note which other tests fail (they will be fixed in Task 3).

- [ ] **Verify go.mod** — Confirm removals and additions:
  ```bash
  grep -E "gin-gonic|spf13/viper|go-chi|huma|koanf" go.mod
  # Expected:
  #   github.com/danielgtaylor/huma/v2 vX.Y.Z
  #   github.com/go-chi/chi/v5 vX.Y.Z
  #   github.com/knadh/koanf/v2 vX.Y.Z
  # NOT expected (must be gone):
  #   github.com/gin-gonic/gin
  #   github.com/spf13/viper
  ```

- [ ] **Commit**:
  ```bash
  git add go.mod go.sum internal/config/deps_test.go
  git commit -m "📦 replace Gin/Viper with Chi v5, Huma v2, koanf v2"
  ```

---

## Task 2: DDD directory scaffold

Create the `internal/` DDD skeleton. Each bounded context gets a `domain/`, `app/`, `infra/`, `delivery/` sub-directory with a placeholder `.go` file so Go can compile the package.

### Directory structure to create

```
internal/
├── identity/
│   ├── domain/doc.go
│   ├── app/doc.go
│   ├── infra/doc.go
│   └── delivery/doc.go
├── content/
│   ├── domain/doc.go
│   ├── app/doc.go
│   ├── infra/doc.go
│   └── delivery/doc.go
├── media/
│   ├── domain/doc.go
│   ├── app/doc.go
│   ├── infra/doc.go
│   └── delivery/doc.go
├── site/
│   ├── domain/doc.go
│   ├── app/doc.go
│   ├── infra/doc.go
│   └── delivery/doc.go
├── delivery/
│   └── doc.go          # (top-level delivery: public API, RSS, Sitemap)
├── platform/
│   ├── domain/doc.go
│   ├── app/doc.go
│   ├── infra/doc.go
│   └── delivery/doc.go
└── shared/
    └── doc.go          # apperror, middleware, event bus stubs
```

### Steps

- [ ] **RED** — Write a test that imports one of the new packages (which don't exist yet):

  Create `/Users/martinadamsdev/workspace/sky-flux-cms/internal/shared/scaffold_test.go`:
  ```go
  package shared_test

  import (
      "testing"
      _ "github.com/sky-flux/cms/internal/identity/domain"
      _ "github.com/sky-flux/cms/internal/content/domain"
      _ "github.com/sky-flux/cms/internal/media/domain"
      _ "github.com/sky-flux/cms/internal/site/domain"
      _ "github.com/sky-flux/cms/internal/platform/domain"
      _ "github.com/sky-flux/cms/internal/shared"
  )

  func TestDDDScaffoldExists(t *testing.T) {
      t.Log("all DDD bounded context packages are importable")
  }
  ```

- [ ] **Verify RED**:
  ```bash
  go test ./internal/shared/... -v -count=1
  # Expected: build failed: no required module provides .../internal/identity/domain
  ```

- [ ] **GREEN** — Create all `doc.go` placeholder files. Each file follows this pattern (adjust `package` name):

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/identity/domain/doc.go`:
  ```go
  // Package domain contains the identity bounded context's domain model:
  // entities, value objects, repository interfaces, and domain events.
  // No framework dependencies are allowed in this package.
  package domain
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/identity/app/doc.go`:
  ```go
  // Package app contains identity use cases (application services).
  // Depends only on identity/domain interfaces — no infrastructure imports.
  package app
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/identity/infra/doc.go`:
  ```go
  // Package infra contains infrastructure implementations for the identity
  // bounded context: bun repositories, Redis adapters.
  package infra
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/identity/delivery/doc.go`:
  ```go
  // Package delivery contains Huma HTTP handlers and DTOs for the identity
  // bounded context.
  package delivery
  ```

  Repeat the same four files for `content/`, `media/`, `site/`, `platform/` — changing the package comment to reflect the BC name.

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/content/domain/doc.go`:
  ```go
  // Package domain contains the content bounded context's domain model:
  // Post, Category, Tag, Comment aggregates, repository interfaces, and events.
  package domain
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/content/app/doc.go`:
  ```go
  // Package app contains content use cases: CreatePost, PublishPost, ListPosts, etc.
  package app
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/content/infra/doc.go`:
  ```go
  // Package infra contains bun/Meilisearch implementations for content repositories.
  package infra
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/content/delivery/doc.go`:
  ```go
  // Package delivery contains Huma handlers and DTOs for content management endpoints.
  package delivery
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/media/domain/doc.go`:
  ```go
  // Package domain contains the media bounded context's domain model:
  // MediaFile aggregate, repository interface, storage port.
  package domain
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/media/app/doc.go`:
  ```go
  // Package app contains media use cases: UploadFile, DeleteFile, ListFiles.
  package app
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/media/infra/doc.go`:
  ```go
  // Package infra contains bun/RustFS (S3) implementations for media repositories.
  package infra
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/media/delivery/doc.go`:
  ```go
  // Package delivery contains Huma handlers and DTOs for media upload/management.
  package delivery
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/site/domain/doc.go`:
  ```go
  // Package domain contains the site bounded context's domain model:
  // Site, Menu, Redirect aggregates and repository interfaces.
  package domain
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/site/app/doc.go`:
  ```go
  // Package app contains site use cases: ConfigureSite, ManageMenus, ManageRedirects.
  package app
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/site/infra/doc.go`:
  ```go
  // Package infra contains bun implementations for site repositories.
  package infra
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/site/delivery/doc.go`:
  ```go
  // Package delivery contains Huma handlers for site configuration endpoints.
  package delivery
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/platform/domain/doc.go`:
  ```go
  // Package domain contains the platform bounded context's domain model:
  // AuditLog, SystemConfig value objects and repository interfaces.
  package domain
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/platform/app/doc.go`:
  ```go
  // Package app contains platform use cases: RunInstallWizard, RecordAudit, GetConfig.
  package app
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/platform/infra/doc.go`:
  ```go
  // Package infra contains bun implementations for platform repositories.
  package infra
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/platform/delivery/doc.go`:
  ```go
  // Package delivery contains Huma handlers for install wizard and system config.
  package delivery
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/delivery/doc.go`:
  ```go
  // Package delivery contains the top-level delivery layer: public REST API handlers,
  // RSS/Atom feed generation, Sitemap generation, and web (Templ+HTMX) handlers.
  package delivery
  ```

  `/Users/martinadamsdev/workspace/sky-flux-cms/internal/shared/doc.go`:
  ```go
  // Package shared contains cross-cutting concerns shared across all bounded contexts:
  // apperror sentinel errors, HTTP middleware, and the in-memory event bus.
  package shared
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./internal/shared/... -v -count=1
  # Expected: PASS — TestDDDScaffoldExists
  go build ./internal/...
  # Expected: no errors
  ```

- [ ] **Commit**:
  ```bash
  git add internal/identity/ internal/content/ internal/media/ \
          internal/site/ internal/platform/ internal/delivery/ \
          internal/shared/
  git commit -m "✨ add DDD bounded context directory scaffold"
  ```

---

## Task 3: koanf config

Rewrite `internal/config/config.go` to use koanf instead of viper. The public `Config` struct and all field names remain identical — only the loader internals change.

### Steps

- [ ] **RED** — The existing tests in `config_test.go` currently import `github.com/spf13/viper` for `resetViper()`. After viper removal they already fail. Confirm this:
  ```bash
  go test ./internal/config/... -v -count=1
  # Expected: build failed — "no required module provides github.com/spf13/viper"
  # (viper was removed in Task 1)
  ```

  This failing build IS our red state. Proceed to GREEN.

- [ ] **GREEN (step A)** — Rewrite `/Users/martinadamsdev/workspace/sky-flux-cms/internal/config/config.go`:

  ```go
  package config

  import (
      "fmt"
      "time"

      "github.com/knadh/koanf/parsers/dotenv"
      "github.com/knadh/koanf/providers/env"
      "github.com/knadh/koanf/providers/file"
      "github.com/knadh/koanf/v2"
  )

  // Config holds all application configuration. Field names and types are
  // identical to the previous Viper-based version to preserve compatibility.
  type Config struct {
      Server ServerConfig
      DB     DBConfig
      Redis  RedisConfig
      Meili  MeiliConfig
      JWT    JWTConfig
      TOTP   TOTPConfig
      RustFS RustFSConfig
      Resend ResendConfig
      Log    LogConfig
  }

  type ServerConfig struct {
      Port        string
      Mode        string
      FrontendURL string
  }

  type DBConfig struct {
      Host            string
      Port            string
      Name            string
      User            string
      Password        string
      SSLMode         string
      MaxOpenConns    int
      MaxIdleConns    int
      ConnMaxLifetime time.Duration
      ConnMaxIdleTime time.Duration
  }

  func (c *DBConfig) DSN() string {
      return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=%s",
          c.User, c.Password, c.Host, c.Port, c.Name, c.SSLMode)
  }

  type RedisConfig struct {
      Host     string
      Port     string
      Password string
      DB       int
  }

  func (c *RedisConfig) Addr() string {
      return c.Host + ":" + c.Port
  }

  type MeiliConfig struct {
      URL       string
      MasterKey string
  }

  type JWTConfig struct {
      Secret        string
      AccessExpiry  time.Duration
      RefreshExpiry time.Duration
  }

  type TOTPConfig struct {
      EncryptionKey string
  }

  type RustFSConfig struct {
      Endpoint  string
      AccessKey string
      SecretKey string
      Bucket    string
      Region    string
  }

  type ResendConfig struct {
      APIKey    string
      FromName  string
      FromEmail string
  }

  type LogConfig struct {
      Level string
  }

  // defaults maps flat ENV key names to their default values.
  // koanf uses these as the base layer before file and env providers.
  var defaults = map[string]any{
      "SERVER_PORT":          "8080",
      "SERVER_MODE":          "debug",
      "FRONTEND_URL":         "http://localhost:3000",
      "DB_HOST":              "localhost",
      "DB_PORT":              "5432",
      "DB_SSLMODE":           "disable",
      "DB_MAX_OPEN_CONNS":    25,
      "DB_MAX_IDLE_CONNS":    5,
      "DB_CONN_MAX_LIFETIME": "1h",
      "DB_CONN_MAX_IDLE_TIME": "30m",
      "REDIS_HOST":           "localhost",
      "REDIS_PORT":           "6379",
      "REDIS_DB":             0,
      "MEILI_URL":            "http://localhost:7700",
      "JWT_ACCESS_EXPIRY":    "15m",
      "JWT_REFRESH_EXPIRY":   "168h",
      "RUSTFS_ENDPOINT":      "http://localhost:9000",
      "RUSTFS_ACCESS_KEY":    "rustfsadmin",
      "RUSTFS_SECRET_KEY":    "rustfsadmin",
      "RUSTFS_BUCKET":        "cms-media",
      "RUSTFS_REGION":        "us-east-1",
      "RESEND_FROM_NAME":     "Sky Flux CMS",
      "RESEND_FROM_EMAIL":    "noreply@example.com",
      "LOG_LEVEL":            "debug",
  }

  // Load reads configuration with priority: ENV vars > .env file > built-in defaults.
  // cfgFile may be empty string, in which case ".env" in the working directory is
  // attempted (failure is silently ignored — env vars alone are sufficient).
  func Load(cfgFile string) (*Config, error) {
      k := koanf.New(".")

      // Layer 1: built-in defaults (lowest priority)
      for key, val := range defaults {
          if err := k.Set(key, val); err != nil {
              return nil, fmt.Errorf("set default %s: %w", key, err)
          }
      }

      // Layer 2: .env file (optional, silently ignored if missing)
      envPath := ".env"
      if cfgFile != "" {
          envPath = cfgFile
      }
      _ = k.Load(file.Provider(envPath), dotenv.Parser())

      // Layer 3: environment variables (highest priority among non-flag sources)
      // Pass-through transformer: upper-case env vars map directly to koanf keys.
      if err := k.Load(env.Provider("", ".", func(s string) string { return s }), nil); err != nil {
          return nil, fmt.Errorf("load env vars: %w", err)
      }

      // Parse duration fields
      accessExpiry, err := time.ParseDuration(k.String("JWT_ACCESS_EXPIRY"))
      if err != nil {
          return nil, fmt.Errorf("parse JWT_ACCESS_EXPIRY: %w", err)
      }

      refreshExpiry, err := time.ParseDuration(k.String("JWT_REFRESH_EXPIRY"))
      if err != nil {
          return nil, fmt.Errorf("parse JWT_REFRESH_EXPIRY: %w", err)
      }

      connMaxLifetime, err := time.ParseDuration(k.String("DB_CONN_MAX_LIFETIME"))
      if err != nil {
          return nil, fmt.Errorf("parse DB_CONN_MAX_LIFETIME: %w", err)
      }

      connMaxIdleTime, err := time.ParseDuration(k.String("DB_CONN_MAX_IDLE_TIME"))
      if err != nil {
          return nil, fmt.Errorf("parse DB_CONN_MAX_IDLE_TIME: %w", err)
      }

      cfg := &Config{
          Server: ServerConfig{
              Port:        k.String("SERVER_PORT"),
              Mode:        k.String("SERVER_MODE"),
              FrontendURL: k.String("FRONTEND_URL"),
          },
          DB: DBConfig{
              Host:            k.String("DB_HOST"),
              Port:            k.String("DB_PORT"),
              Name:            k.String("DB_NAME"),
              User:            k.String("DB_USER"),
              Password:        k.String("DB_PASSWORD"),
              SSLMode:         k.String("DB_SSLMODE"),
              MaxOpenConns:    k.Int("DB_MAX_OPEN_CONNS"),
              MaxIdleConns:    k.Int("DB_MAX_IDLE_CONNS"),
              ConnMaxLifetime: connMaxLifetime,
              ConnMaxIdleTime: connMaxIdleTime,
          },
          Redis: RedisConfig{
              Host:     k.String("REDIS_HOST"),
              Port:     k.String("REDIS_PORT"),
              Password: k.String("REDIS_PASSWORD"),
              DB:       k.Int("REDIS_DB"),
          },
          Meili: MeiliConfig{
              URL:       k.String("MEILI_URL"),
              MasterKey: k.String("MEILI_MASTER_KEY"),
          },
          JWT: JWTConfig{
              Secret:        k.String("JWT_SECRET"),
              AccessExpiry:  accessExpiry,
              RefreshExpiry: refreshExpiry,
          },
          TOTP: TOTPConfig{
              EncryptionKey: k.String("TOTP_ENCRYPTION_KEY"),
          },
          RustFS: RustFSConfig{
              Endpoint:  k.String("RUSTFS_ENDPOINT"),
              AccessKey: k.String("RUSTFS_ACCESS_KEY"),
              SecretKey: k.String("RUSTFS_SECRET_KEY"),
              Bucket:    k.String("RUSTFS_BUCKET"),
              Region:    k.String("RUSTFS_REGION"),
          },
          Resend: ResendConfig{
              APIKey:    k.String("RESEND_API_KEY"),
              FromName:  k.String("RESEND_FROM_NAME"),
              FromEmail: k.String("RESEND_FROM_EMAIL"),
          },
          Log: LogConfig{
              Level: k.String("LOG_LEVEL"),
          },
      }

      return cfg, nil
  }
  ```

- [ ] **GREEN (step B)** — Rewrite `/Users/martinadamsdev/workspace/sky-flux-cms/internal/config/config_test.go`:

  ```go
  package config

  import (
      "os"
      "path/filepath"
      "testing"
      "time"

      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestLoad_Defaults(t *testing.T) {
      // Run in isolation: no .env file, no relevant env vars set
      cfg, err := Load("/nonexistent/.env")
      require.NoError(t, err)

      assert.Equal(t, "8080", cfg.Server.Port)
      assert.Equal(t, "debug", cfg.Server.Mode)
      assert.Equal(t, "http://localhost:3000", cfg.Server.FrontendURL)
      assert.Equal(t, "localhost", cfg.DB.Host)
      assert.Equal(t, "5432", cfg.DB.Port)
      assert.Equal(t, "disable", cfg.DB.SSLMode)
      assert.Equal(t, 25, cfg.DB.MaxOpenConns)
      assert.Equal(t, 5, cfg.DB.MaxIdleConns)
      assert.Equal(t, time.Hour, cfg.DB.ConnMaxLifetime)
      assert.Equal(t, 30*time.Minute, cfg.DB.ConnMaxIdleTime)
      assert.Equal(t, "localhost", cfg.Redis.Host)
      assert.Equal(t, "6379", cfg.Redis.Port)
      assert.Equal(t, 0, cfg.Redis.DB)
      assert.Equal(t, "http://localhost:7700", cfg.Meili.URL)
      assert.Equal(t, 15*time.Minute, cfg.JWT.AccessExpiry)
      assert.Equal(t, 168*time.Hour, cfg.JWT.RefreshExpiry)
      assert.Equal(t, "http://localhost:9000", cfg.RustFS.Endpoint)
      assert.Equal(t, "cms-media", cfg.RustFS.Bucket)
      assert.Equal(t, "debug", cfg.Log.Level)
  }

  func TestLoad_FromEnvFile(t *testing.T) {
      dir := t.TempDir()
      envFile := filepath.Join(dir, ".env")
      content := "SERVER_PORT=9090\nSERVER_MODE=release\nDB_NAME=mydb\nDB_USER=admin\nDB_PASSWORD=secret\n"
      require.NoError(t, os.WriteFile(envFile, []byte(content), 0644))

      cfg, err := Load(envFile)
      require.NoError(t, err)

      assert.Equal(t, "9090", cfg.Server.Port)
      assert.Equal(t, "release", cfg.Server.Mode)
      assert.Equal(t, "mydb", cfg.DB.Name)
      assert.Equal(t, "admin", cfg.DB.User)
      assert.Equal(t, "secret", cfg.DB.Password)
  }

  func TestLoad_EnvVarOverride(t *testing.T) {
      t.Setenv("SERVER_PORT", "7777")
      t.Setenv("DB_NAME", "override_db")

      cfg, err := Load("/nonexistent/.env")
      require.NoError(t, err)

      assert.Equal(t, "7777", cfg.Server.Port)
      assert.Equal(t, "override_db", cfg.DB.Name)
  }

  func TestLoad_InvalidDuration(t *testing.T) {
      t.Setenv("JWT_ACCESS_EXPIRY", "not-a-duration")
      t.Cleanup(func() { os.Unsetenv("JWT_ACCESS_EXPIRY") })

      _, err := Load("/nonexistent/.env")
      require.Error(t, err)
      assert.Contains(t, err.Error(), "JWT_ACCESS_EXPIRY")
  }

  func TestLoad_InvalidDBConnMaxLifetime(t *testing.T) {
      t.Setenv("DB_CONN_MAX_LIFETIME", "bad")
      t.Cleanup(func() { os.Unsetenv("DB_CONN_MAX_LIFETIME") })

      _, err := Load("/nonexistent/.env")
      require.Error(t, err)
      assert.Contains(t, err.Error(), "DB_CONN_MAX_LIFETIME")
  }

  func TestDBConfig_DSN(t *testing.T) {
      cfg := &DBConfig{
          User:     "admin",
          Password: "secret",
          Host:     "db.example.com",
          Port:     "5433",
          Name:     "cms",
          SSLMode:  "require",
      }
      expected := "postgres://admin:secret@db.example.com:5433/cms?sslmode=require"
      assert.Equal(t, expected, cfg.DSN())
  }

  func TestRedisConfig_Addr(t *testing.T) {
      cfg := &RedisConfig{Host: "redis.local", Port: "6380"}
      assert.Equal(t, "redis.local:6380", cfg.Addr())
  }
  ```

  > **Important:** `TestLoad_Defaults` passes `"/nonexistent/.env"` so koanf silently skips the file. It still picks up any env vars from the test environment. Run with `t.Parallel()` is intentionally omitted because `t.Setenv` in sibling tests would cause races.

- [ ] **Delete** the `deps_test.go` created in Task 1 (it was a temporary scaffold test):
  ```bash
  rm /Users/martinadamsdev/workspace/sky-flux-cms/internal/config/deps_test.go
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./internal/config/... -v -count=1
  # Expected: all 7 tests PASS
  ```

- [ ] **Commit**:
  ```bash
  git add internal/config/config.go internal/config/config_test.go
  git rm internal/config/deps_test.go
  git commit -m "♻️ rewrite config loader: Viper → koanf"
  ```

---

## Task 4: Chi + Huma server bootstrap

Rewrite `cmd/cms/serve.go` to wire up a Chi router and register a Huma API instance. No domain handlers yet — only the structural wiring and the health endpoint.

### Steps

- [ ] **RED** — Write an HTTP test that calls `GET /health` and expects `{"status":"ok"}`:

  Create `/Users/martinadamsdev/workspace/sky-flux-cms/cmd/cms/serve_test.go`:
  ```go
  package main

  import (
      "encoding/json"
      "net/http"
      "net/http/httptest"
      "testing"

      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestHealthEndpoint(t *testing.T) {
      // newServer() will be defined in serve.go; it returns an http.Handler
      // built from Chi + Huma without starting a listener.
      handler := newServer()

      req := httptest.NewRequest(http.MethodGet, "/health", nil)
      rec := httptest.NewRecorder()
      handler.ServeHTTP(rec, req)

      require.Equal(t, http.StatusOK, rec.Code)

      var body map[string]string
      require.NoError(t, json.NewDecoder(rec.Body).Decode(&body))
      assert.Equal(t, "ok", body["status"])
  }
  ```

- [ ] **Verify RED**:
  ```bash
  go test ./cmd/cms/... -v -count=1 -run TestHealthEndpoint
  # Expected: build failed — newServer undefined
  ```

- [ ] **GREEN** — Rewrite `/Users/martinadamsdev/workspace/sky-flux-cms/cmd/cms/serve.go`:

  ```go
  package main

  import (
      "context"
      "fmt"
      "log/slog"
      "net/http"
      "os"
      "os/signal"
      "syscall"
      "time"

      "github.com/danielgtaylor/huma/v2"
      "github.com/danielgtaylor/huma/v2/adapters/humachi"
      "github.com/go-chi/chi/v5"
      chimiddleware "github.com/go-chi/chi/v5/middleware"
      "github.com/spf13/cobra"

      "github.com/sky-flux/cms/internal/config"
  )

  var serveCmd = &cobra.Command{
      Use:   "serve",
      Short: "Start the CMS HTTP server",
      RunE:  runServe,
  }

  func init() {
      serveCmd.Flags().String("port", "", "HTTP listen port (overrides SERVER_PORT env)")
      serveCmd.Flags().String("mode", "", "Server mode: debug|release (overrides SERVER_MODE env)")
      rootCmd.AddCommand(serveCmd)
  }

  func runServe(cmd *cobra.Command, _ []string) error {
      cfgFile, _ := cmd.Root().PersistentFlags().GetString("config")
      cfg, err := config.Load(cfgFile)
      if err != nil {
          return fmt.Errorf("load config: %w", err)
      }

      // CLI flag overrides
      if p, _ := cmd.Flags().GetString("port"); p != "" {
          cfg.Server.Port = p
      }
      if m, _ := cmd.Flags().GetString("mode"); m != "" {
          cfg.Server.Mode = m
      }

      handler := newServer()

      srv := &http.Server{
          Addr:         ":" + cfg.Server.Port,
          Handler:      handler,
          ReadTimeout:  15 * time.Second,
          WriteTimeout: 15 * time.Second,
          IdleTimeout:  60 * time.Second,
      }

      ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
      defer stop()

      go func() {
          slog.Info("server starting", "addr", srv.Addr, "mode", cfg.Server.Mode)
          if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
              slog.Error("server error", "err", err)
              stop()
          }
      }()

      <-ctx.Done()
      slog.Info("shutting down gracefully")

      shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
      defer cancel()
      return srv.Shutdown(shutdownCtx)
  }

  // newServer constructs the Chi router and Huma API instance.
  // It is a separate function (not a method) so tests can call it
  // without starting a real listener.
  func newServer() http.Handler {
      r := chi.NewRouter()

      // Global middleware
      r.Use(chimiddleware.RealIP)
      r.Use(chimiddleware.RequestID)
      r.Use(chimiddleware.Recoverer)

      // Huma API on /api/v1
      api := humachi.New(r, huma.DefaultConfig("Sky Flux CMS API", "1.0.0"))

      // Health check — registered directly on Huma so it appears in OpenAPI spec
      huma.Register(api, huma.Operation{
          OperationID: "health-check",
          Method:      http.MethodGet,
          Path:        "/health",
          Summary:     "Health check",
          Tags:        []string{"system"},
      }, func(ctx context.Context, _ *struct{}) (*struct {
          Body struct {
              Status string `json:"status"`
          }
      }, error) {
          resp := &struct {
              Body struct {
                  Status string `json:"status"`
              }
          }{}
          resp.Body.Status = "ok"
          return resp, nil
      })

      return r
  }
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./cmd/cms/... -v -count=1 -run TestHealthEndpoint
  # Expected: PASS
  go build ./cmd/cms/
  # Expected: binary built without errors
  ```

- [ ] **Commit**:
  ```bash
  git add cmd/cms/serve.go cmd/cms/serve_test.go
  git commit -m "✨ bootstrap Chi router + Huma v2 API with /health endpoint"
  ```

---

## Task 5: Migration cleanup

**Delete** the old multi-site placeholder migration (migration 3) and renumber accordingly. **Write** two new migrations:
- Migration 4 (renumbered from old 4 seed): keep the RBAC seed as-is — it already has the correct numbering relative to migrations 1 and 2 after deleting migration 3.
- New migration 4 (content tables): `sfc_posts`, `sfc_categories`, `sfc_tags`, `sfc_media_files`, `sfc_comments`, `sfc_menus`, `sfc_menu_items`, `sfc_redirects`, `sfc_audits` — all in `public` schema, no `site_id`, `sfc_` prefix.
- New migration 5 (indexes + constraints + triggers).

### Current state of migrations/

```
20260224000001_create_core_tables.go    ← keep
20260224000002_create_rbac_tables.go    ← keep
20260224000003_site_schema_placeholder.go ← DELETE
20260224000004_seed_rbac_builtins.go    ← keep (becomes effective migration 3 after deletion)
20260224000005_boolean_to_smallint.go   ← keep (becomes effective migration 4)
20260225000006_add_menu_columns.go      ← keep (becomes effective migration 5)
```

After deletion, bun renumbers by timestamp naturally — the `init()` registration order in `migrations/main.go` determines sequence. Deleting the placeholder file removes its `init()` from compilation.

**New files to add:**
- `20260319000007_content_tables.go` — content tables in public schema
- `20260319000008_content_indexes.go` — indexes, triggers, and constraints

### Steps

- [ ] **RED** — Write a compile test that checks the placeholder is gone and new migrations compile:

  Create `/Users/martinadamsdev/workspace/sky-flux-cms/migrations/migrations_test.go`:
  ```go
  package migrations

  import (
      "testing"
  )

  // TestMigrationsCount verifies we have the expected number of registered
  // migrations. The site_schema_placeholder (old migration 3) must be deleted.
  // After Phase 0 additions: 1 (core) + 2 (rbac) + 3 (seed) + 4 (bool→smallint)
  // + 5 (menu columns) + 6 (content tables) + 7 (content indexes) = 7 migrations.
  func TestMigrationsCount(t *testing.T) {
      got := len(Migrations.Sorted())
      want := 7
      if got != want {
          t.Errorf("expected %d migrations, got %d — did you forget to delete the placeholder or add new files?", want, got)
      }
  }
  ```

- [ ] **Verify RED**:
  ```bash
  go test ./migrations/... -v -count=1 -run TestMigrationsCount
  # Expected: FAIL — got 6 (or 7 if placeholder still counted), want 7
  # The exact count depends on whether placeholder is still present.
  # Either way, confirm the number is NOT 7 before proceeding.
  ```

- [ ] **GREEN (step A)** — Delete the placeholder migration:
  ```bash
  rm /Users/martinadamsdev/workspace/sky-flux-cms/migrations/20260224000003_site_schema_placeholder.go
  ```

- [ ] **GREEN (step B)** — Create `/Users/martinadamsdev/workspace/sky-flux-cms/migrations/20260319000007_content_tables.go`:

  ```go
  package migrations

  import (
      "context"
      "fmt"

      "github.com/uptrace/bun"
  )

  func init() {
      Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
          _, err := db.ExecContext(ctx, `
  -- v1 content tables: all in public schema, sfc_ prefix, no site_id.
  -- v2 multi-site will move these to site_{slug} schemas.

  -- Categories (tree via materialized path)
  CREATE TABLE public.sfc_categories (
      id          UUID PRIMARY KEY DEFAULT uuidv7(),
      name        VARCHAR(200) NOT NULL,
      slug        VARCHAR(200) NOT NULL UNIQUE,
      description TEXT,
      parent_id   UUID REFERENCES public.sfc_categories(id) ON DELETE SET NULL,
      path        TEXT NOT NULL DEFAULT '',   -- materialized path: /uuid/uuid/
      depth       SMALLINT NOT NULL DEFAULT 0,
      sort_order  INTEGER NOT NULL DEFAULT 0,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at  TIMESTAMPTZ
  );

  -- Tags (flat)
  CREATE TABLE public.sfc_tags (
      id         UUID PRIMARY KEY DEFAULT uuidv7(),
      name       VARCHAR(100) NOT NULL,
      slug       VARCHAR(100) NOT NULL UNIQUE,
      created_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at TIMESTAMPTZ
  );

  -- Posts (state machine: draft → published → archived)
  CREATE TYPE post_status AS ENUM ('draft', 'published', 'archived', 'scheduled');
  CREATE TYPE post_type   AS ENUM ('article', 'page');

  CREATE TABLE public.sfc_posts (
      id               UUID PRIMARY KEY DEFAULT uuidv7(),
      title            VARCHAR(500) NOT NULL,
      slug             VARCHAR(500) NOT NULL UNIQUE,
      excerpt          TEXT,
      content          TEXT,
      content_json     JSONB,
      cover_image_url  TEXT,
      status           post_status NOT NULL DEFAULT 'draft',
      type             post_type NOT NULL DEFAULT 'article',
      author_id        UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
      category_id      UUID REFERENCES public.sfc_categories(id) ON DELETE SET NULL,
      published_at     TIMESTAMPTZ,
      scheduled_at     TIMESTAMPTZ,
      view_count       INTEGER NOT NULL DEFAULT 0,
      comment_count    INTEGER NOT NULL DEFAULT 0,
      is_featured      BOOLEAN NOT NULL DEFAULT FALSE,
      allow_comments   BOOLEAN NOT NULL DEFAULT TRUE,
      meta_title       VARCHAR(500),
      meta_description TEXT,
      created_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at       TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at       TIMESTAMPTZ
  );

  -- Post ↔ Tag many-to-many
  CREATE TABLE public.sfc_post_tags (
      post_id UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
      tag_id  UUID NOT NULL REFERENCES public.sfc_tags(id) ON DELETE CASCADE,
      PRIMARY KEY (post_id, tag_id)
  );

  -- Post revisions (immutable history)
  CREATE TABLE public.sfc_post_revisions (
      id           UUID PRIMARY KEY DEFAULT uuidv7(),
      post_id      UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
      title        VARCHAR(500) NOT NULL,
      content      TEXT,
      content_json JSONB,
      author_id    UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
      created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- Media files
  CREATE TYPE media_status AS ENUM ('active', 'deleted');

  CREATE TABLE public.sfc_media_files (
      id              UUID PRIMARY KEY DEFAULT uuidv7(),
      filename        VARCHAR(500) NOT NULL,
      original_name   VARCHAR(500) NOT NULL,
      mime_type       VARCHAR(100) NOT NULL,
      size_bytes      BIGINT NOT NULL,
      storage_key     TEXT NOT NULL UNIQUE,  -- RustFS object key
      url             TEXT NOT NULL,
      width           INTEGER,
      height          INTEGER,
      thumb_sm_url    TEXT,                  -- 150×150 crop
      thumb_md_url    TEXT,                  -- 400×400 fit
      alt_text        TEXT,
      caption         TEXT,
      uploader_id     UUID NOT NULL REFERENCES public.sfc_users(id) ON DELETE RESTRICT,
      status          media_status NOT NULL DEFAULT 'active',
      created_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at      TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      deleted_at      TIMESTAMPTZ
  );

  -- Comments (3-level nesting max, moderated)
  CREATE TYPE comment_status AS ENUM ('pending', 'approved', 'spam', 'rejected');

  CREATE TABLE public.sfc_comments (
      id           UUID PRIMARY KEY DEFAULT uuidv7(),
      post_id      UUID NOT NULL REFERENCES public.sfc_posts(id) ON DELETE CASCADE,
      parent_id    UUID REFERENCES public.sfc_comments(id) ON DELETE CASCADE,
      depth        SMALLINT NOT NULL DEFAULT 0 CHECK (depth <= 2),
      author_name  VARCHAR(100) NOT NULL,
      author_email VARCHAR(255) NOT NULL,
      author_url   TEXT,
      author_ip    INET,
      content      TEXT NOT NULL,
      status       comment_status NOT NULL DEFAULT 'pending',
      is_pinned    BOOLEAN NOT NULL DEFAULT FALSE,
      is_admin     BOOLEAN NOT NULL DEFAULT FALSE,  -- TRUE = admin reply
      created_at   TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at   TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- Navigation menus
  CREATE TABLE public.sfc_menus (
      id          UUID PRIMARY KEY DEFAULT uuidv7(),
      name        VARCHAR(200) NOT NULL,
      slug        VARCHAR(200) NOT NULL UNIQUE,
      description TEXT,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  CREATE TABLE public.sfc_menu_items (
      id          UUID PRIMARY KEY DEFAULT uuidv7(),
      menu_id     UUID NOT NULL REFERENCES public.sfc_menus(id) ON DELETE CASCADE,
      parent_id   UUID REFERENCES public.sfc_menu_items(id) ON DELETE CASCADE,
      depth       SMALLINT NOT NULL DEFAULT 0 CHECK (depth <= 2),
      label       VARCHAR(200) NOT NULL,
      url         TEXT NOT NULL,
      target      VARCHAR(20) NOT NULL DEFAULT '_self',
      icon        VARCHAR(50),
      css_class   VARCHAR(100),
      sort_order  INTEGER NOT NULL DEFAULT 0,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- URL redirects
  CREATE TYPE redirect_type AS ENUM ('301', '302');

  CREATE TABLE public.sfc_redirects (
      id          UUID PRIMARY KEY DEFAULT uuidv7(),
      from_path   TEXT NOT NULL UNIQUE,
      to_path     TEXT NOT NULL,
      type        redirect_type NOT NULL DEFAULT '301',
      hit_count   INTEGER NOT NULL DEFAULT 0,
      is_active   BOOLEAN NOT NULL DEFAULT TRUE,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW(),
      updated_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );

  -- Audit log (immutable)
  CREATE TABLE public.sfc_audits (
      id          UUID PRIMARY KEY DEFAULT uuidv7(),
      user_id     UUID REFERENCES public.sfc_users(id) ON DELETE SET NULL,
      action      VARCHAR(100) NOT NULL,
      entity_type VARCHAR(100) NOT NULL,
      entity_id   UUID,
      meta        JSONB NOT NULL DEFAULT '{}',
      ip_address  INET,
      user_agent  TEXT,
      created_at  TIMESTAMPTZ NOT NULL DEFAULT NOW()
  );
          `)
          if err != nil {
              return fmt.Errorf("create content tables: %w", err)
          }
          return nil
      }, func(ctx context.Context, db *bun.DB) error {
          _, err := db.ExecContext(ctx, `
  DROP TABLE IF EXISTS public.sfc_audits CASCADE;
  DROP TABLE IF EXISTS public.sfc_redirects CASCADE;
  DROP TABLE IF EXISTS public.sfc_menu_items CASCADE;
  DROP TABLE IF EXISTS public.sfc_menus CASCADE;
  DROP TABLE IF EXISTS public.sfc_comments CASCADE;
  DROP TABLE IF EXISTS public.sfc_media_files CASCADE;
  DROP TABLE IF EXISTS public.sfc_post_revisions CASCADE;
  DROP TABLE IF EXISTS public.sfc_post_tags CASCADE;
  DROP TABLE IF EXISTS public.sfc_posts CASCADE;
  DROP TABLE IF EXISTS public.sfc_tags CASCADE;
  DROP TABLE IF EXISTS public.sfc_categories CASCADE;
  DROP TYPE IF EXISTS redirect_type CASCADE;
  DROP TYPE IF EXISTS comment_status CASCADE;
  DROP TYPE IF EXISTS media_status CASCADE;
  DROP TYPE IF EXISTS post_type CASCADE;
  DROP TYPE IF EXISTS post_status CASCADE;
          `)
          if err != nil {
              return fmt.Errorf("drop content tables: %w", err)
          }
          return nil
      })
  }
  ```

- [ ] **GREEN (step C)** — Create `/Users/martinadamsdev/workspace/sky-flux-cms/migrations/20260319000008_content_indexes.go`:

  ```go
  package migrations

  import (
      "context"
      "fmt"

      "github.com/uptrace/bun"
  )

  func init() {
      Migrations.MustRegister(func(ctx context.Context, db *bun.DB) error {
          _, err := db.ExecContext(ctx, `
  -- sfc_posts indexes
  CREATE INDEX idx_sfc_posts_status     ON public.sfc_posts(status) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_posts_author     ON public.sfc_posts(author_id) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_posts_category   ON public.sfc_posts(category_id) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_posts_published  ON public.sfc_posts(published_at DESC) WHERE status = 'published' AND deleted_at IS NULL;
  CREATE INDEX idx_sfc_posts_scheduled  ON public.sfc_posts(scheduled_at) WHERE status = 'scheduled' AND deleted_at IS NULL;
  CREATE INDEX idx_sfc_posts_featured   ON public.sfc_posts(is_featured) WHERE is_featured = TRUE AND deleted_at IS NULL;

  -- sfc_categories indexes
  CREATE INDEX idx_sfc_cats_parent      ON public.sfc_categories(parent_id) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_cats_path        ON public.sfc_categories(path) WHERE deleted_at IS NULL;

  -- sfc_tags indexes
  CREATE INDEX idx_sfc_tags_slug        ON public.sfc_tags(slug) WHERE deleted_at IS NULL;

  -- sfc_post_tags indexes
  CREATE INDEX idx_sfc_pt_tag_id        ON public.sfc_post_tags(tag_id);

  -- sfc_post_revisions indexes
  CREATE INDEX idx_sfc_revisions_post   ON public.sfc_post_revisions(post_id, created_at DESC);

  -- sfc_media_files indexes
  CREATE INDEX idx_sfc_media_uploader   ON public.sfc_media_files(uploader_id) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_media_mime       ON public.sfc_media_files(mime_type) WHERE deleted_at IS NULL;
  CREATE INDEX idx_sfc_media_status     ON public.sfc_media_files(status) WHERE deleted_at IS NULL;

  -- sfc_comments indexes
  CREATE INDEX idx_sfc_comments_post    ON public.sfc_comments(post_id, status);
  CREATE INDEX idx_sfc_comments_parent  ON public.sfc_comments(parent_id) WHERE parent_id IS NOT NULL;
  CREATE INDEX idx_sfc_comments_status  ON public.sfc_comments(status);
  CREATE INDEX idx_sfc_comments_email   ON public.sfc_comments(author_email);

  -- sfc_menu_items indexes
  CREATE INDEX idx_sfc_mi_menu         ON public.sfc_menu_items(menu_id, sort_order);
  CREATE INDEX idx_sfc_mi_parent       ON public.sfc_menu_items(parent_id) WHERE parent_id IS NOT NULL;

  -- sfc_redirects indexes
  CREATE INDEX idx_sfc_redirects_from  ON public.sfc_redirects(from_path) WHERE is_active = TRUE;

  -- sfc_audits indexes
  CREATE INDEX idx_sfc_audits_user     ON public.sfc_audits(user_id) WHERE user_id IS NOT NULL;
  CREATE INDEX idx_sfc_audits_entity   ON public.sfc_audits(entity_type, entity_id) WHERE entity_id IS NOT NULL;
  CREATE INDEX idx_sfc_audits_created  ON public.sfc_audits(created_at DESC);

  -- updated_at auto-update trigger for content tables
  -- (update_updated_at() function created in migration 1)
  CREATE TRIGGER trg_sfc_categories_updated_at
      BEFORE UPDATE ON public.sfc_categories
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_tags_updated_at
      BEFORE UPDATE ON public.sfc_tags
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_posts_updated_at
      BEFORE UPDATE ON public.sfc_posts
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_media_files_updated_at
      BEFORE UPDATE ON public.sfc_media_files
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_comments_updated_at
      BEFORE UPDATE ON public.sfc_comments
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_menus_updated_at
      BEFORE UPDATE ON public.sfc_menus
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_menu_items_updated_at
      BEFORE UPDATE ON public.sfc_menu_items
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();

  CREATE TRIGGER trg_sfc_redirects_updated_at
      BEFORE UPDATE ON public.sfc_redirects
      FOR EACH ROW EXECUTE FUNCTION update_updated_at();
          `)
          if err != nil {
              return fmt.Errorf("create content indexes and triggers: %w", err)
          }
          return nil
      }, func(ctx context.Context, db *bun.DB) error {
          _, err := db.ExecContext(ctx, `
  -- Triggers drop with their tables in migration 7 rollback.
  -- Explicit index drops here in case partial rollback is needed.
  DROP INDEX IF EXISTS public.idx_sfc_redirects_from;
  DROP INDEX IF EXISTS public.idx_sfc_mi_parent;
  DROP INDEX IF EXISTS public.idx_sfc_mi_menu;
  DROP INDEX IF EXISTS public.idx_sfc_audits_created;
  DROP INDEX IF EXISTS public.idx_sfc_audits_entity;
  DROP INDEX IF EXISTS public.idx_sfc_audits_user;
  DROP INDEX IF EXISTS public.idx_sfc_comments_email;
  DROP INDEX IF EXISTS public.idx_sfc_comments_status;
  DROP INDEX IF EXISTS public.idx_sfc_comments_parent;
  DROP INDEX IF EXISTS public.idx_sfc_comments_post;
  DROP INDEX IF EXISTS public.idx_sfc_media_status;
  DROP INDEX IF EXISTS public.idx_sfc_media_mime;
  DROP INDEX IF EXISTS public.idx_sfc_media_uploader;
  DROP INDEX IF EXISTS public.idx_sfc_revisions_post;
  DROP INDEX IF EXISTS public.idx_sfc_pt_tag_id;
  DROP INDEX IF EXISTS public.idx_sfc_tags_slug;
  DROP INDEX IF EXISTS public.idx_sfc_cats_path;
  DROP INDEX IF EXISTS public.idx_sfc_cats_parent;
  DROP INDEX IF EXISTS public.idx_sfc_posts_featured;
  DROP INDEX IF EXISTS public.idx_sfc_posts_scheduled;
  DROP INDEX IF EXISTS public.idx_sfc_posts_published;
  DROP INDEX IF EXISTS public.idx_sfc_posts_category;
  DROP INDEX IF EXISTS public.idx_sfc_posts_author;
  DROP INDEX IF EXISTS public.idx_sfc_posts_status;
          `)
          if err != nil {
              return fmt.Errorf("drop content indexes: %w", err)
          }
          return nil
      })
  }
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./migrations/... -v -count=1 -run TestMigrationsCount
  # Expected: PASS — got 7 migrations
  go build ./migrations/
  # Expected: no compile errors
  ```

- [ ] **Commit**:
  ```bash
  git rm migrations/20260224000003_site_schema_placeholder.go
  git add migrations/20260319000007_content_tables.go \
          migrations/20260319000008_content_indexes.go \
          migrations/migrations_test.go
  git commit -m "🗄️ replace multi-site placeholder with v1 single-site content tables"
  ```

---

## Task 6: embed.go + console/dist/.gitkeep

Create the root `embed.go` file and the `console/dist/.gitkeep` placeholder so `go:embed` compiles without requiring a pre-built console.

### Steps

- [ ] **RED** — Write a test that verifies `ConsoleFS` and `WebStaticFS` are accessible:

  Create `/Users/martinadamsdev/workspace/sky-flux-cms/embed_test.go`:
  ```go
  package cms_test

  import (
      "io/fs"
      "testing"

      cms "github.com/sky-flux/cms"
      "github.com/stretchr/testify/assert"
      "github.com/stretchr/testify/require"
  )

  func TestConsoleFSReadable(t *testing.T) {
      // console/dist/.gitkeep must exist for go:embed to compile.
      // This test verifies the embedded FS is accessible at runtime.
      _, err := fs.Stat(cms.ConsoleFS, "console/dist/.gitkeep")
      require.NoError(t, err, "console/dist/.gitkeep must be embedded — run: mkdir -p console/dist && touch console/dist/.gitkeep")
  }

  func TestWebStaticFSReadable(t *testing.T) {
      _, err := fs.Stat(cms.WebStaticFS, "web/static/.gitkeep")
      require.NoError(t, err, "web/static/.gitkeep must be embedded — run: mkdir -p web/static && touch web/static/.gitkeep")
  }
  ```

- [ ] **Verify RED**:
  ```bash
  go test ./... -v -count=1 -run "TestConsoleFSReadable|TestWebStaticFSReadable"
  # Expected: build failed — package cms (root package) not found / embed.go missing
  ```

- [ ] **GREEN (step A)** — Create placeholder files:
  ```bash
  mkdir -p /Users/martinadamsdev/workspace/sky-flux-cms/console/dist
  touch /Users/martinadamsdev/workspace/sky-flux-cms/console/dist/.gitkeep

  mkdir -p /Users/martinadamsdev/workspace/sky-flux-cms/web/static
  touch /Users/martinadamsdev/workspace/sky-flux-cms/web/static/.gitkeep
  ```

- [ ] **GREEN (step B)** — Create `/Users/martinadamsdev/workspace/sky-flux-cms/embed.go`:

  ```go
  // Package cms is the root package of the Sky Flux CMS module.
  // Its sole responsibility is to expose embedded file systems for the
  // compiled console SPA and web static assets, so cmd/cms/main.go can
  // import them via a single import path.
  //
  // Build: ensure console/dist is populated before `go build`:
  //   cd console && bun run build
  //
  // Development: console/dist/.gitkeep is committed so `go:embed` compiles.
  // The Chi router detects --dev flag and proxies /console/* to Vite :3000.
  package cms

  import "embed"

  // ConsoleFS holds the React admin SPA production build.
  // In development, the Go server proxies /console/* to Vite.
  //
  //go:embed all:console/dist
  var ConsoleFS embed.FS

  // WebStaticFS holds compiled Tailwind CSS and HTMX for the public site.
  //
  //go:embed all:web/static
  var WebStaticFS embed.FS
  ```

- [ ] **Verify GREEN**:
  ```bash
  go test ./... -v -count=1 -run "TestConsoleFSReadable|TestWebStaticFSReadable"
  # Expected: PASS — both FS tests pass
  go build ./...
  # Expected: entire module builds without errors
  ```

- [ ] **Commit**:
  ```bash
  git add embed.go embed_test.go console/dist/.gitkeep web/static/.gitkeep
  git commit -m "✨ add root embed.go for console/dist and web/static"
  ```

---

## Final verification

After all 6 tasks, run the full test suite to confirm nothing is broken:

```bash
go test ./... -count=1
# Expected: all packages PASS
# Packages with integration tests that need Docker will be skipped with -short if needed:
go test ./... -short -count=1
```

Check that the binary builds cleanly:

```bash
go build -o /tmp/cms-phase0 ./cmd/cms/
/tmp/cms-phase0 --help
# Expected: Cobra help output with serve, migrate, version subcommands
```

---

## Dependency notes

| Package | Version pin | Rationale |
|---------|-------------|-----------|
| `github.com/go-chi/chi/v5` | `@latest` (≥ 5.2.0) | Stable v5 API, no breaking changes expected |
| `github.com/danielgtaylor/huma/v2` | `@latest` (≥ 2.27.0) | Huma v2 for Chi adapter: `humachi.New` |
| `github.com/knadh/koanf/v2` | `@latest` (≥ 2.1.0) | Core koanf v2 |
| `github.com/knadh/koanf/providers/env` | `@latest` | Environment variable provider |
| `github.com/knadh/koanf/providers/file` | `@latest` | File provider for .env loading |
| `github.com/knadh/koanf/parsers/dotenv` | `@latest` | .env file parser |

**koanf module paths:** All koanf sub-packages (`providers/env`, `providers/file`, `parsers/dotenv`) are separate Go modules under the `github.com/knadh/koanf` umbrella — each requires its own `go get` call.

**Huma Chi adapter import:** `github.com/danielgtaylor/huma/v2/adapters/humachi` — this sub-package is part of the huma v2 module, not a separate module.

---

## Rollback

If any task fails and cannot be resolved:

1. `git revert HEAD` — revert the failing commit
2. Fix the root cause (check go.mod, missing files, import paths)
3. Re-run from the failing step's **RED** phase

Do **not** skip the RED verification step — it is the only proof that the test would catch a regression.
