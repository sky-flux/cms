# Console SPA Migration Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Migrate console/ from TanStack Start SSR to pure client-side Vite SPA, preserving all existing feature modules.

**Architecture:** Remove SSR infrastructure (~5 files), keep all 90+ feature module files unchanged. New entry: src/main.tsx → TanStack Router → Feature components.

**Tech Stack:** React 19, TanStack Router v1, TanStack Query v5, Vite, Tailwind V4, Paraglide, Vitest

---

## Context & Background

The `console/` directory is currently scaffolded as a **TanStack Start** SSR project. The spec (`docs/superpowers/specs/2026-03-19-project-redesign-design.md`) mandates migrating it to a **pure client-side Vite SPA** that produces a static `dist/` folder for `go:embed`.

### Files to DELETE or REWRITE (SSR shell)

| File | Action | Reason |
|------|--------|--------|
| `console/src/routes/__root.tsx` | **Rewrite** | Remove `shellComponent`, `HeadContent`, `Scripts`, `createRootRouteWithContext` |
| `console/src/router.tsx` | **Rewrite** | Remove `getContext()` SSR wiring |
| `console/src/integrations/tanstack-query/root-provider.tsx` | **Rewrite** | Remove SSR context pattern |
| `console/src/integrations/tanstack-query/devtools.tsx` | **Rewrite** | Switch from `ReactQueryDevtoolsPanel` to `ReactQueryDevtools` |
| `console/vite.config.ts` | **Rewrite** | Remove `tanstackStart()` plugin |
| `console/package.json` | **Update** | Remove `@tanstack/react-start`, `@tanstack/react-router-ssr-query` |

### Files to CREATE (SPA entry)

| File | Action | Reason |
|------|--------|--------|
| `console/index.html` | **Create** | Vite SPA HTML entry point |
| `console/src/main.tsx` | **Create** | React DOM render entry, replaces TanStack Start server entry |
| `console/src/queryClient.ts` | **Create** | Singleton QueryClient for SPA (not SSR context) |

### Files to PRESERVE UNCHANGED (90+ files)

All files in `console/src/features/` — hooks, components, types, tests. These have zero SSR dependencies and work identically in SPA mode.

---

## Task 1: Remove TanStack Start Dependencies

**Estimated time:** 3 minutes
**Risk:** Low — package.json edit only, no TypeScript changes

### Steps

- [ ] Open `console/package.json`
- [ ] Remove from `dependencies`:
  - `@tanstack/react-start` (entire entry)
  - `@tanstack/react-router-ssr-query` (entire entry)
- [ ] Keep all other dependencies unchanged (especially `@tanstack/react-router`, `@tanstack/router-plugin`, `@tanstack/react-query`, `@tanstack/react-query-devtools`, `@tanstack/react-router-devtools`, `@tanstack/react-devtools`, `@tanstack/devtools-vite`)

### Exact diff for `console/package.json` dependencies section

```json
// REMOVE these two lines:
"@tanstack/react-router-ssr-query": "^1.163.2",
"@tanstack/react-start": "^1.163.2",
```

Final `dependencies` block (relevant entries, all others unchanged):
```json
{
  "dependencies": {
    "@faker-js/faker": "^10.3.0",
    "@hookform/resolvers": "^5.2.2",
    "@posthog/react": "^1.8.1",
    "@radix-ui/react-dialog": "^1.1.15",
    "@t3-oss/env-core": "^0.13.10",
    "@tailwindcss/vite": "^4.2.1",
    "@tanstack/match-sorter-utils": "^8.19.4",
    "@tanstack/react-devtools": "^0.9.6",
    "@tanstack/react-form": "^1.28.3",
    "@tanstack/react-query": "^5.90.21",
    "@tanstack/react-query-devtools": "^5.91.3",
    "@tanstack/react-router": "^1.163.2",
    "@tanstack/react-router-devtools": "^1.163.2",
    "@tanstack/react-store": "^0.9.1",
    "@tanstack/react-table": "^8.21.3",
    "@tanstack/router-plugin": "^1.163.2",
    "@tanstack/store": "^0.9.1",
    "class-variance-authority": "^0.7.1",
    "clsx": "^2.1.1",
    "lucide-react": "^0.575.0",
    "posthog-js": "^1.356.0",
    "radix-ui": "^1.4.3",
    "react": "^19.2.4",
    "react-dom": "^19.2.4",
    "react-dropzone": "^15.0.0",
    "react-hook-form": "^7.71.2",
    "tailwind-merge": "^3.5.0",
    "tailwindcss": "^4.2.1",
    "tw-animate-css": "^1.4.0",
    "zod": "^4.3.6"
  }
}
```

