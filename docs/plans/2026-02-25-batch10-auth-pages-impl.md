# Batch 10: Auth Pages Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build all 7 auth pages (Login, 2FA, Forgot Password, Email Sent, Reset Password, Setup Wizard, Setup Complete) with React Islands on Astro SSR, using TDD.

**Architecture:** Each page is an Astro `.astro` file (SSR shell) embedding a React Island (`client:load`) for form interactivity. All pages share an `AuthLayout.astro` centered-card layout. Forms use React Hook Form + Zod for validation, `api-client.ts` for HTTP calls, and Sonner for error toasts.

**Tech Stack:** Astro 5 SSR, React 19, shadcn/ui (InputOTP, Card, Input, Button, Form), React Hook Form, Zod, Vitest, React Testing Library, Zustand (auth store), react-i18next.

---

## Pre-requisites & Conventions

**Existing infrastructure to reuse (do NOT recreate):**
- `web/src/lib/api-client.ts` — `api.post()`, `api.get()`, `ApiError` class
- `web/src/stores/auth-store.ts` — `useAuthStore` with `setAuth(user, token)`, `clearAuth()`
- `web/src/i18n/config.ts` + `web/src/i18n/locales/{en,zh-CN}.json`
- `web/src/components/layout/LocaleSwitcher.tsx` + `ThemeToggle.tsx`
- `web/src/components/ui/*` — shadcn components (button, card, input, label, form, separator, etc.)
- `web/src/lib/utils.ts` — `cn()` helper
- `web/src/layouts/BaseLayout.astro` — base HTML shell
- `web/src/middleware.ts` — Astro middleware, `PUBLIC_PATHS` already includes `/login`, `/setup`, `/forgot-password`, `/reset-password`
- `web/vitest.config.ts` — configured with jsdom, `@/` alias, `src/test/setup.ts`
- `web/src/test/setup.ts` — imports `@testing-library/jest-dom/vitest`

**Backend API response format:**
All endpoints return `{ success: boolean, data?: T, error?: string }`.
- Login success (no 2FA): `200 { success: true, data: { user: {id,email,display_name}, access_token, token_type, expires_in } }` + `Set-Cookie: refresh_token=...`
- Login 2FA challenge: `200 { success: true, data: { temp_token, token_type, expires_in, requires: "totp" } }`
- 2FA validate: `POST /api/v1/auth/2fa/validate` with `{ code }` + `Authorization: Bearer <temp_token>` → same as login success
- Forgot password: `200 { success: true, data: { message: "..." } }` — always 200
- Reset password: `POST /api/v1/auth/reset-password` with `{ token, new_password }` → `200 { success: true, data: { message: "..." } }`
- Setup check: `POST /api/v1/setup/check` → `200 { success: true, data: { installed: boolean } }`
- Setup install: `POST /api/v1/setup/initialize` with full body → `200 { success: true, data: { user, site, access_token, ... } }`

**Test conventions (follow existing patterns):**
```tsx
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach } from 'vitest';

// Mock api-client
vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class ApiError extends Error {
    status: number;
    constructor(status: number, message: string) {
      super(message);
      this.status = status;
      this.name = 'ApiError';
    }
  },
}));

// Mock i18n
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));
```

**Run commands (from `web/` directory):**
```bash
# Run all tests
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run

# Run single test file
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/LoginForm.test.tsx

# Type check
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run astro check

# Add shadcn component
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx shadcn@latest add input-otp
```

---

## Agent 1: Infrastructure (AuthLayout + auth-api + deps + i18n)

Agent 1 runs FIRST. Agents 2 and 3 depend on its output.

### Task 1.1: Install Dependencies

**Step 1: Install form libraries**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun add react-hook-form @hookform/resolvers zod
```

**Step 2: Install shadcn InputOTP component**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx shadcn@latest add input-otp
```

This installs `input-otp` peer dep + creates `src/components/ui/input-otp.tsx`.

**Step 3: Verify**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run
```

Expected: all 86 existing tests pass.

**Step 4: Commit**

```bash
git add web/package.json web/bun.lock web/src/components/ui/input-otp.tsx
git commit -m "feat(web): add react-hook-form, zod, and shadcn input-otp"
```

---

### Task 1.2: Create auth-api.ts

**Files:**
- Create: `web/src/lib/auth-api.ts`
- Test: `web/src/lib/__tests__/auth-api.test.ts`

**Step 1: Write the failing test**

Create `web/src/lib/__tests__/auth-api.test.ts`:

```ts
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { api } from '@/lib/api-client';

vi.mock('@/lib/api-client', () => ({
  api: {
    post: vi.fn(),
    get: vi.fn(),
  },
}));

import { authApi } from '@/lib/auth-api';

