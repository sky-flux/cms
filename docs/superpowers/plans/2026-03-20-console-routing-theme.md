# Console Routing + Theme Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Establish the console SPA's complete routing architecture, Sage Teal theme, layout components (Sidebar + Header), auth guard with token refresh, and wire all existing feature components into TanStack Router file routes.

**Architecture:** TanStack Router file-based routing with `_auth` and `_dashboard` layout groups. Auth guard in `_dashboard.tsx` beforeLoad calls `/api/v1/auth/me`. Sage Teal oklch theme with Geist Sans font and Phosphor duotone icons.

**Tech Stack:** React 19, TanStack Router v1, TanStack Query v5, shadcn/ui, Tailwind V4, Phosphor Icons, Geist Sans, Vitest

---

## Task 1: Install dependencies + theme tokens

**Goal:** Replace the tropical island theme with Sage Teal oklch design system. Swap Lucide for Phosphor Icons. Add Geist Sans font.

**Estimated time:** 3 minutes

### Steps

- [ ] **1.1** Install new dependencies and remove lucide-react:
  ```bash
  cd /Users/martinadamsdev/workspace/sky-flux-cms/console
  bun add geist @phosphor-icons/react
  bun remove lucide-react
  ```

- [ ] **1.2** Rewrite `console/src/styles.css` — replace the entire file with Sage Teal oklch variables. Remove all tropical island custom CSS (`.island-shell`, `.feature-card`, `.nav-link`, `.display-title`, `.island-kicker`, `.page-wrap`, `.rise-in`, `.site-footer`, body gradient/noise, Fraunces/Manrope font imports). Keep the `@import 'tailwindcss'`, `@plugin`, `tw-animate-css`, `@custom-variant dark`, `@theme inline` block, and `@layer base` block.

  New `:root` variables (from spec Section 8):
  ```css
  @import url('https://fonts.googleapis.com/css2?family=Noto+Sans+SC:wght@400;500;600;700&display=swap');
  @import 'tailwindcss';
  @plugin '@tailwindcss/typography';
  @import 'tw-animate-css';

  @custom-variant dark (&:is(.dark *));

  :root {
    --background: oklch(0.985 0.003 90);
    --foreground: oklch(0.15 0.01 60);
    --primary: oklch(0.55 0.12 175);
    --primary-hover: oklch(0.48 0.13 175);
    --primary-foreground: oklch(0.98 0.005 175);
    --primary-light: oklch(0.95 0.03 175);
    --card: oklch(1 0 0);
    --card-foreground: oklch(0.15 0.01 60);
    --popover: oklch(1 0 0);
    --popover-foreground: oklch(0.15 0.01 60);
    --secondary: oklch(0.955 0.005 90);
    --secondary-foreground: oklch(0.15 0.01 60);
    --muted: oklch(0.955 0.005 90);
    --muted-foreground: oklch(0.45 0.01 60);
    --accent: oklch(0.955 0.005 90);
    --accent-foreground: oklch(0.15 0.01 60);
    --destructive: oklch(0.55 0.2 25);
    --destructive-foreground: oklch(0.98 0.005 175);
    --border: oklch(0.90 0.005 90);
    --input: oklch(0.90 0.005 90);
    --ring: oklch(0.55 0.12 175 / 0.3);
    --sidebar: oklch(0.975 0.005 90);
    --sidebar-foreground: oklch(0.15 0.01 60);
    --sidebar-primary: oklch(0.55 0.12 175);
    --sidebar-primary-foreground: oklch(0.98 0.005 175);
    --sidebar-accent: oklch(0.955 0.005 90);
    --sidebar-accent-foreground: oklch(0.15 0.01 60);
    --sidebar-border: oklch(0.90 0.005 90);
    --sidebar-ring: oklch(0.55 0.12 175 / 0.3);
    --success: oklch(0.60 0.15 155);
    --warning: oklch(0.75 0.15 85);
    --info: oklch(0.60 0.12 240);
    --chart-1: oklch(0.55 0.12 175);
    --chart-2: oklch(0.60 0.15 155);
    --chart-3: oklch(0.75 0.15 85);
    --chart-4: oklch(0.60 0.12 240);
    --chart-5: oklch(0.55 0.2 25);
    --radius: 0.5rem;
  }

  .dark {
    --background: oklch(0.16 0.01 260);
    --foreground: oklch(0.93 0.005 90);
    --primary: oklch(0.65 0.13 175);
    --primary-hover: oklch(0.70 0.14 175);
    --primary-foreground: oklch(0.15 0.01 175);
    --primary-light: oklch(0.22 0.04 175);
    --card: oklch(0.19 0.01 260);
    --card-foreground: oklch(0.93 0.005 90);
    --popover: oklch(0.19 0.01 260);
    --popover-foreground: oklch(0.93 0.005 90);
    --secondary: oklch(0.22 0.01 260);
    --secondary-foreground: oklch(0.93 0.005 90);
    --muted: oklch(0.22 0.01 260);
    --muted-foreground: oklch(0.65 0.008 90);
    --accent: oklch(0.22 0.01 260);
    --accent-foreground: oklch(0.93 0.005 90);
    --destructive: oklch(0.60 0.2 25);
    --destructive-foreground: oklch(0.93 0.005 90);
    --border: oklch(0.28 0.01 260);
    --input: oklch(0.28 0.01 260);
    --ring: oklch(0.65 0.13 175 / 0.3);
    --sidebar: oklch(0.14 0.01 260);
    --sidebar-foreground: oklch(0.93 0.005 90);
    --sidebar-primary: oklch(0.65 0.13 175);
    --sidebar-primary-foreground: oklch(0.15 0.01 175);
    --sidebar-accent: oklch(0.22 0.01 260);
    --sidebar-accent-foreground: oklch(0.93 0.005 90);
    --sidebar-border: oklch(0.28 0.01 260);
    --sidebar-ring: oklch(0.65 0.13 175 / 0.3);
    --success: oklch(0.65 0.15 155);
    --warning: oklch(0.78 0.15 85);
    --info: oklch(0.65 0.12 240);
    --chart-1: oklch(0.65 0.13 175);
    --chart-2: oklch(0.65 0.15 155);
    --chart-3: oklch(0.78 0.15 85);
    --chart-4: oklch(0.65 0.12 240);
    --chart-5: oklch(0.60 0.2 25);
  }
  ```

  Update the `@theme inline` block:
  - Change `--font-sans` to `'Geist Sans', 'Noto Sans SC', system-ui, sans-serif`
  - Add `--font-mono: 'Geist Mono', 'JetBrains Mono', monospace;`
  - Add mappings for new semantic colors: `--color-success`, `--color-warning`, `--color-info`, `--color-primary-hover`, `--color-primary-light`
  - Keep existing radius/sidebar/color mappings updated to match new vars

  Simplify `body` styles — remove all gradients, noise pseudo-elements, and island-specific CSS. Body should just be:
  ```css
  html, body, #app {
    min-height: 100%;
  }

  @layer base {
    * {
      @apply border-border outline-ring/50;
    }
    body {
      background-color: var(--background);
      color: var(--foreground);
      font-family: var(--font-sans);
      -webkit-font-smoothing: antialiased;
      -moz-osx-font-smoothing: grayscale;
    }
  }
  ```