- [ ] Run `bun install` to update `bun.lock`

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun install
```

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: FAILS (still has tanstackStart() in vite.config.ts) — this is the RED step
# The error proves the dependency was actually providing tanstackStart
```

---

## Task 2: Rewrite vite.config.ts

**Estimated time:** 3 minutes
**Risk:** Low — remove one plugin import/call, keep all others

### Current file analysis

```typescript
// console/vite.config.ts (CURRENT — SSR)
import { tanstackStart } from '@tanstack/react-start/plugin/vite'  // REMOVE
// tanstackStart(),  // REMOVE from plugins array
```

### Target file: `console/vite.config.ts`

```typescript
import { defineConfig } from 'vite'
import { devtools } from '@tanstack/devtools-vite'
import tsconfigPaths from 'vite-tsconfig-paths'
import { paraglideVitePlugin } from '@inlang/paraglide-js'
import { TanStackRouterVite } from '@tanstack/router-plugin/vite'
import viteReact from '@vitejs/plugin-react'
import tailwindcss from '@tailwindcss/vite'

const config = defineConfig({
  plugins: [
    devtools(),
    paraglideVitePlugin({
      project: './project.inlang',
      outdir: './src/paraglide',
      strategy: ['url', 'baseLocale'],
    }),
    tsconfigPaths({ projects: ['./tsconfig.json'] }),
    tailwindcss(),
    TanStackRouterVite({ routesDirectory: './src/routes', generatedRouteTree: './src/routeTree.gen.ts' }),
    viteReact(),
  ],
})

export default config
```

**Key changes:**
1. Remove `import { tanstackStart } from '@tanstack/react-start/plugin/vite'`
2. Add `import { TanStackRouterVite } from '@tanstack/router-plugin/vite'` — this is the SPA-mode file-router plugin (already in `devDependencies` as `@tanstack/router-plugin`)
3. Remove `tanstackStart()` from plugins array
4. Add `TanStackRouterVite(...)` in its place — regenerates `routeTree.gen.ts` on each dev/build run

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: FAILS — missing src/main.tsx and index.html (still RED, progress made)
```

---

## Task 3: Create src/queryClient.ts

**Estimated time:** 2 minutes
**Risk:** None — new file, no existing file touched

### Why a separate file?

In SSR mode, `QueryClient` lived in the SSR context singleton (`integrations/tanstack-query/root-provider.tsx → getContext()`). In SPA mode, a module-level singleton is the standard pattern. Extracting it to its own file allows both `src/router.tsx` and `src/main.tsx` to import the same instance without circular dependencies.

### Create `console/src/queryClient.ts`

```typescript
import { QueryClient } from '@tanstack/react-query'

export const queryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes — matches existing QueryProvider default
      retry: 3,
      refetchOnWindowFocus: false,
    },
    mutations: {
      retry: 0,
    },
  },
})
```

**Notes:**
- `staleTime: 5min` and `retry: 3` mirror the existing `src/features/shared/components/QueryProvider.tsx` defaults — ensures no behavioral regression
- `refetchOnWindowFocus: false` is standard CMS admin UX (prevents unexpected refetches while user edits)
- This singleton is imported by `src/router.tsx` (for router context) and `src/integrations/tanstack-query/root-provider.tsx` (for `QueryClientProvider`)

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: FAILS — still missing main.tsx / index.html (still RED)
```

---

## Task 4: Rewrite src/router.tsx

**Estimated time:** 3 minutes
**Risk:** Low — remove SSR context import, keep all router options

### Current file analysis

```typescript
// CURRENT: imports getContext() SSR singleton
import { getContext } from './integrations/tanstack-query/root-provider'

export function getRouter() {
  const router = createTanStackRouter({
    routeTree,
    context: getContext(),  // SSR: passes queryClient via context
    ...
  })
}
```

### Target file: `console/src/router.tsx`