describe('authApi', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  describe('login', () => {
    it('calls POST /api/v1/auth/login with email and password', async () => {
      const mockResp = {
        success: true,
        data: { user: { id: '1', email: 'a@b.com', display_name: 'A' }, access_token: 'tok' },
      };
      vi.mocked(api.post).mockResolvedValue(mockResp);

      const result = await authApi.login('a@b.com', 'password123');
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/login', {
        email: 'a@b.com',
        password: 'password123',
      });
      expect(result).toEqual(mockResp);
    });
  });

  describe('validate2FA', () => {
    it('calls POST /api/v1/auth/2fa/validate with code and temp token header', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.validate2FA('123456', 'temp-tok');
      expect(api.post).toHaveBeenCalledWith(
        '/api/v1/auth/2fa/validate',
        { code: '123456' },
        { headers: { Authorization: 'Bearer temp-tok' } },
      );
    });
  });

  describe('forgotPassword', () => {
    it('calls POST /api/v1/auth/forgot-password with email', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.forgotPassword('a@b.com');
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/forgot-password', {
        email: 'a@b.com',
      });
    });
  });

  describe('resetPassword', () => {
    it('calls POST /api/v1/auth/reset-password with token and new password', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.resetPassword('reset-tok', 'newpass123');
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/reset-password', {
        token: 'reset-tok',
        new_password: 'newpass123',
      });
    });
  });

  describe('setupCheck', () => {
    it('calls POST /api/v1/setup/check', async () => {
      vi.mocked(api.post).mockResolvedValue({ success: true, data: { installed: false } });

      const result = await authApi.setupCheck();
      expect(api.post).toHaveBeenCalledWith('/api/v1/setup/check');
      expect(result).toEqual({ success: true, data: { installed: false } });
    });
  });

  describe('setupInstall', () => {
    it('calls POST /api/v1/setup/initialize with full payload', async () => {
      const payload = {
        site_name: 'My Site',
        site_slug: 'my-site',
        site_url: 'https://example.com',
        admin_email: 'a@b.com',
        admin_password: 'pass1234',
        admin_display_name: 'Admin',
        locale: 'en',
      };
      vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

      await authApi.setupInstall(payload);
      expect(api.post).toHaveBeenCalledWith('/api/v1/setup/initialize', payload);
    });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/lib/__tests__/auth-api.test.ts`

Expected: FAIL — `authApi` is not exported from `@/lib/auth-api`.

**Step 3: Write implementation**

Create `web/src/lib/auth-api.ts`:

```ts
import { api } from './api-client';

export interface LoginResponse {
  success: boolean;
  data: LoginSuccessData | Login2FAData;
}

export interface LoginSuccessData {
  user: { id: string; email: string; display_name: string };
  access_token: string;
  token_type: string;
  expires_in: number;
}

export interface Login2FAData {
  temp_token: string;
  token_type: string;
  expires_in: number;
  requires: 'totp';
}

export interface SetupInstallPayload {
  site_name: string;
  site_slug: string;
  site_url: string;
  admin_email: string;
  admin_password: string;
  admin_display_name: string;
  locale?: string;
}

export function isLogin2FA(data: LoginSuccessData | Login2FAData): data is Login2FAData {
  return 'requires' in data && data.requires === 'totp';
}

export const authApi = {
  login: (email: string, password: string) =>
    api.post<LoginResponse>('/api/v1/auth/login', { email, password }),

  validate2FA: (code: string, tempToken: string) =>
    api.post<LoginResponse>('/api/v1/auth/2fa/validate', { code }, {
      headers: { Authorization: `Bearer ${tempToken}` },
    }),

  forgotPassword: (email: string) =>
    api.post('/api/v1/auth/forgot-password', { email }),

  resetPassword: (token: string, newPassword: string) =>
    api.post('/api/v1/auth/reset-password', { token, new_password: newPassword }),

  setupCheck: () =>
    api.post<{ success: boolean; data: { installed: boolean } }>('/api/v1/setup/check'),

  setupInstall: (payload: SetupInstallPayload) =>
    api.post('/api/v1/setup/initialize', payload),
};
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/lib/__tests__/auth-api.test.ts`

Expected: 6 tests PASS.

**Step 5: Commit**

```bash
git add web/src/lib/auth-api.ts web/src/lib/__tests__/auth-api.test.ts
git commit -m "feat(web): add auth-api wrapper with typed endpoints"
```

---

### Task 1.3: Create AuthLayout.astro

**Files:**
- Create: `web/src/layouts/AuthLayout.astro`
- Modify: `web/src/pages/login.astro` (rewrite to use AuthLayout)

**Step 1: Create AuthLayout**

Create `web/src/layouts/AuthLayout.astro`:

```astro
---
import BaseLayout from './BaseLayout.astro';
import { Sparkles } from 'lucide-astro';

interface Props {
  title?: string;
  wide?: boolean;
}

const { title = 'Sky Flux CMS', wide = false } = Astro.props;
const cardWidth = wide ? 'max-w-lg' : 'max-w-md';
---

<BaseLayout title={title}>
  <div class="flex min-h-screen flex-col items-center justify-center px-4">
    <div class={`w-full ${cardWidth}`}>
      <div class="mb-8 flex items-center justify-center gap-2">
        <Sparkles class="size-6 text-primary" />
        <span class="text-xl font-bold">Sky Flux CMS</span>
      </div>

      <slot />

      <p class="mt-8 text-center text-xs text-muted-foreground">
        &copy; {new Date().getFullYear()} Sky Flux CMS
      </p>
    </div>
  </div>
</BaseLayout>
```

**Step 2: Verify astro check passes**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run astro check`

Expected: 0 errors.

**Step 3: Commit**

```bash
git add web/src/layouts/AuthLayout.astro
git commit -m "feat(web): add AuthLayout with centered card and logo"
```

---

### Task 1.4: Extend i18n Translations

**Files:**
- Modify: `web/src/i18n/locales/en.json`
- Modify: `web/src/i18n/locales/zh-CN.json`

**Step 1: Add auth page keys to en.json**

Add these keys to the existing `"auth"` section in `web/src/i18n/locales/en.json`:

```json
{
  "auth": {
    "login": "Login",
    "logout": "Logout",
    "email": "Email",
    "password": "Password",
    "forgotPassword": "Forgot Password",
    "resetPassword": "Reset Password",
    "twoFactor": "Two-Factor Authentication",
    "loginButton": "Sign In",
    "rememberMe": "Remember Me",
    "loginTitle": "Sign in to your account",
    "emailPlaceholder": "you@example.com",
    "passwordPlaceholder": "Enter your password",
    "forgotPasswordLink": "Forgot password?",
    "twoFactorTitle": "Two-factor authentication",
    "twoFactorDescription": "Enter the 6-digit code from your authenticator app",
    "twoFactorBackToLogin": "Back to login",
    "forgotPasswordTitle": "Reset your password",
    "forgotPasswordDescription": "Enter your email and we'll send you a reset link",
    "forgotPasswordSubmit": "Send reset link",
    "emailSentTitle": "Check your email",
    "emailSentDescription": "If an account exists with that email, we've sent a password reset link.",
    "emailSentBackToLogin": "Back to login",
    "resetPasswordTitle": "Set new password",
    "resetPasswordDescription": "Enter your new password below",
    "newPassword": "New password",
    "confirmPassword": "Confirm password",
    "resetPasswordSubmit": "Reset password",
    "resetPasswordSuccess": "Password reset successfully. Please sign in.",
    "resetTokenInvalid": "This reset link is invalid or has expired.",
    "resetTokenRequestNew": "Request a new link",
    "setupTitle": "Welcome to Sky Flux CMS",
    "setupDescription": "Let's set up your CMS in a few steps",
    "setupStep1": "Admin Account",
    "setupStep2": "Site Information",
    "setupStep3": "Review & Install",
    "setupAdminUsername": "Display name",
    "setupAdminUsernamePlaceholder": "Admin",
    "setupAdminEmail": "Admin email",
    "setupAdminEmailPlaceholder": "admin@example.com",
    "setupSiteName": "Site name",
    "setupSiteNamePlaceholder": "My Awesome Site",
    "setupSiteSlug": "Site slug",
    "setupSiteSlugPlaceholder": "my-site",
    "setupSiteUrl": "Site URL",
    "setupSiteUrlPlaceholder": "https://example.com",
    "setupLocale": "Language",
    "setupNext": "Next",
    "setupBack": "Back",
    "setupInstall": "Install",
    "setupInstalling": "Installing...",
    "setupComplete": "Installation Complete",
    "setupCompleteDescription": "Your CMS is ready to use.",
    "setupGoToLogin": "Go to login",
    "passwordMinLength": "Password must be at least 8 characters",
    "passwordsDoNotMatch": "Passwords do not match",
    "slugFormat": "Only lowercase letters, numbers, and hyphens (3-50 chars)",
    "showPassword": "Show password",
    "hidePassword": "Hide password"
  }
}
```

**Step 2: Add auth page keys to zh-CN.json**

Add the same keys in Chinese to `web/src/i18n/locales/zh-CN.json`:

```json
{
  "auth": {
    "login": "登录",
    "logout": "退出登录",
    "email": "邮箱",
    "password": "密码",
    "forgotPassword": "忘记密码",
    "resetPassword": "重置密码",
    "twoFactor": "两步验证",
    "loginButton": "登录",
    "rememberMe": "记住我",
    "loginTitle": "登录到您的账户",
    "emailPlaceholder": "you@example.com",
    "passwordPlaceholder": "输入密码",
    "forgotPasswordLink": "忘记密码？",
    "twoFactorTitle": "两步验证",
    "twoFactorDescription": "请输入身份验证器应用中的 6 位验证码",
    "twoFactorBackToLogin": "返回登录",
    "forgotPasswordTitle": "重置密码",
    "forgotPasswordDescription": "输入您的邮箱，我们将发送重置链接",
    "forgotPasswordSubmit": "发送重置链接",
    "emailSentTitle": "请查收邮件",
    "emailSentDescription": "如果该邮箱对应的账户存在，我们已发送密码重置链接。",
    "emailSentBackToLogin": "返回登录",
    "resetPasswordTitle": "设置新密码",
    "resetPasswordDescription": "请在下方输入新密码",
    "newPassword": "新密码",
    "confirmPassword": "确认密码",
    "resetPasswordSubmit": "重置密码",
    "resetPasswordSuccess": "密码重置成功，请重新登录。",
    "resetTokenInvalid": "此重置链接无效或已过期。",
    "resetTokenRequestNew": "重新获取链接",
    "setupTitle": "欢迎使用 Sky Flux CMS",
    "setupDescription": "让我们通过几个步骤完成初始设置",
    "setupStep1": "管理员账户",
    "setupStep2": "站点信息",
    "setupStep3": "确认并安装",
    "setupAdminUsername": "显示名称",
    "setupAdminUsernamePlaceholder": "管理员",
    "setupAdminEmail": "管理员邮箱",
    "setupAdminEmailPlaceholder": "admin@example.com",
    "setupSiteName": "站点名称",
    "setupSiteNamePlaceholder": "我的站点",
    "setupSiteSlug": "站点标识",
    "setupSiteSlugPlaceholder": "my-site",
    "setupSiteUrl": "站点地址",
    "setupSiteUrlPlaceholder": "https://example.com",
    "setupLocale": "语言",
    "setupNext": "下一步",
    "setupBack": "上一步",
    "setupInstall": "安装",
    "setupInstalling": "正在安装...",
    "setupComplete": "安装完成",
    "setupCompleteDescription": "您的 CMS 已准备就绪。",
    "setupGoToLogin": "前往登录",
    "passwordMinLength": "密码至少 8 个字符",
    "passwordsDoNotMatch": "两次输入的密码不一致",
    "slugFormat": "仅限小写字母、数字和连字符（3-50 个字符）",
    "showPassword": "显示密码",
    "hidePassword": "隐藏密码"
  }
}
```

**Step 2: Run tests**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run`

Expected: all tests pass (i18n changes are data-only, no code breaks).

**Step 3: Commit**

```bash
git add web/src/i18n/locales/en.json web/src/i18n/locales/zh-CN.json
git commit -m "feat(web): add auth page i18n keys for login, 2FA, reset, and setup"
```

---

## Agent 2: Login + 2FA Pages

Depends on Agent 1 completing Tasks 1.1–1.4.

### Task 2.1: LoginForm Component (TDD)

**Files:**
- Test: `web/src/components/auth/__tests__/LoginForm.test.tsx`
- Create: `web/src/components/auth/LoginForm.tsx`

**Step 1: Write the failing tests**

Create `web/src/components/auth/__tests__/LoginForm.test.tsx`:

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { LoginForm } from '../LoginForm';

describe('LoginForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset window.location
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders email and password fields', () => {
    render(<LoginForm />);
    expect(screen.getByLabelText(/auth\.email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.password/i)).toBeInTheDocument();
  });

  it('renders sign in button', () => {
    render(<LoginForm />);
    expect(screen.getByRole('button', { name: /auth\.loginButton/i })).toBeInTheDocument();
  });

  it('renders forgot password link', () => {
    render(<LoginForm />);
    expect(screen.getByText(/auth\.forgotPasswordLink/i)).toBeInTheDocument();
  });

  it('shows validation error for empty email', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('shows validation error for short password', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.type(screen.getByLabelText(/auth\.email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'short');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls login API and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        user: { id: '1', email: 'a@b.com', display_name: 'Admin' },
        access_token: 'jwt-tok',
        token_type: 'Bearer',
        expires_in: 900,
      },
    });

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/login', {
        email: 'a@b.com',
        password: 'password123',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/dashboard');
    });
  });

  it('redirects to 2FA page when requires_2fa', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        temp_token: 'tmp-tok',
        token_type: 'Bearer',
        expires_in: 300,
        requires: 'totp',
      },
    });

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(window.location.href).toContain('/login/2fa');
      expect(window.location.href).toContain('temp=tmp-tok');
    });
  });

  it('shows toast on API error', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(401, 'Invalid credentials'));

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Invalid credentials');
    });
  });

  it('disables button while submitting', async () => {
    const user = userEvent.setup();
    let resolveLogin: (v: unknown) => void;
    vi.mocked(api.post).mockImplementation(
      () => new Promise((resolve) => { resolveLogin = resolve; }),
    );

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    expect(screen.getByRole('button', { name: /auth\.loginButton/i })).toBeDisabled();

    // Resolve to clean up
    resolveLogin!({ success: true, data: { user: {}, access_token: '' } });
  });

  it('toggles password visibility', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(/auth\.password/i);
    expect(passwordInput).toHaveAttribute('type', 'password');

    const toggleBtn = screen.getByRole('button', { name: /auth\.showPassword/i });
    await user.click(toggleBtn);
    expect(passwordInput).toHaveAttribute('type', 'text');
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/LoginForm.test.tsx`

Expected: FAIL — `LoginForm` not found.

**Step 3: Write implementation**

Create `web/src/components/auth/LoginForm.tsx`:

```tsx
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { api, ApiError } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import type { LoginSuccessData, Login2FAData } from '@/lib/auth-api';
import { isLogin2FA } from '@/lib/auth-api';

const loginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(8),
});

type LoginFormValues = z.infer<typeof loginSchema>;

export function LoginForm() {
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);
  const setAuth = useAuthStore((s) => s.setAuth);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<LoginFormValues>({
    resolver: zodResolver(loginSchema),
  });

  const onSubmit = async (values: LoginFormValues) => {
    try {
      const resp = await api.post<{
        success: boolean;
        data: LoginSuccessData | Login2FAData;
      }>('/api/v1/auth/login', {
        email: values.email,
        password: values.password,
      });

      if (isLogin2FA(resp.data)) {
        window.location.href = `/login/2fa?temp=${resp.data.temp_token}`;
        return;
      }

      const data = resp.data as LoginSuccessData;
      setAuth(
        {
          id: data.user.id,
          email: data.user.email,
          displayName: data.user.display_name,
          avatarUrl: '',
        },
        data.access_token,
      );
      window.location.href = '/dashboard';
    } catch (err) {
      if (err instanceof ApiError) {
        toast.error(err.message);
      } else {
        toast.error(t('errors.networkError'));
      }
    }
  };

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.loginTitle')}</CardTitle>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">{t('auth.email')}</Label>
            <Input
              id="email"
              type="email"
              placeholder={t('auth.emailPlaceholder')}
              {...register('email')}
              aria-invalid={!!errors.email}
            />
            {errors.email && (
              <p className="text-sm text-destructive">{errors.email.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="password">{t('auth.password')}</Label>
            <div className="relative">
              <Input
                id="password"
                type={showPassword ? 'text' : 'password'}
                placeholder={t('auth.passwordPlaceholder')}
                {...register('password')}
                aria-invalid={!!errors.password}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="absolute right-2 top-1/2 -translate-y-1/2"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? t('auth.hidePassword') : t('auth.showPassword')}
              >
                {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </Button>
            </div>
            {errors.password && (
              <p className="text-sm text-destructive">{t('auth.passwordMinLength')}</p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="mr-2 size-4 animate-spin" />}
            {t('auth.loginButton')}
          </Button>

          <div className="text-center">
            <a
              href="/forgot-password"
              className="text-sm text-muted-foreground hover:text-primary"
            >
              {t('auth.forgotPasswordLink')}
            </a>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/LoginForm.test.tsx`

Expected: 10 tests PASS.

**Step 5: Commit**

```bash
git add web/src/components/auth/LoginForm.tsx web/src/components/auth/__tests__/LoginForm.test.tsx
git commit -m "feat(web): add LoginForm component with validation and 2FA redirect"
```

---

### Task 2.2: TwoFactorForm Component (TDD)

**Files:**
- Test: `web/src/components/auth/__tests__/TwoFactorForm.test.tsx`
- Create: `web/src/components/auth/TwoFactorForm.tsx`

**Step 1: Write the failing tests**

Create `web/src/components/auth/__tests__/TwoFactorForm.test.tsx`:

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { TwoFactorForm } from '../TwoFactorForm';

describe('TwoFactorForm', () => {
  const defaultProps = { tempToken: 'test-temp-token' };

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders title and description', () => {
    render(<TwoFactorForm {...defaultProps} />);
    expect(screen.getByText(/auth\.twoFactorTitle/i)).toBeInTheDocument();
    expect(screen.getByText(/auth\.twoFactorDescription/i)).toBeInTheDocument();
  });

  it('renders 6 OTP input slots', () => {
    render(<TwoFactorForm {...defaultProps} />);
    // InputOTP renders individual slots
    const inputs = screen.getAllByRole('textbox');
    expect(inputs.length).toBeGreaterThanOrEqual(1);
  });

  it('renders back to login link', () => {
    render(<TwoFactorForm {...defaultProps} />);
    expect(screen.getByText(/auth\.twoFactorBackToLogin/i)).toBeInTheDocument();
  });

  it('calls validate2FA API when 6 digits entered and redirects on success', async () => {
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        user: { id: '1', email: 'a@b.com', display_name: 'Admin' },
        access_token: 'jwt-tok',
        token_type: 'Bearer',
        expires_in: 900,
      },
    });

    render(<TwoFactorForm {...defaultProps} />);

    // Type 6 digits — InputOTP accepts keyboard input
    const otpInput = screen.getByRole('textbox');
    await userEvent.click(otpInput);
    await userEvent.keyboard('123456');

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        '/api/v1/auth/2fa/validate',
        { code: '123456' },
        { headers: { Authorization: 'Bearer test-temp-token' } },
      );
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/dashboard');
    });
  });

  it('shows toast on invalid code', async () => {
    vi.mocked(api.post).mockRejectedValue(new ApiError(401, 'Invalid TOTP code'));

    render(<TwoFactorForm {...defaultProps} />);
    const otpInput = screen.getByRole('textbox');
    await userEvent.click(otpInput);
    await userEvent.keyboard('000000');

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Invalid TOTP code');
    });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/TwoFactorForm.test.tsx`

Expected: FAIL — `TwoFactorForm` not found.

**Step 3: Write implementation**

Create `web/src/components/auth/TwoFactorForm.tsx`:

```tsx
import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import {
  InputOTP,
  InputOTPGroup,
  InputOTPSlot,
} from '@/components/ui/input-otp';
import { api, ApiError } from '@/lib/api-client';
import { useAuthStore } from '@/stores/auth-store';
import type { LoginSuccessData } from '@/lib/auth-api';

