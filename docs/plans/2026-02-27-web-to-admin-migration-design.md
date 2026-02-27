# Web to Admin Migration Design

**Date:** 2026-02-27
**Author:** Claude Code
**Status:** Draft

## Overview

Migrate all management dashboard functionality from `@web/` (Astro SSR) to `@admin/` (TanStack Start) with TanStack Query v5, TanStack Router, and Paraglide for internationalization.

## Goals

1. Replace `@web/` management dashboard with `@admin/`
2. Use TanStack Start best practices (TanStack Query, TanStack Router)
3. Migrate to Paraglide for i18n
4. Rewrite tests in Vitest + TanStack Query style
5. Keep `@web/` for public-facing pages only

## Architecture

### Before Migration

```
@web/ (Astro SSR)
├── pages/          # Dashboard + Auth + Public
├── components/     # React Islands
├── lib/           # API clients (raw fetch)
├── stores/        # Zustand
├── i18n/          # i18next
└── vitest/        # Tests
```

### After Migration

```
@web/ (Astro SSR)
├── pages/         # Public only (index, blog, etc.)
└── (removed: components, lib, stores, i18n, tests)

@admin/ (TanStack Start - Feature-Based)
├── src/
│   ├── features/              # Feature modules
│   │   ├── auth/
│   │   │   ├── components/    # LoginForm, TwoFactorForm, etc.
│   │   │   ├── hooks/         # useLogin, useLogout, useMe
│   │   │   ├── types/         # LoginRequest, User types
│   │   │   └── index.ts       # Barrel export
│   │   ├── posts/
│   │   │   ├── components/    # PostsTable, PostEditor
│   │   │   ├── hooks/         # usePosts, useCreatePost
│   │   │   └── index.ts
│   │   ├── categories/
│   │   ├── tags/
│   │   ├── media/
│   │   ├── users/
│   │   ├── roles/
│   │   ├── sites/
│   │   ├── settings/
│   │   ├── comments/
│   │   ├── menus/
│   │   ├── redirects/
│   │   ├── api-keys/
│   │   └── audit/
│   ├── shared/                # Shared utilities
│   │   ├── components/       # UI components, providers
│   │   ├── hooks/            # Shared hooks
│   │   └── lib/              # API client base
│   ├── routes/               # TanStack Router (file-based)
│   └── integrations/         # TanStack Query, etc.
├── messages/                  # Paraglide i18n messages
└── tests/                    # Test utilities
```

### Feature Module Structure

Each feature follows this pattern:

```
features/{feature-name}/
├── components/     # Feature-specific React components
│   ├── FeatureList.tsx
│   └── FeatureForm.tsx
├── hooks/          # TanStack Query hooks
│   ├── useFeatureList.ts
│   └── useCreateFeature.ts
├── types/          # TypeScript types
│   └── index.ts
└── index.ts        # Barrel exports
```

### Key Benefits

1. **Colocation** - Components, hooks, and types live together
2. **TanStack Query Integration** - Query/mutation hooks in same dir as components
3. **Easy Testing** - Test file next to implementation
4. **Clear Dependencies** - No cross-feature spaghetti

## Migration Layers (Feature-Based Order)

### Layer 1: Shared Infrastructure

Files to migrate:
- `web/src/components/providers/` → `admin/src/shared/components/providers/`
  - ConsoleProvider.tsx
  - ErrorBoundary.tsx
  - SuspenseBoundary.tsx
  - QueryProvider.tsx
  - ThemeProvider.tsx

Files to create:
- `admin/src/shared/lib/` - Base API client
- `admin/src/shared/hooks/` - Shared hooks

### Layer 2: Auth Feature

```
admin/src/features/auth/
├── components/    # LoginForm, TwoFactorForm, ForgotPasswordForm
├── hooks/         # useLogin, useLogout, useMe
├── types/         # LoginRequest, User types
└── index.ts       # Barrel export
```

Migrate:
- `web/src/components/auth/` → `admin/src/features/auth/components/`
- `web/src/lib/auth-api.ts` → `admin/src/features/auth/hooks/`
- `web/src/stores/auth-store.ts` → `admin/src/features/auth/types/`

### Layer 3: Content Features (Posts, Categories, Tags, Media)

Each content feature follows the same pattern:

```
admin/src/features/posts/
├── components/    # PostsTable, PostEditor
├── hooks/         # usePosts, useCreatePost
├── types/         # Post, ListParams
└── index.ts       # Barrel export
```

Migrate:
- `web/src/components/content/` → Split into `admin/src/features/{posts,categories,tags,media}/components/`
- `web/src/lib/content-api.ts` → Split into `admin/src/features/{posts,categories,tags,media}/hooks/`

### Layer 4: System Features

Same pattern:

```
admin/src/features/users/
├── components/    # UsersTable, UserFormDialog
├── hooks/         # useUsers, useCreateUser
├── types/
└── index.ts
```

