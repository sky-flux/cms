# Global Endpoints — Batch 1 Design: Setup + Auth

**Date**: 2026-02-24
**Scope**: 19 endpoints (Setup 2 + Auth 17)
**Architecture**: Handler → Service → Repository (三层)

---

## 1. Overview

Batch 1 implements the authentication foundation for Sky Flux CMS:
- Installation wizard (check + initialize)
- Full auth flow (login, refresh, logout, profile, password)
- 2FA TOTP support (setup, verify, validate, disable, backup codes, status, force-disable)

### Dependencies on Existing Code
- Models: User, RefreshToken, PasswordResetToken, UserTOTP, Config, Site, UserRole (all exist)
- Middleware: Recovery, RequestID, Logger, CORS (exist); Auth, InstallationGuard (to implement)
- RBAC Service: CheckPermission (exists, used by RBAC middleware)
- Schema: CreateSiteSchema (exists)
- Utilities: apperror, response (exist)

---

## 2. New Files

### Infrastructure Layer

| File | Purpose |
|------|---------|
| `internal/pkg/jwt/jwt.go` | JWT sign (access + temp_2fa), verify, blacklist check via Redis |
| `internal/pkg/jwt/jwt_test.go` | Unit tests |
| `internal/pkg/crypto/password.go` | bcrypt hash/compare (cost=12) |
| `internal/pkg/crypto/totp.go` | TOTP secret AES-256-GCM encrypt/decrypt, TOTP verify, backup code gen/match |
| `internal/pkg/crypto/token.go` | Secure random token generation + SHA-256 hash |
| `internal/pkg/crypto/*_test.go` | Unit tests |

### Middleware Layer

| File | Purpose |
|------|---------|
| `internal/middleware/auth.go` | JWT auth middleware: extract Bearer → verify → check blacklist → set user_id in context |
| `internal/middleware/auth_test.go` | Unit tests |
| `internal/middleware/installation_guard.go` | Installation guard: atomic → Redis → DB triple check |
| `internal/middleware/installation_guard_test.go` | Unit tests |

### Setup Module

| File | Purpose |
|------|---------|
| `internal/setup/handler.go` | HTTP handlers: Check, Initialize |
| `internal/setup/service.go` | Business logic: triple-cache check, transactional initialization |
| `internal/setup/repository.go` | Data access: sfc_configs read/write |
| `internal/setup/dto.go` | InitializeReq DTO with validation tags |
| `internal/setup/handler_test.go` | HTTP handler tests |

### Auth Module

| File | Purpose |
|------|---------|
| `internal/auth/handler.go` | HTTP handlers: 17 endpoint methods |
| `internal/auth/service.go` | Business logic: login flow, token mgmt, 2FA, password reset |
| `internal/auth/repository.go` | Data access: user queries, token CRUD, TOTP CRUD |
| `internal/auth/interfaces.go` | Repository interface definitions |
| `internal/auth/dto.go` | Request/Response DTOs with validation |
| `internal/auth/service_test.go` | Service unit tests (mock repo) |
| `internal/auth/handler_test.go` | Handler HTTP tests (mock service) |

### Router Update

| File | Change |
|------|--------|
| `internal/router/router.go` | Register setup + auth route groups with DI |

---

## 3. Endpoint Mapping

### Setup (2 endpoints, no auth)
```
POST /api/v1/setup/check       → setup.Handler.Check
POST /api/v1/setup/initialize  → setup.Handler.Initialize
```

### Auth Core (9 endpoints)
```
POST /api/v1/auth/login            → auth.Handler.Login           [public]
POST /api/v1/auth/refresh          → auth.Handler.Refresh         [cookie]
POST /api/v1/auth/logout           → auth.Handler.Logout          [JWT]
GET  /api/v1/auth/me               → auth.Handler.Me              [JWT]
PUT  /api/v1/auth/password         → auth.Handler.ChangePassword  [JWT]
POST /api/v1/auth/forgot-password  → auth.Handler.ForgotPassword  [public]
POST /api/v1/auth/reset-password   → auth.Handler.ResetPassword   [public]
```

### Auth 2FA (8 endpoints)
```
POST   /api/v1/auth/2fa/setup          → auth.Handler.Setup2FA             [JWT]
POST   /api/v1/auth/2fa/verify         → auth.Handler.Verify2FA            [JWT]
POST   /api/v1/auth/2fa/validate       → auth.Handler.Validate2FA          [temp_token]
POST   /api/v1/auth/2fa/disable        → auth.Handler.Disable2FA           [JWT]
POST   /api/v1/auth/2fa/backup-codes   → auth.Handler.RegenerateBackupCodes [JWT]
GET    /api/v1/auth/2fa/status          → auth.Handler.Get2FAStatus         [JWT]
DELETE /api/v1/auth/2fa/users/:user_id  → auth.Handler.ForceDisable2FA      [JWT+Super]
```

---

## 4. Key Business Flows

### Login Flow
1. Validate request body → query user by email
2. Check login lockout (Redis `login_fail:{email}`, 5 attempts / 15 min)
3. bcrypt compare password → on failure: INCR counter
4. On success → reset counter → check 2FA status
5. No 2FA → issue access_token + refresh_token (httpOnly cookie)
6. Has 2FA → issue temp_token (purpose=2fa_verification, 5min TTL)