interface TwoFactorFormProps {
  tempToken: string;
}

export function TwoFactorForm({ tempToken }: TwoFactorFormProps) {
  const { t } = useTranslation();
  const [isSubmitting, setIsSubmitting] = useState(false);
  const setAuth = useAuthStore((s) => s.setAuth);

  const handleComplete = async (code: string) => {
    if (code.length !== 6) return;
    setIsSubmitting(true);

    try {
      const resp = await api.post<{
        success: boolean;
        data: LoginSuccessData;
      }>('/api/v1/auth/2fa/validate', { code }, {
        headers: { Authorization: `Bearer ${tempToken}` },
      });

      const data = resp.data;
      setAuth(
        {
          id: data.user.id,
          email: data.user.email,
          displayName: data.user.display_name,
          avatarUrl: '',
        },
        data.access_token,
      );
      window.location.href = '/dashboard';
    } catch (err) {
      setIsSubmitting(false);
      if (err instanceof ApiError) {
        toast.error(err.message);
      } else {
        toast.error(t('errors.networkError'));
      }
    }
  };

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.twoFactorTitle')}</CardTitle>
        <CardDescription>{t('auth.twoFactorDescription')}</CardDescription>
      </CardHeader>
      <CardContent className="flex flex-col items-center gap-6">
        {isSubmitting ? (
          <Loader2 className="size-8 animate-spin text-primary" />
        ) : (
          <InputOTP maxLength={6} onComplete={handleComplete}>
            <InputOTPGroup>
              <InputOTPSlot index={0} />
              <InputOTPSlot index={1} />
              <InputOTPSlot index={2} />
              <InputOTPSlot index={3} />
              <InputOTPSlot index={4} />
              <InputOTPSlot index={5} />
            </InputOTPGroup>
          </InputOTP>
        )}

        <a
          href="/login"
          className="text-sm text-muted-foreground hover:text-primary"
        >
          {t('auth.twoFactorBackToLogin')}
        </a>
      </CardContent>
    </Card>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/TwoFactorForm.test.tsx`

Expected: 5 tests PASS.

**Step 5: Commit**

```bash
git add web/src/components/auth/TwoFactorForm.tsx web/src/components/auth/__tests__/TwoFactorForm.test.tsx
git commit -m "feat(web): add TwoFactorForm with OTP input and auto-submit"
```

---

### Task 2.3: Login and 2FA Astro Pages

**Files:**
- Modify: `web/src/pages/login.astro` (rewrite)
- Create: `web/src/pages/login/2fa.astro`

**Step 1: Rewrite login.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { LoginForm } from '@/components/auth/LoginForm';
---

<AuthLayout title="Login - Sky Flux CMS">
  <LoginForm client:load />
</AuthLayout>
```

**Step 2: Create 2fa.astro**

Create `web/src/pages/login/2fa.astro`:

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { TwoFactorForm } from '@/components/auth/TwoFactorForm';

const tempToken = Astro.url.searchParams.get('temp') || '';
if (!tempToken) {
  return Astro.redirect('/login');
}
---

<AuthLayout title="Two-Factor Authentication - Sky Flux CMS">
  <TwoFactorForm client:load tempToken={tempToken} />
</AuthLayout>
```

**Step 3: Verify**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run astro check`

Expected: 0 errors.

**Step 4: Commit**

```bash
git add web/src/pages/login.astro web/src/pages/login/2fa.astro
git commit -m "feat(web): add login and 2FA Astro pages with React islands"
```

---

## Agent 3: Reset Password + Setup Wizard Pages

Depends on Agent 1 completing Tasks 1.1–1.4.

### Task 3.1: ForgotPasswordForm Component (TDD)

**Files:**
- Test: `web/src/components/auth/__tests__/ForgotPasswordForm.test.tsx`
- Create: `web/src/components/auth/ForgotPasswordForm.tsx`

**Step 1: Write the failing tests**

Create `web/src/components/auth/__tests__/ForgotPasswordForm.test.tsx`:

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { api } from '@/lib/api-client';
import { ForgotPasswordForm } from '../ForgotPasswordForm';

describe('ForgotPasswordForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders email field and submit button', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByLabelText(/auth\.email/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i })).toBeInTheDocument();
  });

  it('renders title and description', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByText(/auth\.forgotPasswordTitle/i)).toBeInTheDocument();
    expect(screen.getByText(/auth\.forgotPasswordDescription/i)).toBeInTheDocument();
  });

  it('renders back to login link', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByText(/auth\.twoFactorBackToLogin/i)).toBeInTheDocument();
  });

  it('does not submit with empty email', async () => {
    const user = userEvent.setup();
    render(<ForgotPasswordForm />);

    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls forgot password API and redirects to sent page', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

    render(<ForgotPasswordForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/forgot-password', {
        email: 'a@b.com',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/forgot-password/sent');
    });
  });

  it('redirects to sent page even on API error (prevent email enumeration)', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new Error('network'));

    render(<ForgotPasswordForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    // Should still redirect — we never reveal if email exists
    await waitFor(() => {
      expect(window.location.href).toBe('/forgot-password/sent');
    });
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/ForgotPasswordForm.test.tsx`

Expected: FAIL.

**Step 3: Write implementation**

Create `web/src/components/auth/ForgotPasswordForm.tsx`:

```tsx
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { Loader2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { api } from '@/lib/api-client';

const schema = z.object({
  email: z.string().email(),
});

type FormValues = z.infer<typeof schema>;

export function ForgotPasswordForm() {
  const { t } = useTranslation();

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (values: FormValues) => {
    try {
      await api.post('/api/v1/auth/forgot-password', { email: values.email });
    } catch {
      // Intentionally swallow — never reveal if email exists
    }
    window.location.href = '/forgot-password/sent';
  };

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.forgotPasswordTitle')}</CardTitle>
        <CardDescription>{t('auth.forgotPasswordDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="email">{t('auth.email')}</Label>
            <Input
              id="email"
              type="email"
              placeholder={t('auth.emailPlaceholder')}
              {...register('email')}
              aria-invalid={!!errors.email}
            />
            {errors.email && (
              <p className="text-sm text-destructive">{errors.email.message}</p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="mr-2 size-4 animate-spin" />}
            {t('auth.forgotPasswordSubmit')}
          </Button>

          <div className="text-center">
            <a href="/login" className="text-sm text-muted-foreground hover:text-primary">
              {t('auth.twoFactorBackToLogin')}
            </a>
          </div>
        </form>
      </CardContent>
    </Card>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run src/components/auth/__tests__/ForgotPasswordForm.test.tsx`

Expected: 6 tests PASS.

**Step 5: Commit**

```bash
git add web/src/components/auth/ForgotPasswordForm.tsx web/src/components/auth/__tests__/ForgotPasswordForm.test.tsx
git commit -m "feat(web): add ForgotPasswordForm with anti-enumeration redirect"
```

---

### Task 3.2: ResetPasswordForm Component (TDD)

**Files:**
- Test: `web/src/components/auth/__tests__/ResetPasswordForm.test.tsx`
- Create: `web/src/components/auth/ResetPasswordForm.tsx`

**Step 1: Write the failing tests**

Create `web/src/components/auth/__tests__/ResetPasswordForm.test.tsx`:

```tsx
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { ResetPasswordForm } from '../ResetPasswordForm';

describe('ResetPasswordForm', () => {
  const defaultProps = { token: 'valid-reset-token' };

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders new password and confirm password fields', () => {
    render(<ResetPasswordForm {...defaultProps} />);
    expect(screen.getByLabelText(/auth\.newPassword/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.confirmPassword/i)).toBeInTheDocument();
  });

  it('renders title', () => {
    render(<ResetPasswordForm {...defaultProps} />);
    expect(screen.getByText(/auth\.resetPasswordTitle/i)).toBeInTheDocument();
  });

  it('shows error when passwords do not match', async () => {
    const user = userEvent.setup();
    render(<ResetPasswordForm {...defaultProps} />);

    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'different123');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('shows error for short password', async () => {
    const user = userEvent.setup();
    render(<ResetPasswordForm {...defaultProps} />);

    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'short');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'short');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls reset password API and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

    render(<ResetPasswordForm {...defaultProps} />);
    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'newpass1234');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'newpass1234');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/api/v1/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'newpass1234',
      });
    });

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('auth.resetPasswordSuccess');
    });
  });

  it('shows toast on API error', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(400, 'Token expired'));

    render(<ResetPasswordForm {...defaultProps} />);
    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'newpass1234');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'newpass1234');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Token expired');
    });
  });
});
```

**Step 2: Run test to verify it fails**

Expected: FAIL.

**Step 3: Write implementation**

Create `web/src/components/auth/ResetPasswordForm.tsx`:

```tsx
import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { Eye, EyeOff, Loader2 } from 'lucide-react';
import { toast } from 'sonner';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/ui/card';
import { api, ApiError } from '@/lib/api-client';

