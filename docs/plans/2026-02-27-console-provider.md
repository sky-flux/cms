# ConsoleProvider Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Create a unified provider composition component that combines React Query, i18n, theme, error handling, and toast notifications for all dashboard pages.

**Architecture:**
- `ConsoleProvider`: Main component wrapping all providers in correct order
- `ErrorBoundary`: Class component catching render errors
- `SuspenseBoundary`: Functional component handling Suspense states
- `config/providers.ts`: Environment-aware configuration
- Provider nesting: ErrorBoundary → QueryProvider → I18nProvider → ThemeProvider → SuspenseBoundary → Toaster

**Tech Stack:** React 19, TypeScript, React Query v5, react-i18next, sonner, Vitest, @testing-library/react

---

## Task 1: Create Provider Configuration

**Files:**
- Create: `web/src/config/providers.ts`
- Test: `web/src/config/__tests__/providers.config.test.ts`

**Step 1: Write the failing test**

Create `web/src/config/__tests__/providers.config.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { providerConfig } from '../providers';

describe('providerConfig', () => {
  beforeEach(() => {
    // Reset import.meta.env before each test
    vi.unstubAllGlobals();
  });

  it('should use development config when DEV mode', () => {
    vi.stubGlobal('import.meta', { env: { DEV: true } });

    // Dynamic import to get fresh config
    vi.resetModules();
    const { providerConfig: devConfig } = require('../providers');

    expect(devConfig.queryClient.logger).toBe(true);
    expect(devConfig.queryClient.retries).toBe(0);
    expect(devConfig.queryClient.staleTime).toBe(1000 * 60);
    expect(devConfig.theme.defaultTheme).toBe('system');
  });

  it('should use production config when PROD mode', () => {
    vi.stubGlobal('import.meta', { env: { DEV: false } });

    vi.resetModules();
    const { providerConfig: prodConfig } = require('../providers');

    expect(prodConfig.queryClient.logger).toBe(false);
    expect(prodConfig.queryClient.retries).toBe(3);
    expect(prodConfig.queryClient.retryDelay).toBe(1000);
    expect(prodConfig.queryClient.staleTime).toBe(1000 * 60 * 5);
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web && bun run vitest run src/config/__tests__/providers.config.test.ts`

Expected: FAIL with "Cannot find module '../providers'"

**Step 3: Write minimal implementation**

Create `web/src/config/providers.ts`:

```typescript
interface ProviderConfig {
  queryClient: {
    staleTime: number;
    cacheTime: number;
    retries: number;
    retryDelay: number;
    logger: boolean;
  };
  theme: {
    defaultTheme: 'light' | 'dark' | 'system';
  };
}

const devConfig: ProviderConfig = {
  queryClient: {
    staleTime: 1000 * 60, // 1 minute
    cacheTime: 1000 * 60 * 5, // 5 minutes
    retries: 0,
    retryDelay: 0,
    logger: true,
  },
  theme: {
    defaultTheme: 'system',
  },
};

const prodConfig: ProviderConfig = {
  queryClient: {
    staleTime: 1000 * 60 * 5, // 5 minutes
    cacheTime: 1000 * 60 * 30, // 30 minutes
    retries: 3,
    retryDelay: 1000,
    logger: false,
  },
  theme: {
    defaultTheme: 'system',
  },
};

export const providerConfig: ProviderConfig = import.meta.env.DEV ? devConfig : prodConfig;
```

**Step 4: Run test to verify it passes**

Run: `cd web && bun run vitest run src/config/__tests__/providers.config.test.ts`

Expected: PASS (2 tests)

**Step 5: Commit**

```bash
git add web/src/config/providers.ts web/src/config/__tests__/providers.config.test.ts
git commit -m "feat: add environment-aware provider configuration

- Development: detailed logging, no retries, shorter cache
- Production: minimal logging, 3 retries, longer cache
- Config includes QueryClient and theme settings"
```

---

## Task 2: Create ErrorBoundary Component

**Files:**
- Create: `web/src/components/providers/ErrorBoundary.tsx`
- Create: `web/src/components/providers/__tests__/ErrorBoundary.test.tsx`

**Step 1: Write the failing test**

Create `web/src/components/providers/__tests__/ErrorBoundary.test.tsx`:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ErrorBoundary } from '../ErrorBoundary';