### JWT Token Design
- **Access Token Claims**: `{ sub: user_id, jti: token_id, iat, exp }`
- **Temp 2FA Token Claims**: `{ sub: user_id, jti: token_id, purpose: "2fa_verification", iat, exp }`
- **No role in JWT** — RBAC middleware resolves dynamically

### Installation Guard (triple check)
```
atomic.LoadInt32(&installed) == 1  →  pass through
Redis GET system:installed == "true"  →  update atomic → pass
DB SELECT sfc_configs WHERE key='system.installed'  →  update Redis + atomic
```

### Setup Initialize (single transaction)
1. Verify not installed (advisory lock)
2. Create admin user (bcrypt password)
3. Create site record
4. Create site schema (via schema.CreateSiteSchema)
5. Assign super role to admin
6. Set system.installed = true in sfc_configs
7. COMMIT → update Redis + atomic flag
8. Issue JWT for admin

### Password Reset Flow
1. ForgotPassword: generate random token → SHA-256 hash → store in sfc_password_reset_tokens (30min TTL) → send email via Resend
2. ResetPassword: verify token hash → check expiry → update password → mark token used → revoke all refresh tokens

### 2FA Setup Flow
1. Setup2FA: generate TOTP secret → AES-256-GCM encrypt → store in sfc_user_totp (is_enabled=false) → generate 10 backup codes → return secret + QR URI + backup codes
2. Verify2FA: validate TOTP code against secret → set is_enabled=true + verified_at
3. Validate2FA (login): verify temp_token purpose → validate TOTP/backup code → issue access + refresh tokens
4. Disable2FA: verify password + TOTP code → delete sfc_user_totp record → revoke all refresh tokens

---

## 5. Interface Definitions

### Auth Repository Interface
```go
type UserRepository interface {
    GetByEmail(ctx context.Context, email string) (*model.User, error)
    GetByID(ctx context.Context, id string) (*model.User, error)
    UpdatePassword(ctx context.Context, id, passwordHash string) error
    UpdateLastLogin(ctx context.Context, id string) error
    Create(ctx context.Context, user *model.User) error
}

type TokenRepository interface {
    CreateRefreshToken(ctx context.Context, token *model.RefreshToken) error
    GetRefreshTokenByHash(ctx context.Context, hash string) (*model.RefreshToken, error)
    RevokeRefreshToken(ctx context.Context, id string) error
    RevokeAllUserTokens(ctx context.Context, userID string) error
    CreatePasswordResetToken(ctx context.Context, token *model.PasswordResetToken) error
    GetPasswordResetTokenByHash(ctx context.Context, hash string) (*model.PasswordResetToken, error)
    MarkPasswordResetTokenUsed(ctx context.Context, id string) error
}

type TOTPRepository interface {
    GetByUserID(ctx context.Context, userID string) (*model.UserTOTP, error)
    Upsert(ctx context.Context, totp *model.UserTOTP) error
    Enable(ctx context.Context, id string) error
    Delete(ctx context.Context, userID string) error
    UpdateBackupCodes(ctx context.Context, id string, codes []string) error
}
```

### Setup Repository Interface
```go
type ConfigRepository interface {
    GetByKey(ctx context.Context, key string) (*model.Config, error)
    SetByKey(ctx context.Context, key string, value interface{}) error
}
```

---

## 6. Testing Strategy

| Package | Test Type | Key Scenarios |
|---------|-----------|---------------|
| pkg/jwt | Unit | sign/verify, expired token, blacklisted token, invalid signature, temp token purpose |
| pkg/crypto | Unit | bcrypt hash/verify, TOTP encrypt/decrypt/validate, token generation |
| middleware/auth | Unit | valid token, missing header, expired token, blacklisted JTI |
| middleware/installation_guard | Unit | atomic hit, Redis hit, DB hit, not installed redirect |
| setup/handler | HTTP | check status, successful init, already installed 409, validation errors |
| auth/service | Unit (mock repo) | login success, login lockout, 2FA flow, password reset flow |
| auth/handler | HTTP (mock svc) | all 17 endpoints happy path + error cases |

---

## 7. Implementation Order

```
Phase 1: Infrastructure
  1. pkg/crypto (password, totp, token) + tests
  2. pkg/jwt (sign, verify, blacklist) + tests

Phase 2: Middleware
  3. middleware/installation_guard + tests
  4. middleware/auth + tests

Phase 3: Setup Module
  5. setup/dto + repository + service + handler + tests

Phase 4: Auth Module
  6. auth/dto + interfaces
  7. auth/repository
  8. auth/service + tests
  9. auth/handler + tests

Phase 5: Router Integration
  10. router.go update — register all routes with DI
  11. Smoke test — verify all endpoints respond correctly
```

---

## 8. Future Batches

| Batch | Scope | Depends On |
|-------|-------|------------|
| Batch 2 | Sites management (8 endpoints) | Batch 1 auth middleware |
| Batch 3 | RBAC audit + route registration (20 endpoints) | Batch 1+2 middleware chain |
| Batch 4 | Integration tests & E2E verification | All batches |
