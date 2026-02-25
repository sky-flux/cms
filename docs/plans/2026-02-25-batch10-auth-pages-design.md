# Batch 10: Auth Pages Design

> Status: Approved
> Date: 2026-02-25
> Scope: Login, 2FA, Forgot Password, Reset Password, Setup Wizard

## Overview

Batch 10 implements all authentication-related frontend pages for Sky Flux CMS.
Backend API is 100% complete (14 auth + 2 setup endpoints). This batch creates
the React Islands that consume those endpoints.

## Pages & Routes

| Page | Route | Backend Endpoint | Auth |
|------|-------|-----------------|------|
| Login | `/login` | `POST /api/v1/auth/login` | Public |
| 2FA Verify | `/login/2fa` | `POST /api/v1/auth/2fa/verify` | temp_token |
| Forgot Password | `/forgot-password` | `POST /api/v1/auth/forgot-password` | Public |
| Email Sent | `/forgot-password/sent` | — | Public |
| Reset Password | `/reset-password?token=xxx` | `POST /api/v1/auth/reset-password` | token param |
| Setup Wizard | `/setup` | `POST /api/v1/setup/install` | InstallGuard |
| Setup Complete | `/setup/complete` | — | Public |

## Layout

**AuthLayout** — centered card on neutral background.

```
┌─────────────────────────────────────┐
│           (bg-background)           │
│                                     │
│      ┌───────────────────────┐      │
│      │   [Logo] Sky Flux CMS │      │
│      │                       │      │
│      │   <form content>      │      │
│      │                       │      │
│      │   [link: forgot pwd]  │      │
│      └───────────────────────┘      │
│                                     │
│     © 2026 Sky Flux CMS             │
└─────────────────────────────────────┘
```

- Card width: `max-w-md` (login/reset), `max-w-lg` (setup wizard)
- Logo: Lucide `Sparkles` icon + "Sky Flux CMS" text
- Footer: copyright + language switcher (LocaleSwitcher component)
- Dark mode: automatic via `.dark` class on `<html>`

## File Structure

```
src/
├── layouts/
│   └── AuthLayout.astro              # Centered card layout (NEW)
├── pages/
│   ├── login.astro                   # Login page (REWRITE)
│   ├── login/
│   │   └── 2fa.astro                 # 2FA verification
│   ├── forgot-password/
│   │   ├── index.astro               # Forgot password form
│   │   └── sent.astro                # "Check your email" page
│   ├── reset-password.astro          # Reset password form
│   └── setup/
│       ├── index.astro               # 3-step wizard
│       └── complete.astro            # Setup complete
├── components/auth/
│   ├── LoginForm.tsx                 # Email + password form
│   ├── TwoFactorForm.tsx            # 6-digit OTP input
│   ├── ForgotPasswordForm.tsx       # Email input
│   ├── ResetPasswordForm.tsx        # New password + confirm
│   └── SetupWizard.tsx              # 3-step wizard form
│   └── __tests__/
│       ├── LoginForm.test.tsx
│       ├── TwoFactorForm.test.tsx
│       ├── ForgotPasswordForm.test.tsx
│       ├── ResetPasswordForm.test.tsx
│       └── SetupWizard.test.tsx
└── lib/
    └── auth-api.ts                   # Auth API call wrappers
```

## User Flows

### Login Flow

```
/login → enter email + password → POST /auth/login
  ├─ Success (no 2FA): store token → redirect /dashboard
  ├─ Requires 2FA (401 + requires_2fa): redirect /login/2fa?temp=xxx
  └─ Error: show toast error message
```

### 2FA Flow

```
/login/2fa?temp=xxx → enter 6-digit OTP → POST /auth/2fa/verify
  ├─ Success: store token → redirect /dashboard
  └─ Error: shake animation + clear input
```

### Password Reset Flow

```
/forgot-password → enter email → POST /auth/forgot-password
  → redirect /forgot-password/sent

User clicks email link → /reset-password?token=xxx
  → validate token (GET /auth/verify-reset-token)
  → enter new password → POST /auth/reset-password
  → redirect /login (with success toast)
```

### Setup Wizard Flow

```
/setup → Step 1: Admin Account (username / email / password / confirm)
       → Step 2: Site Info (name / slug / domain / timezone / language)
       → Step 3: Review summary
       → POST /setup/install
       → redirect /setup/complete
```