- [ ] **1.3** Update `console/components.json`:
  - Change `"baseColor": "zinc"` to `"baseColor": "stone"`
  - Change `"iconLibrary": "lucide"` to `"iconLibrary": "phosphor"`

- [ ] **1.4** Add Geist font import to `console/src/routes/__root.tsx` or the styles.css entry point. Since Geist is from npm (`geist` package), add to the top of `styles.css`:
  ```css
  @import 'geist/font/sans.css';
  @import 'geist/font/mono.css';
  ```

- [ ] **1.5** Verify: Run `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build` — must pass with zero errors.

### Files to modify
- `console/src/styles.css` — full rewrite
- `console/components.json` — baseColor + iconLibrary
- `console/package.json` — deps change (via bun add/remove)

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
```

### Notes
- The `geist` npm package exports CSS files at `geist/font/sans.css` and `geist/font/mono.css`. Import these in `styles.css` before the Tailwind import.
- Noto Sans SC is loaded from Google Fonts via `@import url(...)`.
- Any existing shadcn/ui components that reference Lucide icons will break after `bun remove lucide-react`. The components in `console/src/components/ui/` do NOT directly import lucide (verified: only `form.tsx` might, but shadcn new-york style uses slot-based icons). If build fails due to lucide imports in UI components, replace those imports with Phosphor equivalents inline.
- The old styles.css has custom CSS classes (`.island-shell`, `.feature-card`, `.nav-link`, etc.) used by `routes/index.tsx`. That file will be rewritten in Task 4 to a redirect, so these classes can be safely removed.

---

## Task 2: Fix auth types + useMe hook

**Goal:** Add `permissions` field to `MeResponse`, fix `useMe` to use `useQuery` (not the non-existent `createQuery`), and fix `QueryProvider` to use the singleton `queryClient`.

**Estimated time:** 3 minutes

### Steps

- [ ] **2.1** Update `console/src/features/auth/types/auth.ts` — add `permissions: string[]` to `MeResponse`:
  ```typescript
  export interface MeResponse {
    id: string;
    email: string;
    name: string;
    avatar?: string;
    role: string;
    siteIds: string[];
    permissions: string[];
  }
  ```

- [ ] **2.2** Fix `console/src/features/auth/hooks/useMe.ts` — replace `createQuery` (does not exist in TanStack Query v5) with `useQuery`, update query key to `['auth', 'me']`, reduce staleTime to 60s (matches auth guard):
  ```typescript
  import { useQuery } from '@tanstack/react-query';
  import { apiClient } from '../../shared';
  import type { MeResponse } from '../types/auth';

  export function useMe() {
    return useQuery({
      queryKey: ['auth', 'me'],
      queryFn: async (): Promise<MeResponse> => {
        return apiClient.get<MeResponse>('/auth/me');
      },
      staleTime: 60 * 1000, // 1 minute — balance security vs performance
    });
  }
  ```

- [ ] **2.3** Fix `console/src/features/shared/components/QueryProvider.tsx` — replace the local `new QueryClient()` with the singleton from `@/queryClient`:
  ```typescript
  import type { ReactNode } from 'react';
  import { QueryClientProvider } from '@tanstack/react-query';
  import { queryClient } from '@/queryClient';

  export interface QueryProviderProps {
    children: ReactNode;
  }

  export function QueryProvider({ children }: QueryProviderProps): ReactNode {
    return (
      <QueryClientProvider client={queryClient}>
        {children}
      </QueryClientProvider>
    );
  }
  ```

- [ ] **2.4** Update `console/src/features/auth/hooks/useLogin.ts` — change `queryKey: ['me']` to `queryKey: ['auth', 'me']` in `onSuccess` invalidation to match the new key:
  ```typescript
  onSuccess: () => {
    queryClient.invalidateQueries({ queryKey: ['auth', 'me'] });
  },
  ```

- [ ] **2.5** Verify: Run `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test` — must pass.

### Files to modify
- `console/src/features/auth/types/auth.ts`
- `console/src/features/auth/hooks/useMe.ts`
- `console/src/features/auth/hooks/useLogin.ts`
- `console/src/features/shared/components/QueryProvider.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
```

### Notes
- `createQuery` is a SolidJS API, not React. TanStack Query v5 for React uses `useQuery`.
- The `queryClient` singleton is already correctly used by `integrations/tanstack-query/root-provider.tsx` (the one wired in `__root.tsx`). The `features/shared/components/QueryProvider.tsx` is a separate provider that `ConsoleProvider` uses — it creates its own QueryClient, causing dual-client issues. Fix it to use the singleton.
- Query key change from `['me']` to `['auth', 'me']` is a breaking change — ensure `useLogin.ts` invalidation also updates.

---

## Task 3: API client 401 interceptor

**Goal:** Add automatic token refresh to `api-client.ts`. When any request gets a 401, attempt to refresh the access token via `/api/v1/auth/refresh` (httpOnly cookie carries refresh token). Concurrent requests share the same refresh promise. On refresh failure, redirect to `/login`.

**Estimated time:** 5 minutes

### Steps

- [ ] **3.1** Write the test first: Create `console/src/features/shared/__tests__/api-client.test.ts`:
  ```typescript
  import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

  // Tests for api-client 401 interceptor behavior:
  // 1. Normal successful request returns data
  // 2. 401 response triggers refresh, then retries original request with new token
  // 3. Concurrent 401s share the same refresh promise (only one /auth/refresh call)
  // 4. Refresh failure (401 on /auth/refresh) redirects to /login and clears state
  // 5. Non-401 errors propagate as ApiError without refresh attempt
  // 6. /auth/refresh and /auth/login endpoints are excluded from 401 interception
  ```

  Mock `fetch` globally. Test the following scenarios:
  - **Happy path**: fetch returns 200 → data returned
  - **401 → refresh → retry**: first fetch returns 401, refresh returns new token, retry succeeds
  - **Concurrent 401s**: two requests both get 401, only one refresh call is made, both retry
  - **Refresh failure**: refresh itself returns 401 → `window.location.href` set to `/login`
  - **Non-401 error**: 403/500 → `ApiError` thrown without refresh attempt
  - **Auth endpoints excluded**: requests to `/auth/refresh` or `/auth/login` skip interceptor

- [ ] **3.2** Run test, confirm RED: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test src/features/shared/__tests__/api-client.test.ts`