const schema = z.object({
  newPassword: z.string().min(8),
  confirmPassword: z.string().min(8),
}).refine((data) => data.newPassword === data.confirmPassword, {
  path: ['confirmPassword'],
  message: 'Passwords do not match',
});

type FormValues = z.infer<typeof schema>;

interface ResetPasswordFormProps {
  token: string;
}

export function ResetPasswordForm({ token }: ResetPasswordFormProps) {
  const { t } = useTranslation();
  const [showPassword, setShowPassword] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<FormValues>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (values: FormValues) => {
    try {
      await api.post('/api/v1/auth/reset-password', {
        token,
        new_password: values.newPassword,
      });
      toast.success(t('auth.resetPasswordSuccess'));
      window.location.href = '/login';
    } catch (err) {
      if (err instanceof ApiError) {
        toast.error(err.message);
      } else {
        toast.error(t('errors.networkError'));
      }
    }
  };

  return (
    <Card>
      <CardHeader className="text-center">
        <CardTitle>{t('auth.resetPasswordTitle')}</CardTitle>
        <CardDescription>{t('auth.resetPasswordDescription')}</CardDescription>
      </CardHeader>
      <CardContent>
        <form onSubmit={handleSubmit(onSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="newPassword">{t('auth.newPassword')}</Label>
            <div className="relative">
              <Input
                id="newPassword"
                type={showPassword ? 'text' : 'password'}
                {...register('newPassword')}
                aria-invalid={!!errors.newPassword}
              />
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="absolute right-2 top-1/2 -translate-y-1/2"
                onClick={() => setShowPassword(!showPassword)}
                aria-label={showPassword ? t('auth.hidePassword') : t('auth.showPassword')}
              >
                {showPassword ? <EyeOff className="size-4" /> : <Eye className="size-4" />}
              </Button>
            </div>
            {errors.newPassword && (
              <p className="text-sm text-destructive">{t('auth.passwordMinLength')}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="confirmPassword">{t('auth.confirmPassword')}</Label>
            <Input
              id="confirmPassword"
              type={showPassword ? 'text' : 'password'}
              {...register('confirmPassword')}
              aria-invalid={!!errors.confirmPassword}
            />
            {errors.confirmPassword && (
              <p className="text-sm text-destructive">{t('auth.passwordsDoNotMatch')}</p>
            )}
          </div>

          <Button type="submit" className="w-full" disabled={isSubmitting}>
            {isSubmitting && <Loader2 className="mr-2 size-4 animate-spin" />}
            {t('auth.resetPasswordSubmit')}
          </Button>
        </form>
      </CardContent>
    </Card>
  );
}
```

**Step 4: Run test to verify it passes**

Expected: 6 tests PASS.

**Step 5: Commit**

```bash
git add web/src/components/auth/ResetPasswordForm.tsx web/src/components/auth/__tests__/ResetPasswordForm.test.tsx
git commit -m "feat(web): add ResetPasswordForm with password confirmation validation"
```

---

### Task 3.3: SetupWizard Component (TDD)

**Files:**
- Test: `web/src/components/auth/__tests__/SetupWizard.test.tsx`
- Create: `web/src/components/auth/SetupWizard.tsx`

**Step 1: Write the failing tests**

Create `web/src/components/auth/__tests__/SetupWizard.test.tsx`:

```tsx
import { render, screen, waitFor, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
}));

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { SetupWizard } from '../SetupWizard';

describe('SetupWizard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders step 1 (admin account) by default', () => {
    render(<SetupWizard />);
    expect(screen.getByText(/auth\.setupStep1/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.setupAdminEmail/i)).toBeInTheDocument();
  });

  it('renders title and description', () => {
    render(<SetupWizard />);
    expect(screen.getByText(/auth\.setupTitle/i)).toBeInTheDocument();
  });

  it('validates step 1 before allowing next', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Click next without filling fields
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Should still be on step 1
    expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
  });

  it('navigates to step 2 after valid step 1', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    await waitFor(() => {
      expect(screen.getByText(/auth\.setupStep2/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });
  });

  it('can go back from step 2 to step 1', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Fill step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });

    // Go back
    await user.click(screen.getByRole('button', { name: /auth\.setupBack/i }));

    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
    });
  });

  it('navigates to step 3 (review) after valid step 2', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 — review
    await waitFor(() => {
      expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument();
      expect(screen.getByText('Admin')).toBeInTheDocument();
      expect(screen.getByText('admin@test.com')).toBeInTheDocument();
      expect(screen.getByText('My Site')).toBeInTheDocument();
    });
  });

  it('calls setup API on install and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: { user: {}, site: {}, access_token: 'tok' },
    });

    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 — install
    await waitFor(() => expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /auth\.setupInstall/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/api/v1/setup/initialize', {
        admin_display_name: 'Admin',
        admin_email: 'admin@test.com',
        admin_password: 'password123',
        site_name: 'My Site',
        site_slug: 'my-site',
        site_url: 'https://example.com',
        locale: 'en',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/setup/complete');
    });
  });

  it('shows toast on API error during install', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(500, 'Install failed'));

    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 — install
    await waitFor(() => expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /auth\.setupInstall/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Install failed');
    });
  });

  it('validates slug format on step 2', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2 — invalid slug
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'INVALID SLUG!');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Should stay on step 2
    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteSlug/i)).toBeInTheDocument();
    });
    expect(screen.queryByText(/auth\.setupStep3/i)).not.toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