InstallationGuard middleware blocks access if already installed.

## Component Design

### LoginForm.tsx

- Fields: email (Input), password (Input with show/hide toggle)
- Validation: email format, password >= 8 chars
- Submit: POST /auth/login
- Links: "Forgot password?" → /forgot-password
- Loading state: button spinner

### TwoFactorForm.tsx

- Input: shadcn InputOTP (6 separate digit boxes)
- Auto-submit when 6 digits entered
- Support paste from authenticator app
- Resend not applicable (TOTP, not email code)
- Back link → /login

### ForgotPasswordForm.tsx

- Fields: email (Input)
- Submit: POST /auth/forgot-password
- Always show success (prevent email enumeration)
- Back link → /login

### ResetPasswordForm.tsx

- Fields: new password, confirm password
- On mount: verify token via GET /auth/verify-reset-token
- Invalid/expired token → error page with "Request new link" button
- Submit: POST /auth/reset-password

### SetupWizard.tsx

- 3-step stepper with progress indicator (Step 1/2/3)
- Step navigation: Next/Back buttons
- Per-step client validation before proceeding
- Final step: review all entered data
- Submit: single POST /setup/install with all fields
- Advisory lock handled by backend

## Form Validation

Client-side: React Hook Form + Zod schemas

```
email:    z.string().email()
password: z.string().min(8)
slug:     z.string().regex(/^[a-z0-9-]{3,50}$/)
otp:      z.string().length(6).regex(/^\d{6}$/)
```

Server errors: caught by ApiError, displayed as Sonner toast.

## New Dependencies

- `input-otp` — underlying library for shadcn InputOTP component
- `react-hook-form` + `@hookform/resolvers` + `zod` — form management
- shadcn CLI: `input-otp` component (not yet installed)

## Auth Token Management

- Login success: backend sets `refresh_token` as httpOnly cookie
- Access token: stored in Zustand `useAuthStore` (memory only)
- Page redirect: `window.location.href = '/dashboard'`
- Auth state: `useAuthStore.setAuth(user, token)` on successful login

## i18n

Extend existing zh-CN.json and en.json with auth-specific keys:

```json
{
  "auth": {
    "loginTitle": "Sign in to your account",
    "emailPlaceholder": "you@example.com",
    "passwordPlaceholder": "Enter your password",
    "forgotPasswordLink": "Forgot password?",
    "twoFactorTitle": "Two-factor authentication",
    "twoFactorDescription": "Enter the 6-digit code from your authenticator app",
    "forgotPasswordTitle": "Reset your password",
    "forgotPasswordDescription": "Enter your email and we'll send you a reset link",
    "emailSentTitle": "Check your email",
    "emailSentDescription": "We've sent a password reset link to your email",
    "resetPasswordTitle": "Set new password",
    "setupTitle": "Welcome to Sky Flux CMS",
    "setupStep1": "Admin Account",
    "setupStep2": "Site Information",
    "setupStep3": "Review & Install",
    "setupComplete": "Installation Complete"
  }
}
```

## Middleware Updates

Current `PUBLIC_PATHS` already includes `/login`, `/setup`, `/forgot-password`, `/reset-password`.
Need to also allow `/setup/complete` and `/forgot-password/sent` — both match existing prefix rules.

## Testing Strategy (TDD)

Each React component: test file written FIRST, then implementation.

- LoginForm: submit success, submit error, 2FA redirect, validation, loading state (~12 tests)
- TwoFactorForm: 6-digit input, auto-submit, invalid code, back nav (~10 tests)
- ForgotPasswordForm: submit, always-success UX, validation (~8 tests)
- ResetPasswordForm: token validation, submit, password mismatch, expired token (~10 tests)
- SetupWizard: step navigation, per-step validation, final submit, review display (~15 tests)

Total estimate: ~55-65 tests

## Agent Teams Division

| Agent | Scope | Deliverables |
|-------|-------|-------------|
| Agent 1 (infra) | AuthLayout, auth-api, i18n keys, InputOTP install | Foundation |
| Agent 2 (login-2fa) | LoginForm + TwoFactorForm + pages + tests | Login + 2FA |
| Agent 3 (reset-setup) | ForgotPassword + ResetPassword + SetupWizard + pages + tests | Reset + Setup |

Dependencies: Agent 2 & 3 depend on Agent 1 completing AuthLayout + auth-api first.