- [ ] **3.3** Implement 401 interceptor in `console/src/features/shared/api-client.ts`:

  Add module-level state:
  ```typescript
  let accessToken: string | null = null;
  let refreshPromise: Promise<string> | null = null;

  export function setAccessToken(token: string | null) {
    accessToken = token;
  }

  export function getAccessToken(): string | null {
    return accessToken;
  }
  ```

  Modify the `ApiClient` class:
  - In every method (`get`, `post`, `put`, `patch`, `delete`, `upload`), add `Authorization: Bearer ${accessToken}` header when `accessToken` is set
  - Replace `handleResponse` with a new `request` method that handles the full fetch + 401 retry cycle:

  ```typescript
  private async request<T>(url: string, init: RequestInit): Promise<T> {
    // Add auth header
    if (accessToken) {
      init.headers = { ...init.headers, Authorization: `Bearer ${accessToken}` };
    }
    init.credentials = 'include';

    let response = await fetch(url, init);

    // Skip interceptor for auth endpoints
    const isAuthEndpoint = url.includes('/auth/refresh') || url.includes('/auth/login');

    if (response.status === 401 && !isAuthEndpoint) {
      try {
        const newToken = await this.refreshAccessToken();
        // Retry with new token
        init.headers = { ...init.headers, Authorization: `Bearer ${newToken}` };
        response = await fetch(url, init);
      } catch {
        // Refresh failed — redirect to login
        accessToken = null;
        window.location.href = '/login';
        throw new ApiError('Session expired', 401);
      }
    }

    return this.handleResponse<T>(response);
  }

  private async refreshAccessToken(): Promise<string> {
    if (refreshPromise) {
      return refreshPromise;
    }

    refreshPromise = (async () => {
      try {
        const response = await fetch(`${this.baseURL}/auth/refresh`, {
          method: 'POST',
          credentials: 'include',
        });
        if (!response.ok) {
          throw new ApiError('Refresh failed', response.status);
        }
        const data = await response.json();
        accessToken = data.token;
        return data.token as string;
      } finally {
        refreshPromise = null;
      }
    })();

    return refreshPromise;
  }
  ```

  Refactor all HTTP methods to use `this.request(url, init)` instead of direct `fetch` + `handleResponse`.

- [ ] **3.4** Run test, confirm GREEN: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test src/features/shared/__tests__/api-client.test.ts`

- [ ] **3.5** Update `console/src/features/shared/index.ts` to export `setAccessToken` and `getAccessToken`:
  ```typescript
  export { apiClient, ApiError, setAccessToken, getAccessToken } from './api-client';
  ```

- [ ] **3.6** Wire token storage into login flow: Update `console/src/features/auth/hooks/useLogin.ts` — in `onSuccess`, call `setAccessToken(data.token)`:
  ```typescript
  import { setAccessToken } from '../../shared';
  // ...
  onSuccess: (data) => {
    setAccessToken(data.token);
    queryClient.invalidateQueries({ queryKey: ['auth', 'me'] });
  },
  ```

- [ ] **3.7** Verify all tests pass: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test`

### Files to create
- `console/src/features/shared/__tests__/api-client.test.ts`

### Files to modify
- `console/src/features/shared/api-client.ts`
- `console/src/features/shared/index.ts`
- `console/src/features/auth/hooks/useLogin.ts`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
```

### Notes
- Access token is stored in a module-level variable (memory only) — not localStorage. This is intentional for security (XSS cannot read it).
- Refresh token is in an httpOnly cookie — the browser automatically sends it with `credentials: 'include'`.
- The `refreshPromise` deduplication is critical: if 5 concurrent requests all get 401, only 1 refresh call fires, and all 5 await the same promise then retry.
- Auth endpoints (`/auth/refresh`, `/auth/login`) must NOT trigger the 401 interceptor to avoid infinite loops.
- `window.location.href = '/login'` is used (not TanStack Router navigate) because the api-client module has no access to the router instance.

---

## Task 4: Router type augmentation + route index redirect

**Goal:** Create the TanStack Router module augmentation for `staticData.title` (used by breadcrumbs) and convert the index route to a redirect.

**Estimated time:** 2 minutes

### Steps

- [ ] **4.1** Create `console/src/types/router.d.ts`:
  ```typescript
  import '@tanstack/react-router'

  declare module '@tanstack/react-router' {
    interface StaticDataRouteOption {
      title?: string
    }
  }
  ```

- [ ] **4.2** Rewrite `console/src/routes/index.tsx` — replace the entire island hero page with a redirect to `/dashboard`:
  ```typescript
  import { createFileRoute, redirect } from '@tanstack/react-router'

  export const Route = createFileRoute('/')({
    beforeLoad: () => {
      throw redirect({ to: '/dashboard' })
    },
  })
  ```

- [ ] **4.3** Verify: Run `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build` — TypeScript compilation must pass, route tree regeneration must succeed.

### Files to create
- `console/src/types/router.d.ts`

### Files to modify
- `console/src/routes/index.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
```

### Notes
- The `router.d.ts` module augmentation allows every route file to use `staticData: { title: 'Page Name' }` which the Header component reads via `useMatches()` for breadcrumbs.
- The old `index.tsx` had the tropical island template content — all those CSS classes (`.island-shell`, `.feature-card`, etc.) were removed in Task 1, so this file must be rewritten anyway.

---

## Task 5: AuthLayout + auth routes

**Goal:** Create the auth layout (centered card, logo, no sidebar) and wire the three auth routes (`/login`, `/forgot-password`, `/reset-password`).

**Estimated time:** 4 minutes

### Steps

- [ ] **5.1** Write test first: Create `console/src/components/layouts/__tests__/AuthLayout.test.tsx`:
  ```typescript
  import { describe, it, expect } from 'vitest';
  import { render, screen } from '@testing-library/react';
  import { AuthLayout } from '../AuthLayout';

  // Test cases:
  // 1. Renders children inside a centered card container
  // 2. Renders the Sky Flux CMS logo/title
  // 3. Has max-w-[400px] constraint on the card
  // 4. Centers content vertically and horizontally (min-h-screen flex)
  ```

- [ ] **5.2** Run test, confirm RED.

- [ ] **5.3** Create `console/src/components/layouts/AuthLayout.tsx`:
  ```typescript
  import type { ReactNode } from 'react';
  import { Card, CardContent } from '@/components/ui/card';

  interface AuthLayoutProps {
    children: ReactNode;
  }

  export function AuthLayout({ children }: AuthLayoutProps) {
    return (
      <div className="flex min-h-screen items-center justify-center bg-background p-4">
        <div className="w-full max-w-[400px]">
          <div className="mb-8 text-center">
            <h1 className="text-2xl font-bold text-foreground">Sky Flux CMS</h1>
          </div>
          <Card>
            <CardContent className="pt-6">
              {children}
            </CardContent>
          </Card>
        </div>
      </div>
    );
  }
  ```

- [ ] **5.4** Run test, confirm GREEN.

- [ ] **5.5** Create `console/src/routes/_auth.tsx` — layout route using AuthLayout:
  ```typescript
  import { createFileRoute, Outlet } from '@tanstack/react-router';
  import { AuthLayout } from '@/components/layouts/AuthLayout';

  export const Route = createFileRoute('/_auth')({
    component: AuthLayoutRoute,
  });

  function AuthLayoutRoute() {
    return (
      <AuthLayout>
        <Outlet />
      </AuthLayout>
    );
  }
  ```

- [ ] **5.6** Create `console/src/routes/_auth/login.tsx`:
  ```typescript
  import { createFileRoute, useNavigate } from '@tanstack/react-router';
  import { LoginForm } from '@/features/auth/components/LoginForm';

  export const Route = createFileRoute('/_auth/login')({
    staticData: { title: 'Login' },
    component: LoginPage,
  });

  function LoginPage() {
    const navigate = useNavigate();
    return (
      <div className="space-y-4">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Sign In</h2>
          <p className="text-sm text-muted-foreground">
            Enter your credentials to access the dashboard
          </p>
        </div>
        <LoginForm onSuccess={() => navigate({ to: '/dashboard' })} />
        <div className="text-center text-sm">
          <a href="/forgot-password" className="text-primary hover:underline">
            Forgot password?
          </a>
        </div>
      </div>
    );
  }
  ```

- [ ] **5.7** Create `console/src/routes/_auth/forgot-password.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';

  export const Route = createFileRoute('/_auth/forgot-password')({
    staticData: { title: 'Forgot Password' },
    component: ForgotPasswordPage,
  });

  function ForgotPasswordPage() {
    return (
      <div className="space-y-4">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Forgot Password</h2>
          <p className="text-sm text-muted-foreground">
            Enter your email to receive a password reset link
          </p>
        </div>
        <p className="text-sm text-muted-foreground">Coming soon</p>
      </div>
    );
  }
  ```

- [ ] **5.8** Create `console/src/routes/_auth/reset-password.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';

  export const Route = createFileRoute('/_auth/reset-password')({
    staticData: { title: 'Reset Password' },
    component: ResetPasswordPage,
  });

  function ResetPasswordPage() {
    return (
      <div className="space-y-4">
        <div className="space-y-1">
          <h2 className="text-lg font-semibold">Reset Password</h2>
          <p className="text-sm text-muted-foreground">
            Set a new password for your account
          </p>
        </div>
        <p className="text-sm text-muted-foreground">Coming soon</p>
      </div>
    );
  }
  ```

- [ ] **5.9** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build`