// Create a component that throws an error
const ThrowError = ({ message = 'Test error' }: { message?: string }) => {
  throw new Error(message);
};

describe('ErrorBoundary', () => {
  it('should catch errors and display fallback UI', () => {
    render(
      <ErrorBoundary>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByText(/something went wrong/i)).toBeInTheDocument();
    expect(screen.getByText(/test error/i)).toBeInTheDocument();
  });

  it('should call onError callback when error occurs', () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError}>
        <ThrowError message="Callback test" />
      </ErrorBoundary>
    );

    expect(onError).toHaveBeenCalledWith(
      expect.any(Error),
      expect.objectContaining({
        componentStack: expect.any(String),
      })
    );
  });

  it('should render custom fallback when provided', () => {
    const customFallback = <div>Custom Error UI</div>;

    render(
      <ErrorBoundary fallback={customFallback}>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByText('Custom Error UI')).toBeInTheDocument();
  });

  it('should not show error details when showErrorDetails is false', () => {
    render(
      <ErrorBoundary showErrorDetails={false}>
        <ThrowError message="Secret error" />
      </ErrorBoundary>
    );

    expect(screen.queryByText('Secret error')).not.toBeInTheDocument();
    expect(screen.getByText(/unexpected error occurred/i)).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web && bun run vitest run src/components/providers/__tests__/ErrorBoundary.test.tsx`

Expected: FAIL with "Cannot find module '../ErrorBoundary'"

**Step 3: Write minimal implementation**

Create `web/src/components/providers/ErrorBoundary.tsx`:

```typescript
import { Component, type ReactNode, type ErrorInfo } from 'react';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  showErrorDetails?: boolean;
}

interface ErrorBoundaryState {
  hasError: boolean;
  error?: Error;
}

export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo): void {
    const { onError } = this.props;
    if (onError) {
      onError(error, errorInfo);
    }
  }

  handleReset = (): void => {
    this.setState({ hasError: false, error: undefined });
  };

  render(): ReactNode {
    const { hasError, error } = this.state;
    const { children, fallback, showErrorDetails = true } = this.props;

    if (hasError) {
      if (fallback) {
        return fallback;
      }

      return (
        <div className="flex min-h-[400px] flex-col items-center justify-center p-6">
          <AlertCircle className="h-12 w-12 text-destructive" />
          <h2 className="mt-4 text-lg font-semibold">Something went wrong</h2>
          <p className="mt-2 text-sm text-muted-foreground">
            {showErrorDetails && error ? error.message : 'An unexpected error occurred'}
          </p>

          {showErrorDetails && error && (
            <details className="mt-4 text-xs text-muted-foreground">
              <summary>Error details</summary>
              <pre className="mt-2 overflow-auto bg-muted p-2">
                {error.stack}
              </pre>
            </details>
          )}

          <div className="mt-4 flex gap-2">
            <Button onClick={() => window.location.reload()}>
              Reload Page
            </Button>
            <Button variant="outline" onClick={this.handleReset}>
              Try Again
            </Button>
          </div>
        </div>
      );
    }

    return children;
  }
}
```

**Step 4: Run test to verify it passes**

Run: `cd web && bun run vitest run src/components/providers/__tests__/ErrorBoundary.test.tsx`

Expected: PASS (4 tests)

**Step 5: Commit**

```bash
git add web/src/components/providers/ErrorBoundary.tsx web/src/components/providers/__tests__/ErrorBoundary.test.tsx
git commit -m "feat: add ErrorBoundary component

- Catches JavaScript errors in component tree
- Displays user-friendly error UI
- Supports custom fallback and error callback
- Optional error details display (dev mode)"
```

---

## Task 3: Create SuspenseBoundary Component

**Files:**
- Create: `web/src/components/providers/SuspenseBoundary.tsx`
- Create: `web/src/components/providers/__tests__/SuspenseBoundary.test.tsx`

**Step 1: Write the failing test**

Create `web/src/components/providers/__tests__/SuspenseBoundary.test.tsx`:

```typescript
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SuspenseBoundary } from '../SuspenseBoundary';