Expected: FAIL.

**Step 3: Write implementation**

Create `web/src/components/auth/SetupWizard.tsx`. This is a larger component — the agent should implement:
- `useState` for current step (1, 2, 3)
- `useForm` with Zod for each step's validation
- Step 1: admin fields (display_name, email, password, confirm_password)
- Step 2: site fields (site_name, site_slug, site_url, locale select)
- Step 3: review summary + install button
- Navigation: Next/Back buttons, step indicator
- Submit: POST to `/api/v1/setup/initialize`

Key patterns:
```tsx
// Form data stored in ref to persist across steps
const formData = useRef({
  admin_display_name: '',
  admin_email: '',
  admin_password: '',
  site_name: '',
  site_slug: '',
  site_url: '',
  locale: 'en',
});

// Step 1 schema
const step1Schema = z.object({
  admin_display_name: z.string().min(1).max(100),
  admin_email: z.string().email(),
  password: z.string().min(8),
  confirmPassword: z.string().min(8),
}).refine((d) => d.password === d.confirmPassword, {
  path: ['confirmPassword'],
});

// Step 2 schema
const step2Schema = z.object({
  site_name: z.string().min(1).max(200),
  site_slug: z.string().regex(/^[a-z0-9-]{3,50}$/),
  site_url: z.string().url(),
  locale: z.string().optional(),
});
```