### Files to create
- `console/src/components/layouts/AuthLayout.tsx`
- `console/src/components/layouts/__tests__/AuthLayout.test.tsx`
- `console/src/routes/_auth.tsx`
- `console/src/routes/_auth/login.tsx`
- `console/src/routes/_auth/forgot-password.tsx`
- `console/src/routes/_auth/reset-password.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- `_auth` prefix makes it a pathless layout route — it does NOT add `/auth/` to the URL. Routes render at `/login`, `/forgot-password`, `/reset-password`.
- The `Card` component must already exist in `console/src/components/ui/`. If not, install it first: `cd console && bunx shadcn@latest add card`.
- Check if `card.tsx` exists in `console/src/components/ui/` before writing the AuthLayout import. The glob showed it is NOT currently present — it may need to be installed. **If missing, add a step 5.2.5 to run `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bunx shadcn@latest add card`.**

---

## Task 6: Sidebar component

**Goal:** Create the sidebar navigation with 4 groups, Phosphor duotone icons, RBAC permission filtering, and active route highlighting.

**Estimated time:** 5 minutes

### Steps

- [ ] **6.1** Write test first: Create `console/src/components/layouts/__tests__/Sidebar.test.tsx`:
  ```typescript
  import { describe, it, expect } from 'vitest';
  import { render, screen } from '@testing-library/react';
  // ... wrapper with router context + mock user context

  // Test cases:
  // 1. Renders all 4 navigation groups (Dashboard, Content, Site, System)
  // 2. Renders correct menu items within each group
  // 3. Hides menu items when user lacks permissions (e.g., no 'users.manage' → no Users link)
  // 4. Shows all items when user has all permissions
  // 5. Highlights active route (aria-current="page" or active class)
  // 6. Renders Phosphor duotone icons
  // 7. Renders user info at bottom (name + logout button)
  ```

- [ ] **6.2** Run test, confirm RED.

- [ ] **6.3** Create `console/src/components/layouts/Sidebar.tsx`:

  Navigation structure (from spec Section 5):
  ```typescript
  import { Link, useRouterState } from '@tanstack/react-router';
  import {
    ChartBar,
    Article,
    FolderOpen,
    Tag,
    Image,
    ChatCircle,
    LinkSimple,
    ArrowUUpLeft,
    Gear,
    Users,
    Shield,
    Key,
    ClipboardText,
    SignOut,
  } from '@phosphor-icons/react';
  import { ScrollArea } from '@/components/ui/scroll-area';
  import { Button } from '@/components/ui/button';

  interface SidebarProps {
    user: {
      name: string;
      email: string;
      permissions: string[];
    };
    onLogout: () => void;
    onClose?: () => void; // For mobile sheet
  }

  const navGroups = [
    {
      items: [
        { label: 'Dashboard', icon: ChartBar, to: '/dashboard', permission: null },
      ],
    },
    {
      title: 'Content',
      items: [
        { label: 'Posts', icon: Article, to: '/posts', permission: 'posts.read' },
        { label: 'Categories', icon: FolderOpen, to: '/categories', permission: 'categories.read' },
        { label: 'Tags', icon: Tag, to: '/tags', permission: 'tags.read' },
        { label: 'Media', icon: Image, to: '/media', permission: 'media.read' },
        { label: 'Comments', icon: ChatCircle, to: '/comments', permission: 'comments.read' },
      ],
    },
    {
      title: 'Site',
      items: [
        { label: 'Menus', icon: LinkSimple, to: '/menus', permission: 'menus.read' },
        { label: 'Redirects', icon: ArrowUUpLeft, to: '/redirects', permission: 'redirects.read' },
        { label: 'Settings', icon: Gear, to: '/settings', permission: 'settings.read' },
      ],
    },
    {
      title: 'System',
      items: [
        { label: 'Users', icon: Users, to: '/users', permission: 'users.manage' },
        { label: 'Roles', icon: Shield, to: '/roles', permission: 'roles.manage' },
        { label: 'API Keys', icon: Key, to: '/api-keys', permission: 'api_keys.manage' },
        { label: 'Audit Log', icon: ClipboardText, to: '/audit', permission: 'audit.read' },
      ],
    },
  ];
  ```

  - Filter items by `user.permissions.includes(item.permission)` (null permission = always visible)
  - Active state: use `useRouterState({ select: s => s.location.pathname })` and check `pathname.startsWith(item.to)`
  - Active item gets `bg-primary/10 text-primary font-medium` styling
  - Width: `w-60` (240px)
  - Bottom section: user name + email (truncated) + logout button with `SignOut` icon

- [ ] **6.4** Run test, confirm GREEN.

- [ ] **6.5** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build`