```typescript
import { createRouter as createTanStackRouter } from '@tanstack/react-router'
import { routeTree } from './routeTree.gen'
import { queryClient } from './queryClient'

export function createRouter() {
  const router = createTanStackRouter({
    routeTree,
    context: {
      queryClient,
    },
    scrollRestoration: true,
    defaultPreload: 'intent',
    defaultPreloadStaleTime: 0,
  })

  return router
}

declare module '@tanstack/react-router' {
  interface Register {
    router: ReturnType<typeof createRouter>
  }
}
```

**Key changes:**
1. Import `queryClient` from `./queryClient` instead of `getContext()` from SSR provider
2. Rename `getRouter` → `createRouter` (conventional SPA naming; `getRouter` implied lazy singleton for SSR)
3. Pass `{ queryClient }` directly as `context` — matches the `MyRouterContext` interface that `__root.tsx` will declare

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: FAILS — __root.tsx still has SSR-specific APIs
```

---

## Task 5: Rewrite src/routes/__root.tsx

**Estimated time:** 5 minutes
**Risk:** Medium — most complex SSR-to-SPA transformation; remove `shellComponent`, `HeadContent`, `Scripts`, `head()`

### Current file analysis

The current `__root.tsx` uses:
- `createRootRouteWithContext<MyRouterContext>()` — keep (still valid in SPA with context)
- `shellComponent: RootDocument` — **REMOVE** (SSR-only: renders `<html>`, `<head>`, `<body>`)
- `head: () => ({ meta, links })` — **REMOVE** (SSR-only: injects `<HeadContent />`)
- `HeadContent` — **REMOVE** (SSR import from `@tanstack/react-router`)
- `Scripts` — **REMOVE** (SSR import from `@tanstack/react-router`)
- `appCss` imported as `?url` — **REMOVE** (CSS is imported directly in `main.tsx`)
- `PostHogProvider` — **KEEP** (analytics, works in SPA)
- `TanStackQueryProvider` — **REPLACE** with import from updated root-provider
- `THEME_INIT_SCRIPT` — **MOVE** to `index.html` as inline `<script>` in `<head>`

### Target file: `console/src/routes/__root.tsx`

```typescript
import { Outlet, createRootRouteWithContext } from '@tanstack/react-router'
import { TanStackRouterDevtoolsPanel } from '@tanstack/react-router-devtools'
import { TanStackDevtools } from '@tanstack/react-devtools'

import PostHogProvider from '../integrations/posthog/provider'
import TanStackQueryProvider from '../integrations/tanstack-query/root-provider'
import TanStackQueryDevtools from '../integrations/tanstack-query/devtools'
import StoreDevtools from '../lib/demo-store-devtools'
import { getLocale } from '#/paraglide/runtime'

import type { QueryClient } from '@tanstack/react-query'

interface MyRouterContext {
  queryClient: QueryClient
}

export const Route = createRootRouteWithContext<MyRouterContext>()({
  beforeLoad: async () => {
    if (typeof document !== 'undefined') {
      document.documentElement.setAttribute('lang', getLocale())
    }
  },
  component: RootComponent,
})

function RootComponent() {
  return (
    <PostHogProvider>
      <TanStackQueryProvider>
        <Outlet />
        <TanStackDevtools
          config={{
            position: 'bottom-right',
          }}
          plugins={[
            {
              name: 'Tanstack Router',
              render: <TanStackRouterDevtoolsPanel />,
            },
            StoreDevtools,
            TanStackQueryDevtools,
          ]}
        />
      </TanStackQueryProvider>
    </PostHogProvider>
  )
}
```

**Key changes:**
1. `createRootRouteWithContext<MyRouterContext>()` — unchanged, SPA still supports router context
2. `shellComponent: RootDocument` → `component: RootComponent` — SPA uses `component`, not `shellComponent`
3. `RootComponent` renders `<Outlet />` only (no `<html>/<head>/<body>`) — the HTML shell is in `index.html`
4. Remove `HeadContent`, `Scripts` imports (SSR-only exports from `@tanstack/react-router`)
5. Remove `head()` — meta tags go in `index.html`
6. Remove `appCss?url` import — CSS is imported in `main.tsx`
7. `THEME_INIT_SCRIPT` moves to `index.html` (see Task 6)
8. Keep `beforeLoad` for Paraglide locale detection (still valid in SPA)
9. Keep all devtools plugins — they work identically in SPA mode

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: FAILS — still missing main.tsx / index.html
```