Use a `Separator` between step indicator and form. Step indicator shows "Step 1 of 3" text + progress dots/segments.

**Step 4: Run test to verify it passes**

Expected: 9 tests PASS.

**Step 5: Commit**

```bash
git add web/src/components/auth/SetupWizard.tsx web/src/components/auth/__tests__/SetupWizard.test.tsx
git commit -m "feat(web): add SetupWizard with 3-step form and validation"
```

---

### Task 3.4: Astro Pages for Reset + Setup

**Files:**
- Create: `web/src/pages/forgot-password/index.astro`
- Create: `web/src/pages/forgot-password/sent.astro`
- Create: `web/src/pages/reset-password.astro`
- Create: `web/src/pages/setup/index.astro`
- Create: `web/src/pages/setup/complete.astro`

**Step 1: Create forgot-password/index.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { ForgotPasswordForm } from '@/components/auth/ForgotPasswordForm';
---

<AuthLayout title="Forgot Password - Sky Flux CMS">
  <ForgotPasswordForm client:load />
</AuthLayout>
```

**Step 2: Create forgot-password/sent.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { MailCheck } from 'lucide-astro';
---

<AuthLayout title="Check Your Email - Sky Flux CMS">
  <div class="rounded-lg border bg-card p-6 text-center shadow-sm">
    <MailCheck class="mx-auto mb-4 size-12 text-primary" />
    <h2 class="mb-2 text-xl font-semibold">Check your email</h2>
    <p class="mb-6 text-sm text-muted-foreground">
      If an account exists with that email, we've sent a password reset link.
    </p>
    <a
      href="/login"
      class="text-sm font-medium text-primary hover:underline"
    >
      Back to login
    </a>
  </div>
</AuthLayout>
```