### Files to create
- `console/src/components/layouts/Sidebar.tsx`
- `console/src/components/layouts/__tests__/Sidebar.test.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- All Phosphor icons use `weight="duotone"` for the warm dual-tone effect.
- `ScrollArea` may need to be installed: `cd console && bunx shadcn@latest add scroll-area`. Check first.
- The Sidebar is a pure presentational component — it receives `user` and `onLogout` as props (no direct hook calls). This makes testing straightforward without needing to mock route context for user data.
- Mobile behavior is handled by the parent `DashboardLayout` which wraps Sidebar in a `Sheet` on small screens. The `onClose` prop is called after navigation on mobile.
- Test wrapper must provide TanStack Router context. Use `createMemoryHistory` + `createRootRoute` + `createRouter` for test isolation. Alternatively, mock `useRouterState` via `vi.mock`.

---

## Task 7: Header component

**Goal:** Create the header with breadcrumbs from route `staticData.title`, theme toggle (light/dark/system), and user dropdown menu.

**Estimated time:** 5 minutes

### Steps

- [ ] **7.1** Write test first: Create `console/src/components/layouts/__tests__/Header.test.tsx`:
  ```typescript
  // Test cases:
  // 1. Renders breadcrumbs from staticData.title
  // 2. Renders theme toggle button (sun/moon icon based on current theme)
  // 3. Renders user avatar with dropdown
  // 4. Dropdown shows user name, email, and logout option
  // 5. Renders mobile hamburger button (visible on small screens)
  // 6. Clicking theme toggle cycles through light/dark/system
  ```

- [ ] **7.2** Run test, confirm RED.

- [ ] **7.3** Install shadcn components if missing:
  ```bash
  cd /Users/martinadamsdev/workspace/sky-flux-cms/console
  bunx shadcn@latest add dropdown-menu avatar breadcrumb separator sheet
  ```
  (Check which ones already exist first — `dropdown-menu`, `avatar`, `breadcrumb`, `separator`, `sheet` are needed. The glob showed only `button`, `checkbox`, `dialog`, `form`, `input`, `label`, `select`, `slider`, `switch`, `textarea` currently exist.)

- [ ] **7.4** Create `console/src/components/layouts/Header.tsx`:
  ```typescript
  import { useMatches } from '@tanstack/react-router';
  import { Sun, Moon, List, UserCircle } from '@phosphor-icons/react';
  import { useTheme } from '@/features/shared/components/useTheme';
  import {
    Breadcrumb, BreadcrumbItem, BreadcrumbLink, BreadcrumbList,
    BreadcrumbSeparator
  } from '@/components/ui/breadcrumb';
  import {
    DropdownMenu, DropdownMenuContent, DropdownMenuItem,
    DropdownMenuLabel, DropdownMenuSeparator, DropdownMenuTrigger
  } from '@/components/ui/dropdown-menu';
  import { Button } from '@/components/ui/button';

  interface HeaderProps {
    user: { name: string; email: string; avatar?: string };
    onLogout: () => void;
    onMobileMenuToggle: () => void;
  }
  ```

  Breadcrumbs logic:
  ```typescript
  const matches = useMatches();
  const crumbs = matches
    .filter(match => match.staticData?.title)
    .map(match => ({
      title: match.staticData.title as string,
      path: match.fullPath,
    }));
  ```

  Theme toggle: cycle `light → dark → system → light`. Display `Sun` for light, `Moon` for dark, `Monitor` for system (or just toggle light/dark).

  Mobile: `<Button variant="ghost" size="icon" className="lg:hidden">` with `List` (hamburger) icon, calls `onMobileMenuToggle`.

  Height: `h-16` (64px). Sticky: `sticky top-0 z-30 bg-background/95 backdrop-blur`.

- [ ] **7.5** Run test, confirm GREEN.

- [ ] **7.6** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build`

### Files to create
- `console/src/components/layouts/Header.tsx`
- `console/src/components/layouts/__tests__/Header.test.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- `useMatches()` returns all matched route entries in the tree. Each with `staticData` and `fullPath`. Filter for entries that have a `title` in `staticData`.
- Theme toggle uses the existing `useTheme()` hook from `features/shared/components/useTheme.ts`. It provides `{ theme, setTheme, resolvedTheme }`.
- The user dropdown needs `DropdownMenu` from shadcn. Must be installed first.
- The Header is a presentational component receiving `user`, `onLogout`, `onMobileMenuToggle` as props.
- `localStorage.setItem('theme', theme)` should be handled in the ThemeProvider, not the Header. The Header just calls `setTheme()`.
- Update the existing `ThemeProvider` to persist theme choice to localStorage and read it on mount (if not already doing so).

---

## Task 8: DashboardLayout + auth guard

**Goal:** Create the dashboard layout (sidebar + header + outlet) with auth guard in `beforeLoad` and mobile-responsive sidebar Sheet.

**Estimated time:** 5 minutes

### Steps

- [ ] **8.1** Write test first: Create `console/src/components/layouts/__tests__/DashboardLayout.test.tsx`:
  ```typescript
  // Test cases:
  // 1. Renders Sidebar, Header, and children/outlet area
  // 2. Sidebar is visible on desktop (lg+)
  // 3. Mobile: sidebar hidden, hamburger visible
  // 4. Passes user data to Sidebar and Header
  ```

- [ ] **8.2** Run test, confirm RED.

- [ ] **8.3** Create `console/src/components/layouts/DashboardLayout.tsx`:
  ```typescript
  import { useState, Suspense, type ReactNode } from 'react';
  import { Sidebar } from './Sidebar';
  import { Header } from './Header';
  import { PageSkeleton } from '@/components/shared/PageSkeleton';
  import { Sheet, SheetContent } from '@/components/ui/sheet';

  interface DashboardLayoutProps {
    user: {
      name: string;
      email: string;
      avatar?: string;
      permissions: string[];
    };
    onLogout: () => void;
    children: ReactNode;
  }

  export function DashboardLayout({ user, onLogout, children }: DashboardLayoutProps) {
    const [mobileOpen, setMobileOpen] = useState(false);

    return (
      <div className="flex min-h-screen">
        {/* Desktop sidebar */}
        <aside className="hidden lg:flex lg:w-60 lg:flex-col lg:border-r lg:border-border lg:bg-sidebar">
          <Sidebar user={user} onLogout={onLogout} />
        </aside>

        {/* Mobile sidebar sheet */}
        <Sheet open={mobileOpen} onOpenChange={setMobileOpen}>
          <SheetContent side="left" className="w-60 p-0">
            <Sidebar user={user} onLogout={onLogout} onClose={() => setMobileOpen(false)} />
          </SheetContent>
        </Sheet>

        {/* Main content */}
        <div className="flex flex-1 flex-col">
          <Header
            user={user}
            onLogout={onLogout}
            onMobileMenuToggle={() => setMobileOpen(true)}
          />
          <main className="flex-1 p-6">
            <Suspense fallback={<PageSkeleton />}>
              {children}
            </Suspense>
          </main>
        </div>
      </div>
    );
  }
  ```

- [ ] **8.4** Run test, confirm GREEN.

- [ ] **8.5** Create `console/src/routes/_dashboard.tsx` with auth guard:
  ```typescript
  import { createFileRoute, Outlet, redirect } from '@tanstack/react-router';
  import { DashboardLayout } from '@/components/layouts/DashboardLayout';
  import { apiClient } from '@/features/shared';
  import { setAccessToken } from '@/features/shared/api-client';
  import type { MeResponse } from '@/features/auth/types/auth';

  export const Route = createFileRoute('/_dashboard')({
    beforeLoad: async ({ context }) => {
      try {
        const user = await context.queryClient.ensureQueryData({
          queryKey: ['auth', 'me'],
          queryFn: () => apiClient.get<MeResponse>('/auth/me'),
          staleTime: 60 * 1000,
        });
        return { user };
      } catch {
        throw redirect({ to: '/login' });
      }
    },
    component: DashboardLayoutRoute,
  });

  function DashboardLayoutRoute() {
    const { user } = Route.useRouteContext();
    const navigate = Route.useNavigate();

    const handleLogout = async () => {
      try {
        await apiClient.post('/auth/logout');
      } finally {
        setAccessToken(null);
        navigate({ to: '/login' });
      }
    };

    return (
      <DashboardLayout user={user} onLogout={handleLogout}>
        <Outlet />
      </DashboardLayout>
    );
  }
  ```

- [ ] **8.6** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build`