Migrate:
- `web/src/components/system/` → Split into `admin/src/features/{users,roles,sites,settings,comments,menus,redirects,api-keys,audit}/components/`
- `web/src/lib/system-api.ts` → Split into corresponding `admin/src/features/*/hooks/`

### Layer 5: Routes (TanStack Router)

| Feature | Route |
|---------|-------|
| Auth | `routes/login.tsx`, `routes/forgot-password.tsx`, `routes/reset-password.tsx` |
| Dashboard | `routes/dashboard.index.tsx` |
| Posts | `routes/dashboard.posts.index.tsx`, `routes/dashboard.posts.$id.tsx` |
| Categories | `routes/dashboard.categories.*.tsx` |
| Tags | `routes/dashboard.tags.*.tsx` |
| Media | `routes/dashboard.media.*.tsx` |
| Users | `routes/dashboard.users.*.tsx` |
| Roles | `routes/dashboard.roles.*.tsx` |
| Sites | `routes/dashboard.sites.*.tsx` |
| Settings | `routes/dashboard.settings.*.tsx` |
| Comments | `routes/dashboard.comments.*.tsx` |
| Menus | `routes/dashboard.menus.*.tsx` |
| Redirects | `routes/dashboard.redirects.*.tsx` |
| API Keys | `routes/dashboard.api-keys.*.tsx` |
| Audit | `routes/dashboard.audit.*.tsx` |

### Layer 6: Internationalization (i18n → Paraglide)

Before (i18next):
```typescript
// web/src/i18n/locales/en.json
{ "nav": { "dashboard": "Dashboard" } }
```

After (Paraglide):
```typescript
// admin/messages/en.json
{ "nav_dashboard": "Dashboard" }
```

Migrate:
- `web/src/i18n/locales/en.json` → `admin/messages/en.json`
- `web/src/i18n/locales/zh-CN.json` → `admin/messages/zh-CN.json`
- Create `admin/src/features/shared/components/LanguageSwitcher.tsx`

### Layer 6: Tests

Rewrite in Vitest + TanStack Query style:

```typescript
// Before (web)
import { render, screen, waitFor } from '@testing-library/react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

test('loads posts', async () => {
  server.use(mockGetPosts([]));
  render(<PostsList />, { wrapper: queryWrapper });
  await waitFor(() => expect(screen.getByText('No posts')).toBeInTheDocument());
});

// After (admin)
import { renderHook, waitFor } from '@tanstack/react-query';
import { usePosts } from '$lib/posts';

test('usePosts returns data', async () => {
  const { result } = renderHook(() => usePosts('test-site', {}));
  await waitFor(() => expect(result.current.isSuccess).toBe(true));
});
```

## Migration Order

```
1. Providers + Config
   ↓
2. Base API Client + TanStack Query Setup
   ↓
3. Auth API Hooks + Auth Components + Login Route
   ↓
4. Content API Hooks + Content Components + Content Routes
   ↓
5. System API Hooks + System Components + System Routes
   ↓
6. Internationalization (i18n → Paraglide)
   ↓
7. Tests (rewrite all)
   ↓
8. Cleanup (remove @web/ dashboard code)
```

## Files to Delete After Migration

After successful migration, delete from `@web/`:
- `web/src/components/auth/`
- `web/src/components/content/`
- `web/src/components/dashboard/`
- `web/src/components/layout/`
- `web/src/components/providers/`
- `web/src/components/shared/`
- `web/src/components/system/`
- `web/src/lib/auth-api.ts`
- `web/src/lib/content-api.ts`
- `web/src/lib/dashboard-api.ts`
- `web/src/lib/system-api.ts`
- `web/src/lib/query-client.ts`
- `web/src/stores/`
- `web/src/i18n/`
- `web/src/pages/dashboard/`
- `web/src/pages/login.astro`
- `web/src/pages/forgot-password/`
- `web/src/pages/reset-password.astro`
- `web/src/pages/setup/`

Keep in `@web/`:
- `web/src/pages/index.astro` (public home)
- `web/src/pages/posts/[slug].astro` (public post view - if exists)
- `web/src/lib/api.ts` (if needed for public pages)

## Success Criteria

- [ ] All dashboard functionality works in @admin/
- [ ] Login/2FA works with TanStack Router
- [ ] All API calls use TanStack Query hooks
- [ ] i18n works with Paraglide
- [ ] All tests pass in @admin/
- [ ] @web/ only contains public pages
- [ ] No console errors in production build

## Risk Mitigation

| Risk | Mitigation |
|------|------------|
| TanStack Router learning curve | Follow official TanStack Start patterns |
| Paraglide migration complexity | Use CLI to convert existing JSON |
| Large code changes | Use git worktree for isolation |
| Test coverage gaps | Write tests before each migration layer |