**Step 3: Create reset-password.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { ResetPasswordForm } from '@/components/auth/ResetPasswordForm';

const token = Astro.url.searchParams.get('token') || '';
if (!token) {
  return Astro.redirect('/forgot-password');
}
---

<AuthLayout title="Reset Password - Sky Flux CMS">
  <ResetPasswordForm client:load token={token} />
</AuthLayout>
```

**Step 4: Create setup/index.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { SetupWizard } from '@/components/auth/SetupWizard';
---

<AuthLayout title="Setup - Sky Flux CMS" wide>
  <SetupWizard client:load />
</AuthLayout>
```

**Step 5: Create setup/complete.astro**

```astro
---
import AuthLayout from '@/layouts/AuthLayout.astro';
import { PartyPopper } from 'lucide-astro';
---

<AuthLayout title="Setup Complete - Sky Flux CMS">
  <div class="rounded-lg border bg-card p-6 text-center shadow-sm">
    <PartyPopper class="mx-auto mb-4 size-12 text-primary" />
    <h2 class="mb-2 text-xl font-semibold">Installation Complete</h2>
    <p class="mb-6 text-sm text-muted-foreground">
      Your CMS is ready to use.
    </p>
    <a
      href="/login"
      class="inline-flex h-9 items-center justify-center rounded-md bg-primary px-4 text-sm font-medium text-primary-foreground hover:bg-primary/90"
    >
      Go to login
    </a>
  </div>
</AuthLayout>
```

**Step 6: Verify**

Run: `cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run astro check`

Expected: 0 errors.

**Step 7: Commit**

```bash
git add web/src/pages/forgot-password/ web/src/pages/reset-password.astro web/src/pages/setup/
git commit -m "feat(web): add forgot-password, reset-password, and setup Astro pages"
```

---

## Final Integration

### Task 4.1: Run All Tests and Verify

**Step 1: Run full test suite**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run vitest run
```

Expected: ~145+ tests pass (86 existing + ~60 new).

**Step 2: Run astro check**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun run astro check
```

Expected: 0 errors.

**Step 3: Run Biome lint**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx @biomejs/biome check src/
```

Fix any lint issues.

**Step 4: Final commit if any fixes needed**

```bash
git add -A
git commit -m "fix(web): lint and type fixes for batch 10 auth pages"
```