### Files to create
- `console/src/components/layouts/DashboardLayout.tsx`
- `console/src/components/layouts/__tests__/DashboardLayout.test.tsx`
- `console/src/routes/_dashboard.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- `ensureQueryData` is the key: it checks the TanStack Query cache first (within `staleTime`). If cached data exists and is fresh, no API call is made. Otherwise it fetches `/api/v1/auth/me`.
- On 401 (user not authenticated), the api-client interceptor will try to refresh. If refresh also fails, `ensureQueryData` throws, and `beforeLoad` catches it and redirects to `/login`.
- `Route.useRouteContext()` gives typed access to `{ user }` returned from `beforeLoad`. All child routes inherit this context.
- The `Sheet` component (mobile sidebar) must be installed: `bunx shadcn@latest add sheet` (verified NOT in current UI components).
- Logout clears the access token from memory and navigates to `/login`. The server-side logout (`POST /auth/logout`) invalidates the refresh token.

---

## Task 9: usePermission hook + shared components

**Goal:** Create the `usePermission` hook for RBAC checks, `PageSkeleton` for loading states, `EmptyState` for empty lists, and the 403 page.

**Estimated time:** 4 minutes

### Steps

- [ ] **9.1** Write tests first: Create `console/src/hooks/__tests__/usePermission.test.tsx`:
  ```typescript
  // Test cases:
  // 1. Returns true when user has the specified permission
  // 2. Returns false when user lacks the permission
  // 3. hasAny returns true if user has at least one of the permissions
  // 4. hasAll returns true only if user has ALL of the permissions
  ```

- [ ] **9.2** Write tests: Create `console/src/components/shared/__tests__/EmptyState.test.tsx`:
  ```typescript
  // Test cases:
  // 1. Renders icon, title, description
  // 2. Renders action button when provided
  // 3. Does not render action when not provided
  ```

- [ ] **9.3** Run tests, confirm RED.

- [ ] **9.4** Create `console/src/hooks/usePermission.ts`:
  ```typescript
  import { useRouteContext } from '@tanstack/react-router';
  import type { MeResponse } from '@/features/auth/types/auth';

  export function usePermission(permission: string): boolean {
    const { user } = useRouteContext({ from: '/_dashboard' }) as { user: MeResponse };
    return user.permissions.includes(permission);
  }

  export function usePermissions() {
    const { user } = useRouteContext({ from: '/_dashboard' }) as { user: MeResponse };

    return {
      has: (permission: string) => user.permissions.includes(permission),
      hasAny: (...permissions: string[]) =>
        permissions.some(p => user.permissions.includes(p)),
      hasAll: (...permissions: string[]) =>
        permissions.every(p => user.permissions.includes(p)),
      user,
    };
  }
  ```

- [ ] **9.5** Create `console/src/components/shared/PageSkeleton.tsx`:
  ```typescript
  import { Skeleton } from '@/components/ui/skeleton';

  export function PageSkeleton() {
    return (
      <div className="space-y-6">
        {/* Header skeleton */}
        <div className="space-y-2">
          <Skeleton className="h-8 w-48" />
          <Skeleton className="h-4 w-96" />
        </div>
        {/* Table skeleton: 5 rows */}
        <div className="space-y-3">
          <Skeleton className="h-10 w-full" />
          {Array.from({ length: 5 }).map((_, i) => (
            <Skeleton key={i} className="h-12 w-full" />
          ))}
        </div>
      </div>
    );
  }
  ```

  Note: `Skeleton` must exist. If not: `cd console && bunx shadcn@latest add skeleton`.

- [ ] **9.6** Create `console/src/components/shared/EmptyState.tsx`:
  ```typescript
  import type { ReactNode } from 'react';

  interface EmptyStateProps {
    icon?: ReactNode;
    title: string;
    description?: string;
    action?: ReactNode;
  }

  export function EmptyState({ icon, title, description, action }: EmptyStateProps) {
    return (
      <div className="flex min-h-[400px] flex-col items-center justify-center text-center">
        {icon && (
          <div className="mb-4 text-muted-foreground">{icon}</div>
        )}
        <h3 className="text-lg font-semibold">{title}</h3>
        {description && (
          <p className="mt-1 max-w-md text-sm text-muted-foreground">{description}</p>
        )}
        {action && <div className="mt-4">{action}</div>}
      </div>
    );
  }
  ```

- [ ] **9.7** Run tests, confirm GREEN.

- [ ] **9.8** Create `console/src/routes/_dashboard/403.tsx`:
  ```typescript
  import { createFileRoute, Link } from '@tanstack/react-router';
  import { ShieldWarning } from '@phosphor-icons/react';
  import { EmptyState } from '@/components/shared/EmptyState';
  import { Button } from '@/components/ui/button';

  export const Route = createFileRoute('/_dashboard/403')({
    staticData: { title: 'Permission Denied' },
    component: ForbiddenPage,
  });

  function ForbiddenPage() {
    return (
      <EmptyState
        icon={<ShieldWarning size={48} weight="duotone" />}
        title="Permission Denied"
        description="You don't have permission to access this page. Contact your administrator for access."
        action={
          <Button asChild>
            <Link to="/dashboard">Back to Dashboard</Link>
          </Button>
        }
      />
    );
  }
  ```

- [ ] **9.9** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build`