describe('SuspenseBoundary', () => {
  it('should show fallback while children are suspending', () => {
    // Create a component that never resolves (always suspends)
    const SuspenseComponent = () => {
      throw new Promise(() => {}); // eslint-disable-line no-throw-literal
    };

    render(
      <SuspenseBoundary>
        <SuspenseComponent />
      </SuspenseBoundary>
    );

    expect(screen.getByRole('status')).toBeInTheDocument();
    expect(screen.getByText(/loading/i)).toBeInTheDocument();
  });

  it('should render custom fallback when provided', () => {
    const SuspenseComponent = () => {
      throw new Promise(() => {}); // eslint-disable-line no-throw-literal
    };

    const customFallback = <div>Custom Loading...</div>;

    render(
      <SuspenseBoundary fallback={customFallback}>
        <SuspenseComponent />
      </SuspenseBoundary>
    );

    expect(screen.getByText('Custom Loading...')).toBeInTheDocument();
  });

  it('should render children when no suspension occurs', () => {
    const NormalComponent = () => <div>Normal Content</div>;

    render(
      <SuspenseBoundary>
        <NormalComponent />
      </SuspenseBoundary>
    );

    expect(screen.getByText('Normal Content')).toBeInTheDocument();
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web && bun run vitest run src/components/providers/__tests__/SuspenseBoundary.test.tsx`

Expected: FAIL with "Cannot find module '../SuspenseBoundary'"

**Step 3: Write minimal implementation**

Create `web/src/components/providers/SuspenseBoundary.tsx`:

```typescript
import { Suspense, type ReactNode } from 'react';
import { Loader2 } from 'lucide-react';

interface SuspenseBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

const defaultFallback = (
  <div className="flex min-h-[400px] items-center justify-center">
    <div className="flex flex-col items-center gap-4">
      <Loader2 className="h-8 w-8 animate-spin text-muted-foreground" />
      <p className="text-sm text-muted-foreground">Loading...</p>
    </div>
  </div>
);

export function SuspenseBoundary({
  children,
  fallback = defaultFallback
}: SuspenseBoundaryProps): ReactNode {
  return <Suspense fallback={fallback}>{children}</Suspense>;
}
```

**Step 4: Run test to verify it passes**

Run: `cd web && bun run vitest run src/components/providers/__tests__/SuspenseBoundary.test.tsx`

Expected: PASS (3 tests)

**Step 5: Commit**

```bash
git add web/src/components/providers/SuspenseBoundary.tsx web/src/components/providers/__tests__/SuspenseBoundary.test.tsx
git commit -m "feat: add SuspenseBoundary component

- Wraps React Suspense with consistent loading UI
- Customizable fallback component
- Default loading spinner with 'Loading...' text"
```

---

## Task 4: Create ConsoleProvider Component

**Files:**
- Create: `web/src/components/providers/ConsoleProvider.tsx`
- Modify: `web/src/components/providers/ConsoleProvider.tsx` (already created as placeholder)
- Test: `web/src/components/providers/__tests__/ConsoleProvider.test.tsx`

**Step 1: Write the failing test**

Create `web/src/components/providers/__tests__/ConsoleProvider.test.tsx`:

```typescript
import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { QueryClient, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ConsoleProvider } from '../ConsoleProvider';

// Mock ResizeObserver for jsdom
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
} as unknown as typeof ResizeObserver;

describe('ConsoleProvider', () => {
  afterEach(() => {
    cleanup();
    // Reset document classes
    document.documentElement.classList.remove('dark');
  });

  it('should provide all contexts to children', () => {
    const TestComponent = () => {
      const { t } = useTranslation();
      const queryClient = useQueryClient();

      return (
        <div>
          <span data-testid="i18n">{t('common.loading')}</span>
          <span data-testid="query-client">
            {queryClient instanceof QueryClient ? 'yes' : 'no'}
          </span>
        </div>
      );
    };

    render(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    expect(screen.getByTestId('i18n')).toHaveTextContent('Loading...');
    expect(screen.getByTestId('query-client')).toHaveTextContent('yes');
  });

  it('should render Toaster', () => {
    render(
      <ConsoleProvider>
        <div>Children</div>
      </ConsoleProvider>
    );

    // Sonner Toaster renders with role="status"
    const toasters = screen.getAllByRole('status');
    expect(toasters.length).toBeGreaterThan(0);
  });

  it('should apply light theme when defaultTheme="light"', () => {
    render(
      <ConsoleProvider defaultTheme="light">
        <div>Test</div>
      </ConsoleProvider>
    );

    // ThemeProvider adds 'dark' class only for dark theme
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('should apply dark theme when defaultTheme="dark"', () => {
    render(
      <ConsoleProvider defaultTheme="dark">
        <div>Test</div>
      </ConsoleProvider>
    );

    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('should not recreate QueryClient on re-renders', () => {
    let renderCount = 0;

    const TestComponent = () => {
      const queryClient = useQueryClient();
      renderCount++;
      return <div data-testid="client">{queryClient.toString()}</div>;
    };

    const { rerender } = render(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    const firstClient = screen.getByTestId('client').textContent;
    const firstRenderCount = renderCount;

    rerender(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    const secondClient = screen.getByTestId('client').textContent;

    expect(firstClient).toBe(secondClient);
    expect(renderCount).toBe(firstRenderCount + 1);
  });
});
```

**Step 2: Run test to verify it fails**

Run: `cd web && bun run vitest run src/components/providers/__tests__/ConsoleProvider.test.tsx`

Expected: FAIL - ConsoleProvider returns null

**Step 3: Write minimal implementation**

Replace content of `web/src/components/providers/ConsoleProvider.tsx`:

```typescript
import { useState, type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { QueryProvider } from './QueryProvider';
import { I18nProvider } from './I18nProvider';
import { ThemeProvider } from './ThemeProvider';
import { ErrorBoundary } from './ErrorBoundary';
import { SuspenseBoundary } from './SuspenseBoundary';

export interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
  showErrorDetails?: boolean;
}

/**
 * ConsoleProvider - Unified provider composition for dashboard/console pages
 *
 * Combines all necessary providers in the optimal order:
 * 1. ErrorBoundary - Catches all render errors
 * 2. QueryProvider - React Query for data fetching
 * 3. I18nProvider - Internationalization
 * 4. ThemeProvider - Dark/light theme management
 * 5. SuspenseBoundary - Unified loading states
 * 6. Toaster - Toast notifications
 *
 * Usage: Only for /dashboard pages, not for auth/setup pages.
 */
export function ConsoleProvider({
  children,
  defaultTheme = 'system',
  showErrorDetails = import.meta.env.DEV
}: ConsoleProviderProps): ReactNode {
  return (
    <ErrorBoundary showErrorDetails={showErrorDetails}>
      <QueryProvider>
        <I18nProvider>
          <ThemeProvider defaultTheme={defaultTheme}>
            <SuspenseBoundary>
              {children}
              <Toaster />
            </SuspenseBoundary>
          </ThemeProvider>
        </I18nProvider>
      </QueryProvider>
    </ErrorBoundary>
  );
}
```

**Step 4: Run test to verify it passes**

Run: `cd web && bun run vitest run src/components/providers/__tests__/ConsoleProvider.test.tsx`

Expected: PASS (5 tests)

**Step 5: Commit**

```bash
git add web/src/components/providers/ConsoleProvider.tsx web/src/components/providers/__tests__/ConsoleProvider.test.tsx
git commit -m "feat: add ConsoleProvider component

- Unified provider composition for dashboard pages
- ErrorBoundary → QueryProvider → I18nProvider → ThemeProvider → SuspenseBoundary → Toaster
- Environment-aware showErrorDetails (default: DEV=true)
- Configurable defaultTheme prop (default: 'system')"
```

---

## Task 5: Add i18n Error Messages

**Files:**
- Modify: `web/src/i18n/locales/en.json`
- Modify: `web/src/i18n/locales/zh-CN.json`

**Step 1: Add English error messages**

Edit `web/src/i18n/locales/en.json`, add to `errors` section:

```json
"errors": {
  "fetchFailed": "Failed to fetch data",
  "networkError": "Network error occurred",
  "unknownError": "An unknown error occurred",
  "somethingWentWrong": "Something went wrong",
  "tryAgain": "Try Again",
  "reloadPage": "Reload Page"
}
```

**Step 2: Add Chinese error messages**

Edit `web/src/i18n/locales/zh-CN.json`, add to `errors` section:

```json
"errors": {
  "fetchFailed": "获取数据失败",
  "networkError": "网络错误",
  "unknownError": "未知错误",
  "somethingWentWrong": "出现错误",
  "tryAgain": "重试",
  "reloadPage": "重新加载页面"
}
```

**Step 3: Verify i18n files are valid**

Run: `cd web && bun run astro check`

Expected: No errors

**Step 4: Commit**

```bash
git add web/src/i18n/locales/en.json web/src/i18n/locales/zh-CN.json
git commit -m "feat(i18n): add error messages for ErrorBoundary

- English: fetchFailed, networkError, unknownError, etc.
- Chinese: 对应的中文错误消息"
```

---

## Task 6: Migrate DashboardPage as Example

**Files:**
- Modify: `web/src/components/dashboard/DashboardPage.tsx`

**Step 1: Replace Provider wrappers with ConsoleProvider**

Edit `web/src/components/dashboard/DashboardPage.tsx`, replace:

```typescript
export function DashboardPage() {
  return (
    <QueryProvider>
      <DashboardPageInner />
    </QueryProvider>
  );
}
```

With:

```typescript
import { ConsoleProvider } from '@/components/providers/ConsoleProvider';

export function DashboardPage() {
  return (
    <ConsoleProvider>
      <DashboardPageInner />
    </ConsoleProvider>
  );
}
```

Also remove the `QueryProvider` import if no longer needed.

**Step 2: Run tests to verify**

Run: `cd web && bun run vitest run src/components/dashboard/__tests__/DashboardPage.test.tsx`

Expected: PASS (all tests still work)

**Step 3: Manual verification in browser**

Run: `cd web && bun run dev`

Visit: `http://localhost:4321/dashboard`

Expected:
- Dashboard loads correctly
- Statistics display
- No console errors
- Theme toggle works

**Step 4: Commit**

```bash
git add web/src/components/dashboard/DashboardPage.tsx
git commit -m "refactor(dashboard): migrate to ConsoleProvider

- Replace QueryProvider with ConsoleProvider
- Dashboard serves as migration example for other pages"
```

---

## Task 7: Migrate Remaining Dashboard Pages

**Files to modify:**
- `web/src/components/content/CategoriesPage.tsx`
- `web/src/components/content/PostsPage.tsx`
- `web/src/components/content/PostEditorPage.tsx`
- `web/src/components/content/RevisionsPage.tsx`
- `web/src/components/system/SitesPage.tsx`
- `web/src/components/system/SettingsPage.tsx`
- `web/src/components/system/UsersPage.tsx`
- `web/src/components/system/RolesPage.tsx`
- `web/src/components/system/CommentsPage.tsx`
- `web/src/components/system/AuditPage.tsx`
- `web/src/components/system/MenusPage.tsx`
- `web/src/components/system/RedirectsPage.tsx`
- And any other pages in `/dashboard` route

**Step 1: Migrate CategoriesPage**

Edit `web/src/components/content/CategoriesPage.tsx`:

Find the export function (around line 178):

```typescript
export function CategoriesPage() {
  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <CategoriesPageInner />
        <Toaster />
      </QueryClientProvider>
    </I18nextProvider>
  );
}
```

Replace with:

```typescript
import { ConsoleProvider } from '@/components/providers/ConsoleProvider';

export function CategoriesPage() {
  return (
    <ConsoleProvider>
      <CategoriesPageInner />
    </ConsoleProvider>
  );
}
```

Remove imports:
- `import { QueryClientProvider } from '@tanstack/react-query';`
- `import { I18nextProvider } from 'react-i18next';`
- `import { Toaster } from 'sonner';`
- `import i18n from '@/i18n/config';`
- `import { queryClient } from '@/lib/query-client';`

**Step 2: Test CategoriesPage**

Run: `cd web && bun run vitest run src/components/content/__tests__/CategoriesPage.test.tsx`

Expected: PASS

**Step 3: Repeat for all other dashboard pages**

For each page:
1. Replace provider wrappers with `<ConsoleProvider>`
2. Remove unused imports
3. Run tests
4. Commit in batches

**Example batch commit:**

```bash
git add web/src/components/content/PostsPage.tsx web/src/components/content/PostEditorPage.tsx web/src/components/content/RevisionsPage.tsx
git commit -m "refactor(content): migrate post pages to ConsoleProvider

- PostsPage, PostEditorPage, RevisionsPage now use ConsoleProvider
- Remove duplicate QueryClientProvider/I18nextProvider/Toaster imports"
```

**Step 4: Final commit for remaining pages**

```bash
git add web/src/components/system/*.tsx
git commit -m "refactor(system): migrate system pages to ConsoleProvider

- All system pages (Sites, Settings, Users, Roles, Comments, Audit, Menus, Redirects)
- Unified provider management via ConsoleProvider"
```

**Step 5: Run full test suite**

Run: `cd web && bun run vitest run`

Expected: ALL TESTS PASS

**Step 6: Manual smoke test**

Run: `cd web && bun run dev`

Visit all dashboard pages and verify:
- `/dashboard` - loads
- `/dashboard/posts` - loads
- `/dashboard/categories` - loads
- `/dashboard/tags` - loads
- `/dashboard/media` - loads
- `/dashboard/users` - loads
- `/dashboard/settings` - loads
- `/dashboard/sites` - loads
- `/dashboard/comments` - loads
- `/dashboard/menus` - loads
- `/dashboard/redirects` - loads
- `/dashboard/audit` - loads

**Step 7: Final migration commit**

```bash
git add -A
git commit -m "feat: complete ConsoleProvider migration

- All dashboard pages now use ConsoleProvider
- Removed ~200 lines of duplicate provider wrapper code
- Consistent error handling and loading states across all pages"
```

---

## Task 8: Update Documentation

**Files:**
- Modify: `MEMORY.md`
- Create: `web/src/components/providers/README.md`

**Step 1: Update MEMORY.md**

Add to MEMORY.md in "Frontend" section:

```markdown
### Provider Architecture

- **ConsoleProvider**: Unified provider composition for /dashboard pages
  - Wraps: ErrorBoundary → QueryProvider → I18nProvider → ThemeProvider → SuspenseBoundary → Toaster
  - Location: `web/src/components/providers/ConsoleProvider.tsx`
  - Usage: All dashboard page components
  - Config: `web/src/config/providers.ts` (environment-aware)
```

**Step 2: Create Provider README**

Create `web/src/components/providers/README.md`:

```markdown
# Providers

This directory contains React provider components for the Sky Flux CMS dashboard.

## ConsoleProvider

Unified provider composition for all `/dashboard` pages.

### Usage

\`\`\`typescript
import { ConsoleProvider } from '@/components/providers/ConsoleProvider';

export function MyDashboardPage() {
  return (
    <ConsoleProvider>
      <MyDashboardPageInner />
    </ConsoleProvider>
  );
}
\`\`\`

### Provider Stack (outer → inner)

1. **ErrorBoundary**: Catches JavaScript errors
2. **QueryProvider**: React Query for data fetching
3. **I18nProvider**: Internationalization
4. **ThemeProvider**: Dark/light theme
5. **SuspenseBoundary**: Loading states
6. **Toaster**: Toast notifications

### Props

- `defaultTheme?: 'light' | 'dark' | 'system'` - Default theme (default: 'system')
- `showErrorDetails?: boolean` - Show error stack traces (default: true in DEV)

## Individual Providers

- **QueryProvider**: `QueryProvider.tsx` - React Query wrapper
- **I18nProvider**: `I18nProvider.tsx` - react-i18next wrapper
- **ThemeProvider**: `ThemeProvider.tsx` - Theme management
- **ErrorBoundary**: `ErrorBoundary.tsx` - Error catching
- **SuspenseBoundary**: `SuspenseBoundary.tsx` - Suspense loading states

## Configuration

See `web/src/config/providers.ts` for environment-aware provider settings.
```

**Step 3: Commit documentation**

```bash
git add MEMORY.md web/src/components/providers/README.md
git commit -m "docs: add ConsoleProvider documentation

- Update MEMORY.md with provider architecture
- Create providers README with usage examples
- Document provider stack and configuration"
```

---

## Verification Checklist

After completing all tasks, verify:

- [ ] All tests pass: `bun run vitest run`
- [ ] No TypeScript errors: `bun run astro check`
- [ ] No linting errors: `bun run biome check .`
- [ ] All dashboard pages load in browser
- [ ] Error boundary catches errors (test with manual error throw)
- [ ] Toast notifications work
- [ ] Theme toggle works
- [ ] i18n translations work
- [ ] No console errors in browser
- [ ] Code reduced by ~200 lines (duplicate provider wrappers removed)

---

**Total Estimated Time:** 2-3 hours

**Files Created:** 10
**Files Modified:** 20+
**Tests Added:** 15+

**Next Steps:**
1. Run `bun run vitest run` to verify all tests pass
2. Run `bun run dev` and manually test all dashboard pages
3. Monitor for any regressions in error handling or loading states
4. Consider migrating auth/setup pages if needed (not covered in this plan)