---

## Task 6: Create index.html + src/main.tsx

**Estimated time:** 4 minutes
**Risk:** Low — standard Vite SPA boilerplate

### 6a. Create `console/index.html`

This is the Vite SPA HTML entry point. Vite serves this file in dev mode and uses it as the template for `dist/index.html` in production. The `go:embed` directive in `embed.go` embeds all of `console/dist/`, including this generated `index.html`.

```html
<!DOCTYPE html>
<html lang="en" suppressHydrationWarning>
  <head>
    <meta charset="utf-8" />
    <meta name="viewport" content="width=device-width, initial-scale=1" />
    <title>Sky Flux CMS</title>
    <script>
      (function(){try{var stored=window.localStorage.getItem('theme');var mode=(stored==='light'||stored==='dark'||stored==='auto')?stored:'auto';var prefersDark=window.matchMedia('(prefers-color-scheme: dark)').matches;var resolved=mode==='auto'?(prefersDark?'dark':'light'):mode;var root=document.documentElement;root.classList.remove('light','dark');root.classList.add(resolved);if(mode==='auto'){root.removeAttribute('data-theme')}else{root.setAttribute('data-theme',mode)}root.style.colorScheme=resolved;}catch(e){}})();
    </script>
  </head>
  <body class="font-sans antialiased [overflow-wrap:anywhere] selection:bg-[rgba(79,184,178,0.24)]">
    <div id="root"></div>
    <script type="module" src="/src/main.tsx"></script>
  </body>
</html>
```