### Files to create
- `console/src/hooks/usePermission.ts`
- `console/src/hooks/__tests__/usePermission.test.tsx`
- `console/src/components/shared/PageSkeleton.tsx`
- `console/src/components/shared/EmptyState.tsx`
- `console/src/components/shared/__tests__/EmptyState.test.tsx`
- `console/src/routes/_dashboard/403.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- `usePermission` uses `useRouteContext({ from: '/_dashboard' })` — this only works inside `_dashboard` child routes (which is the only place it should be used).
- `PageSkeleton` is intentionally simple — a generic table-style skeleton. Feature-specific skeletons can be added later.
- The 403 page is a route within `_dashboard/` so the user still sees the sidebar and header. The `beforeLoad` in individual pages (like `_dashboard/users.tsx`) can `throw redirect({ to: '/dashboard/403' })` when permissions are missing.
- Install `skeleton` shadcn component if missing: `bunx shadcn@latest add skeleton`.

---

## Task 10: Dashboard route pages (wire existing feature components)

**Goal:** Create all `_dashboard/*.tsx` route files, wiring existing feature components where available and using `EmptyState` placeholders for missing features. Each route has `staticData.title` for breadcrumbs and `validateSearch` with Zod schema for URL state.

**Estimated time:** 5 minutes

### Steps

- [ ] **10.1** Create `console/src/routes/_dashboard/index.tsx` (Dashboard home):
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { ChartBar } from '@phosphor-icons/react';
  import { EmptyState } from '@/components/shared/EmptyState';

  export const Route = createFileRoute('/_dashboard/')({
    staticData: { title: 'Dashboard' },
    component: DashboardPage,
  });

  function DashboardPage() {
    return (
      <EmptyState
        icon={<ChartBar size={48} weight="duotone" />}
        title="Dashboard"
        description="Analytics and overview coming soon"
      />
    );
  }
  ```

- [ ] **10.2** Create `console/src/routes/_dashboard/posts/index.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { z } from 'zod';
  import { PostsTable } from '@/features/posts/components';

  const postsSearchSchema = z.object({
    page: z.number().default(1),
    perPage: z.number().default(20),
    status: z.enum(['all', 'draft', 'published', 'archived']).default('all'),
    search: z.string().optional(),
    sortBy: z.string().default('created_at'),
    sortOrder: z.enum(['asc', 'desc']).default('desc'),
  });

  export const Route = createFileRoute('/_dashboard/posts/')({
    staticData: { title: 'Posts' },
    validateSearch: postsSearchSchema,
    component: PostsPage,
  });

  function PostsPage() {
    return <PostsTable />;
  }
  ```

- [ ] **10.3** Create `console/src/routes/_dashboard/posts/new.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { Article } from '@phosphor-icons/react';
  import { EmptyState } from '@/components/shared/EmptyState';

  export const Route = createFileRoute('/_dashboard/posts/new')({
    staticData: { title: 'New Post' },
    component: NewPostPage,
  });

  function NewPostPage() {
    return (
      <EmptyState
        icon={<Article size={48} weight="duotone" />}
        title="Post Editor"
        description="Rich text editor coming soon"
      />
    );
  }
  ```

- [ ] **10.4** Create `console/src/routes/_dashboard/posts/$postId.edit.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { Article } from '@phosphor-icons/react';
  import { EmptyState } from '@/components/shared/EmptyState';

  export const Route = createFileRoute('/_dashboard/posts/$postId/edit')({
    staticData: { title: 'Edit Post' },
    component: EditPostPage,
  });

  function EditPostPage() {
    const { postId } = Route.useParams();
    return (
      <EmptyState
        icon={<Article size={48} weight="duotone" />}
        title="Edit Post"
        description={`Editing post ${postId} — coming soon`}
      />
    );
  }
  ```

- [ ] **10.5** Create `console/src/routes/_dashboard/categories.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { CategoryTree } from '@/features/categories/components';

  export const Route = createFileRoute('/_dashboard/categories')({
    staticData: { title: 'Categories' },
    component: CategoriesPage,
  });

  function CategoriesPage() {
    return <CategoryTree />;
  }
  ```

- [ ] **10.6** Create `console/src/routes/_dashboard/tags.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { z } from 'zod';
  import { TagsTable } from '@/features/tags/components';

  const tagsSearchSchema = z.object({
    page: z.number().default(1),
    perPage: z.number().default(20),
    search: z.string().optional(),
  });

  export const Route = createFileRoute('/_dashboard/tags')({
    staticData: { title: 'Tags' },
    validateSearch: tagsSearchSchema,
    component: TagsPage,
  });

  function TagsPage() {
    return <TagsTable />;
  }
  ```

- [ ] **10.7** Create `console/src/routes/_dashboard/media.tsx`:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { z } from 'zod';
  import { MediaLibrary } from '@/features/media/components';

  const mediaSearchSchema = z.object({
    page: z.number().default(1),
    perPage: z.number().default(20),
    search: z.string().optional(),
    type: z.enum(['all', 'image', 'video', 'document']).default('all'),
  });

  export const Route = createFileRoute('/_dashboard/media')({
    staticData: { title: 'Media' },
    validateSearch: mediaSearchSchema,
    component: MediaPage,
  });

  function MediaPage() {
    return <MediaLibrary />;
  }
  ```

- [ ] **10.8** Create `console/src/routes/_dashboard/users.tsx`:
  ```typescript
  import { createFileRoute, redirect } from '@tanstack/react-router';
  import { z } from 'zod';
  import { UsersTable } from '@/features/users/components';

  const usersSearchSchema = z.object({
    page: z.number().default(1),
    perPage: z.number().default(20),
    search: z.string().optional(),
  });

  export const Route = createFileRoute('/_dashboard/users')({
    staticData: { title: 'Users' },
    validateSearch: usersSearchSchema,
    beforeLoad: ({ context }) => {
      if (!context.user.permissions.includes('users.manage')) {
        throw redirect({ to: '/dashboard/403' });
      }
    },
    component: UsersPage,
  });

  function UsersPage() {
    return <UsersTable />;
  }
  ```

- [ ] **10.9** Create `console/src/routes/_dashboard/roles.tsx`:
  ```typescript
  import { createFileRoute, redirect } from '@tanstack/react-router';
  import { z } from 'zod';
  import { RolesTable } from '@/features/roles/components';

  const rolesSearchSchema = z.object({
    page: z.number().default(1),
    perPage: z.number().default(20),
    search: z.string().optional(),
  });

  export const Route = createFileRoute('/_dashboard/roles')({
    staticData: { title: 'Roles' },
    validateSearch: rolesSearchSchema,
    beforeLoad: ({ context }) => {
      if (!context.user.permissions.includes('roles.manage')) {
        throw redirect({ to: '/dashboard/403' });
      }
    },
    component: RolesPage,
  });

  function RolesPage() {
    return <RolesTable />;
  }
  ```

- [ ] **10.10** Create placeholder routes for features not yet built. Each follows the same pattern:
  - `console/src/routes/_dashboard/comments.tsx` — title: "Comments", permission: `comments.read`
  - `console/src/routes/_dashboard/menus.tsx` — title: "Menus", permission: `menus.read`
  - `console/src/routes/_dashboard/redirects.tsx` — title: "Redirects", permission: `redirects.read`
  - `console/src/routes/_dashboard/settings.tsx` — title: "Settings", permission: `settings.read`
  - `console/src/routes/_dashboard/api-keys.tsx` — title: "API Keys", permission: `api_keys.manage`
  - `console/src/routes/_dashboard/audit.tsx` — title: "Audit Log", permission: `audit.read`

  Each placeholder route:
  ```typescript
  import { createFileRoute } from '@tanstack/react-router';
  import { <RelevantIcon> } from '@phosphor-icons/react';
  import { EmptyState } from '@/components/shared/EmptyState';

  export const Route = createFileRoute('/_dashboard/<name>')({
    staticData: { title: '<Title>' },
    component: <Name>Page,
  });

  function <Name>Page() {
    return (
      <EmptyState
        icon={<<Icon> size={48} weight="duotone" />}
        title="<Title>"
        description="Coming soon"
      />
    );
  }
  ```

  Icons for placeholders:
  - Comments: `ChatCircle`
  - Menus: `LinkSimple`
  - Redirects: `ArrowUUpLeft`
  - Settings: `Gear`
  - API Keys: `Key`
  - Audit: `ClipboardText`

- [ ] **10.11** Verify: `cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build` — all routes compile, route tree generates correctly.

### Files to create
- `console/src/routes/_dashboard/index.tsx`
- `console/src/routes/_dashboard/posts/index.tsx`
- `console/src/routes/_dashboard/posts/new.tsx`
- `console/src/routes/_dashboard/posts/$postId.edit.tsx`
- `console/src/routes/_dashboard/categories.tsx`
- `console/src/routes/_dashboard/tags.tsx`
- `console/src/routes/_dashboard/media.tsx`
- `console/src/routes/_dashboard/users.tsx`
- `console/src/routes/_dashboard/roles.tsx`
- `console/src/routes/_dashboard/comments.tsx`
- `console/src/routes/_dashboard/menus.tsx`
- `console/src/routes/_dashboard/redirects.tsx`
- `console/src/routes/_dashboard/settings.tsx`
- `console/src/routes/_dashboard/api-keys.tsx`
- `console/src/routes/_dashboard/audit.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
```

### Notes
- **Zod v4 `z.object({}).default()`**: The `validateSearch` schemas use `.default()` on each field, NOT on the outer object. TanStack Router calls the schema's `parse` on the search params object.
- **Permission-guarded routes** (users, roles, api-keys, audit): use `beforeLoad` to check `context.user.permissions`. If missing, `throw redirect({ to: '/dashboard/403' })`. Content routes (posts, categories, tags, media) don't need page-level guards since the sidebar already hides them.
- **`$postId.edit.tsx` file naming**: TanStack Router file-based routing uses `$param` syntax. The file `$postId.edit.tsx` creates a route at `/posts/:postId/edit`. Confirm this with the TanStack Router docs — it may need to be `$postId/edit.tsx` (subdirectory) instead. Check the generated `routeTree.gen.ts` after build.
- Feature components (`PostsTable`, `CategoryTree`, `TagsTable`, `MediaLibrary`, `UsersTable`, `RolesTable`) are rendered directly. They manage their own data fetching via TanStack Query hooks internally. The route page just mounts them.

---

## Task 11: Update __root.tsx + final integration

**Goal:** Add `notFoundComponent` to `__root.tsx`, add Sonner `<Toaster />`, verify the complete route tree generates correctly, and run all tests + build.

**Estimated time:** 3 minutes

### Steps

- [ ] **11.1** Install Sonner if not already present:
  ```bash
  cd /Users/martinadamsdev/workspace/sky-flux-cms/console
  bunx shadcn@latest add sonner
  ```

- [ ] **11.2** Update `console/src/routes/__root.tsx`:
  ```typescript
  import { Outlet, createRootRouteWithContext, Link } from '@tanstack/react-router';
  import { TanStackRouterDevtools } from '@tanstack/react-router-devtools';
  import { ReactQueryDevtools } from '@tanstack/react-query-devtools';
  import { Toaster } from '@/components/ui/sonner';
  import { WarningCircle } from '@phosphor-icons/react';
  import type { QueryClient } from '@tanstack/react-query';
  import type { MeResponse } from '@/features/auth/types/auth';

  import TanStackQueryProvider from '../integrations/tanstack-query/root-provider';

  interface RouterContext {
    queryClient: QueryClient;
  }

  export const Route = createRootRouteWithContext<RouterContext>()({
    component: RootComponent,
    notFoundComponent: NotFoundComponent,
  });

  function RootComponent() {
    return (
      <TanStackQueryProvider>
        <Outlet />
        <Toaster position="bottom-right" richColors />
        {import.meta.env.DEV && (
          <>
            <TanStackRouterDevtools position="bottom-right" />
            <ReactQueryDevtools buttonPosition="bottom-left" />
          </>
        )}
      </TanStackQueryProvider>
    );
  }

  function NotFoundComponent() {
    return (
      <div className="flex min-h-screen items-center justify-center">
        <div className="text-center">
          <WarningCircle size={64} weight="duotone" className="mx-auto mb-4 text-muted-foreground" />
          <h1 className="text-2xl font-bold">Page Not Found</h1>
          <p className="mt-2 text-muted-foreground">
            The page you're looking for doesn't exist.
          </p>
          <Link to="/dashboard" className="mt-4 inline-block text-primary hover:underline">
            Back to Dashboard
          </Link>
        </div>
      </div>
    );
  }
  ```

  Key changes from current `__root.tsx`:
  - Remove `getLocale()` / `@/paraglide/runtime` import (console uses English only, paraglide is for the public web app)
  - Remove `beforeLoad` lang setting
  - Add `notFoundComponent`
  - Add `<Toaster />`
  - Update `MyRouterContext` → `RouterContext` (cleaner name)

- [ ] **11.3** Verify route tree generation — run build and check `console/src/routeTree.gen.ts` contains all expected routes:
  ```
  /_auth/login
  /_auth/forgot-password
  /_auth/reset-password
  /_dashboard/
  /_dashboard/posts/
  /_dashboard/posts/new
  /_dashboard/posts/$postId/edit
  /_dashboard/categories
  /_dashboard/tags
  /_dashboard/media
  /_dashboard/comments
  /_dashboard/menus
  /_dashboard/redirects
  /_dashboard/users
  /_dashboard/roles
  /_dashboard/settings
  /_dashboard/api-keys
  /_dashboard/audit
  /_dashboard/403
  ```

- [ ] **11.4** Run full verification:
  ```bash
  cd /Users/martinadamsdev/workspace/sky-flux-cms/console
  bun run test
  bun run build
  ```

- [ ] **11.5** Verify no TypeScript errors:
  ```bash
  cd /Users/martinadamsdev/workspace/sky-flux-cms/console
  npx tsc --noEmit
  ```

### Files to modify
- `console/src/routes/__root.tsx`

### Verification
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test && bun run build
```

### Notes
- The `paraglide` import in current `__root.tsx` is from the old CTA template. The console app does not use paraglide for i18n (it's English-only for the admin interface). Remove it.
- `Toaster` from shadcn/ui uses Sonner under the hood. The `position="bottom-right"` and `richColors` props match the spec.
- After all tasks complete, the route tree should have ~20 routes total. The `routeTree.gen.ts` is auto-generated by `TanStackRouterVite` plugin on build.
- If `paraglide` removal causes issues with the Vite plugin config in `vite.config.ts`, also remove the `paraglideVitePlugin` from `vite.config.ts` and uninstall `@inlang/paraglide-js`.

---

## Summary

| Task | Files Created | Files Modified | Tests Added | Estimated Time |
|------|--------------|----------------|-------------|----------------|
| 1. Dependencies + Theme | 0 | 3 (styles.css, components.json, package.json) | 0 (build check) | 3 min |
| 2. Auth types + useMe | 0 | 4 (auth.ts, useMe.ts, useLogin.ts, QueryProvider.tsx) | 0 (existing tests) | 3 min |
| 3. API client 401 | 1 (test) | 3 (api-client.ts, index.ts, useLogin.ts) | ~6 tests | 5 min |
| 4. Router augmentation | 1 (router.d.ts) | 1 (index.tsx) | 0 (build check) | 2 min |
| 5. AuthLayout + auth routes | 6 (layout + test + 4 routes) | 0 | ~4 tests | 4 min |
| 6. Sidebar | 2 (component + test) | 0 | ~7 tests | 5 min |
| 7. Header | 2 (component + test) | 0 | ~6 tests | 5 min |
| 8. DashboardLayout + guard | 3 (layout + test + route) | 0 | ~4 tests | 5 min |
| 9. usePermission + shared | 6 (hook + 2 components + 2 tests + 403 route) | 0 | ~7 tests | 4 min |
| 10. Dashboard routes | 15 (route files) | 0 | 0 (build check) | 5 min |
| 11. Root + integration | 0 | 1 (__root.tsx) | 0 (build check) | 3 min |
| **Total** | **36** | **12** | **~34 tests** | **~44 min** |

### Prerequisite shadcn components to install (before starting tasks)

Check which are missing, then install in one batch:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console
bunx shadcn@latest add card scroll-area dropdown-menu avatar breadcrumb separator sheet skeleton sonner
```

### Execution order

Tasks are numbered in dependency order. Execute sequentially: 1 → 2 → 3 → 4 → 5 → 6 → 7 → 8 → 9 → 10 → 11.

- Tasks 5, 6, 7 could run in parallel (independent layout components) but their tests depend on shadcn components from the prerequisite install.
- Task 8 depends on 6 + 7 (imports Sidebar + Header).
- Task 9 is needed by Task 10 (EmptyState used in placeholder routes).
- Task 10 depends on 8 (the `_dashboard.tsx` layout route must exist).
- Task 11 is the final integration check.