**Notes:**
- `THEME_INIT_SCRIPT` is inlined here (moved from `__root.tsx` `RootDocument`). Keeping it in `<head>` prevents flash of wrong theme.
- `suppressHydrationWarning` on `<html>` — prevents React warnings when theme class is applied before hydration (even though this is SPA, not SSR, it's harmless and future-safe)
- `<div id="root">` is the React mount target
- `<script type="module" src="/src/main.tsx">` — Vite resolves this from project root

### 6b. Create `console/src/main.tsx`

```typescript
import { StrictMode } from 'react'
import { createRoot } from 'react-dom/client'
import { RouterProvider } from '@tanstack/react-router'
import { createRouter } from './router'
import './styles.css'

const router = createRouter()

const rootElement = document.getElementById('root')
if (!rootElement) {
  throw new Error('Root element #root not found in index.html')
}

createRoot(rootElement).render(
  <StrictMode>
    <RouterProvider router={router} />
  </StrictMode>,
)
```

**Notes:**
- `import './styles.css'` — replaces the `appCss?url` import that was in the old `__root.tsx` head. Vite handles CSS injection automatically in SPA mode.
- `createRouter()` (not `getRouter()`) — matches the renamed export in Task 4
- `RouterProvider` from `@tanstack/react-router` — the SPA entry point for TanStack Router
- `StrictMode` — best practice, enables double-invoke checks in dev
- Guard on `rootElement` — fails fast with clear message if `index.html` is missing the `#root` div

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: May fail at __root.tsx (imports HeadContent/Scripts if not yet done)
# If Task 5 is complete: should PASS or fail only at root-provider.tsx
```

---

## Task 7: Update Integrations

**Estimated time:** 4 minutes
**Risk:** Low — simplification only (remove SSR patterns, keep functionality)

### 7a. Rewrite `console/src/integrations/tanstack-query/root-provider.tsx`

**Current:** Uses SSR-safe lazy singleton via `getContext()` — necessary in SSR to avoid sharing state across requests. In SPA, this pattern is unnecessary overhead.

**Target:**

```typescript
import type { ReactNode } from 'react'
import { QueryClientProvider } from '@tanstack/react-query'
import { queryClient } from '../../queryClient'

export default function TanStackQueryProvider({
  children,
}: {
  children: ReactNode
}) {
  return (
    <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>
  )
}
```

**Key changes:**
1. Remove `getContext()` export entirely — no longer needed (was SSR context hook)
2. Import `queryClient` from `../../queryClient` — the module singleton created in Task 3
3. `QueryClientProvider` uses the same `queryClient` instance as `src/router.tsx` — this is required so router's `loader` functions share the same cache as component hooks

**Important:** The existing `src/features/shared/components/QueryProvider.tsx` creates its own `QueryClient` instance with a `new QueryClient()`. That's fine for isolated feature tests — those tests wrap components in their own provider. The `root-provider.tsx` is only used by `__root.tsx` (application shell), so there is no conflict.

### 7b. Rewrite `console/src/integrations/tanstack-query/devtools.tsx`

**Current:** Exports `ReactQueryDevtoolsPanel` (embedded panel, for use inside `@tanstack/react-devtools` composite panel).

**Target:** Keep identical — this is already correct for SPA mode. No change needed.

```typescript
// No changes required — already correct
import { ReactQueryDevtoolsPanel } from '@tanstack/react-query-devtools'

export default {
  name: 'Tanstack Query',
  render: <ReactQueryDevtoolsPanel />,
}
```

### 7c. Verify `console/src/integrations/posthog/provider.tsx`

PostHog provider has no SSR dependencies. Open file and confirm it uses standard React context. No changes expected.

### Verification

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Expected: GREEN — build succeeds, dist/ is generated
```

If build succeeds, also verify the output:
```bash
ls /Users/martinadamsdev/workspace/sky-flux-cms/console/dist/
# Expected: index.html  assets/  (Vite SPA output structure)
```

---

## Task 8: Run All Existing Tests

**Estimated time:** 3 minutes
**Risk:** Low — feature module tests are isolated from SSR infrastructure

### Run full test suite

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run test
```

### Expected outcome

All existing tests in `src/features/` pass unchanged:

| Test file | Expected | Reason |
|-----------|----------|--------|
| `src/features/shared/__tests__/api-client.test.ts` | PASS | No SSR dependency |
| `src/features/shared/components/__tests__/ErrorBoundary.test.tsx` | PASS | Pure React component |
| `src/features/shared/components/__tests__/ThemeProvider.test.tsx` | PASS | Pure React component |
| `src/features/shared/components/__tests__/QueryProvider.test.tsx` | PASS | Uses own `new QueryClient()` — isolated |
| `src/features/shared/components/__tests__/ConsoleProvider.test.tsx` | PASS | Composed of above |
| `src/features/auth/__tests__/auth-hooks.test.ts` | PASS | Hooks only, no routing |
| `src/features/auth/components/__tests__/LoginForm.test.tsx` | PASS | RTL render, no SSR |
| `src/features/posts/__tests__/posts-hooks.test.ts` | PASS | Hooks only |
| `src/features/posts/components/__tests__/PostsTable.test.tsx` | PASS | RTL render |
| `src/features/categories/__tests__/categories-hooks.test.ts` | PASS | Hooks only |
| `src/features/categories/components/__tests__/CategoryTree.test.tsx` | PASS | RTL render |
| `src/features/tags/__tests__/tags-hooks.test.ts` | PASS | Hooks only |
| `src/features/media/__tests__/media-hooks.test.ts` | PASS | Hooks only |
| `src/features/users/__tests__/users-hooks.test.ts` | PASS | Hooks only |
| `src/features/roles/__tests__/roles-hooks.test.ts` | PASS | Hooks only |
| `src/features/sites/__tests__/sites-hooks.test.ts` | PASS | Hooks only |

### If any test fails

**Pattern A — import resolves to removed SSR package:**
```
Error: Cannot find module '@tanstack/react-start'
```
Fix: Search for any remaining `@tanstack/react-start` imports:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && grep -r "react-start\|react-router-ssr" src/
```

**Pattern B — TypeScript error on `getContext` import:**
```
Error: Module '"../integrations/tanstack-query/root-provider"' has no exported member 'getContext'
```
Fix: `src/router.tsx` still imports `getContext` — verify Task 4 rewrite was applied fully.

**Pattern C — `#/paraglide/runtime` not found:**
```
Error: Cannot find module '#/paraglide/runtime'
```
Fix: Paraglide outdir generation may need a dev/build run first:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build
# Then re-run tests
bun run test
```

### Final verification

```bash
# Full build + test in sequence
cd /Users/martinadamsdev/workspace/sky-flux-cms/console && bun run build && bun run test
```

Both commands must exit with code 0.

---

## Commit Strategy

Each task is a standalone commit using gitmoji format (no Co-Authored-By):

```bash
# Task 1
git add console/package.json console/bun.lock
git commit -m "🗑️ remove TanStack Start SSR dependencies from console"

# Task 2
git add console/vite.config.ts
git commit -m "⚡️ switch console Vite config from tanstackStart to TanStackRouterVite"

# Task 3
git add console/src/queryClient.ts
git commit -m "✨ add console QueryClient singleton for SPA mode"

# Task 4
git add console/src/router.tsx
git commit -m "♻️ rewrite console router to pure SPA (remove SSR context)"

# Task 5
git add console/src/routes/__root.tsx
git commit -m "♻️ rewrite __root route: remove shellComponent/HeadContent/Scripts"

# Task 6
git add console/index.html console/src/main.tsx
git commit -m "✨ add SPA entry point: index.html + src/main.tsx"

# Task 7
git add console/src/integrations/
git commit -m "♻️ simplify tanstack-query root-provider: remove SSR getContext pattern"

# Task 8 (only if any test fixes were needed)
git add <fixed files>
git commit -m "🐛 fix import paths after SSR → SPA migration"
```

---

## Architecture After Migration

```
console/
├── index.html                          # NEW: Vite SPA entry (theme init script here)
├── vite.config.ts                      # CHANGED: TanStackRouterVite replaces tanstackStart
├── package.json                        # CHANGED: -react-start, -ssr-query
├── src/
│   ├── main.tsx                        # NEW: createRoot() + RouterProvider
│   ├── queryClient.ts                  # NEW: singleton QueryClient
│   ├── router.tsx                      # CHANGED: imports queryClient directly
│   ├── styles.css                      # UNCHANGED (imported by main.tsx now)
│   ├── routes/
│   │   ├── __root.tsx                  # CHANGED: component= not shellComponent=
│   │   ├── index.tsx                   # UNCHANGED
│   │   └── about.tsx                   # UNCHANGED
│   ├── integrations/
│   │   ├── tanstack-query/
│   │   │   ├── root-provider.tsx       # CHANGED: imports singleton queryClient
│   │   │   └── devtools.tsx            # UNCHANGED
│   │   └── posthog/
│   │       └── provider.tsx            # UNCHANGED
│   └── features/                       # ENTIRELY UNCHANGED (90+ files)
│       ├── auth/
│       ├── posts/
│       ├── categories/
│       ├── tags/
│       ├── media/
│       ├── users/
│       ├── roles/
│       ├── sites/
│       └── shared/
└── dist/                               # BUILD OUTPUT: go:embed target
    ├── index.html
    └── assets/
        ├── index-[hash].js
        └── index-[hash].css
```

### Data Flow (SPA)

```
Browser loads /console/*
  └── Go serves console/dist/index.html (go:embed)
       └── <script src="/src/main.tsx"> (Vite-bundled as assets/index-[hash].js)
            └── createRoot('#root').render(<RouterProvider router={createRouter()} />)
                 └── TanStack Router matches URL → renders route component
                      └── __root.tsx: PostHogProvider → TanStackQueryProvider → <Outlet />
                           └── Route component (e.g. routes/index.tsx)
                                └── Feature components (src/features/*/components/)
                                     └── Feature hooks (src/features/*/hooks/)
                                          └── API client (src/features/shared/api-client.ts)
                                               └── fetch('/api/v1/admin/...')
```

### QueryClient Sharing (Critical)

```
src/queryClient.ts (singleton)
├── imported by src/router.tsx → context.queryClient
│    └── available in route loaders via router.context.queryClient.prefetchQuery(...)
└── imported by src/integrations/tanstack-query/root-provider.tsx
     └── <QueryClientProvider client={queryClient}> wraps entire app
          └── all useQuery/useMutation hooks share this cache
```

This ensures router prefetches and component fetches hit the **same cache** — the canonical SPA pattern.

---

## Rollback Plan

If the migration causes unexpected breakage, revert is trivial:

```bash
git revert HEAD~7..HEAD   # reverts all 7 task commits
# or per-task:
git revert <commit-hash>
```

All feature module files are untouched, so rollback has zero risk of losing business logic.
