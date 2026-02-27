# ConsoleProvider Design Document

**Date:** 2026-02-27
**Author:** Claude Code
**Status:** Approved

## Overview

`ConsoleProvider` is a unified provider composition component for all dashboard/console pages (`/dashboard/*` routes). It combines React Query, i18n, theme management, error handling, and toast notifications into a single reusable component.

**Goal:** Eliminate repetitive provider wrapper code and ensure consistent user experience across all dashboard pages.

## Architecture

### Provider Nesting Order

```
ConsoleProvider
├── ErrorBoundary          [1] 最外层 - 捕获所有内部错误
├── QueryProvider          [2] 数据获取基础设施
│   └── QueryClient 缓存 (useState 避免重复创建)
├── I18nProvider           [3] 国际化上下文
├── ThemeProvider         [4] 主题管理
├── SuspenseBoundary       [5] 统一加载状态
└── Toaster                [6] Toast 通知
    └── {children}         [7] 业务组件
```

### Why This Order?

1. **ErrorBoundary (outermost)**: Catches all possible errors, including provider initialization failures
2. **QueryProvider**: Infrastructure layer, doesn't depend on other contexts
3. **I18nProvider**: Toaster may need translated error messages
4. **ThemeProvider**: Global setting, no dependencies on other contexts
5. **SuspenseBoundary**: Catches lazy-loaded components and data loading states
6. **Toaster (innermost)**: Ensures access to all contexts (especially i18n)

## Components

### 1. ConsoleProvider

**Purpose:** Main component that combines all providers

**TypeScript Interface:**
```typescript
interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
  showErrorDetails?: boolean;
}

export function ConsoleProvider({
  children,
  defaultTheme = 'system',
  showErrorDetails = import.meta.env.DEV
}: ConsoleProviderProps): JSX.Element
```

**Usage Example:**
```typescript
export function DashboardPage() {
  return (
    <ConsoleProvider>
      <DashboardPageInner />
    </ConsoleProvider>
  );
}
```

### 2. ErrorBoundary (New Component)

**Purpose:** Catch JavaScript errors in component tree

**TypeScript Interface:**
```typescript
interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
  onError?: (error: Error, errorInfo: ErrorInfo) => void;
  showErrorDetails?: boolean;
}

export class ErrorBoundary extends Component<
  ErrorBoundaryProps,
  ErrorBoundaryState
>
```

**Fallback UI:**
- Error icon
- "Something went wrong" message
- Error details (dev mode only)
- "Reload Page" and "Try Again" buttons

### 3. SuspenseBoundary (New Component)

**Purpose:** Catch React Suspense states (lazy loading, data fetching)

**TypeScript Interface:**
```typescript
interface SuspenseBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

export function SuspenseBoundary({
  children,
  fallback
}: SuspenseBoundaryProps): JSX.Element
```

**Default Loading UI:**
- Spinner icon
- "Loading..." text

### 4. Configuration File

**Location:** `web/src/config/providers.ts`

**Structure:**
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

export const providerConfig: ProviderConfig =
  import.meta.env.DEV ? devConfig : prodConfig;
```

**Environment Differences:**
- **Development**: Shorter cache times, no retries, detailed logging
- **Production**: Longer cache times, 3 retries, minimal logging

## Data Flow

### Normal Flow
```
User visits /dashboard/categories
    ↓
CategoriesPage renders
    ↓
ConsoleProvider initializes all contexts
    ↓
CategoriesPageInner uses useQuery to fetch data
    ↓
Show loading state → Show data
```

### Error Handling

**Scenario 1: Component Render Error (Sync)**
```
CategoriesPageInner throws exception
    ↓
ErrorBoundary catches
    ↓
Show error UI + "Reload Page" button
```

**Scenario 2: API Request Failure (Async)**
```
useQuery throws ApiError
    ↓
Component handles: toast.error(t('errors.fetchFailed'))
    ↓
Show error message, component stays usable
```

**Scenario 3: Network Error**
```
React Query detects network error
    ↓
Auto retry (3x prod, 0x dev)
    ↓
Still fails → Show error UI
```

**Scenario 4: Lazy Load Failure**
```
React.lazy fails to load component
    ↓
SuspenseBoundary's ErrorBoundary catches
    ↓
Show "Failed to load component" + retry button
```

## Testing Strategy

### Unit Tests

**ErrorBoundary.test.tsx** (90%+ coverage)
- Error catching and fallback UI display
- onError callback invocation
- Error state reset

**SuspenseBoundary.test.tsx** (80%+ coverage)
- Suspense state fallback rendering
- Children rendering after resolution

**ConsoleProvider.test.tsx** (85%+ coverage)
- All contexts provided correctly
- Toaster rendered
- Custom defaultTheme applied

**providers.config.test.ts** (100% coverage)
- Development config activation
- Production config activation

### Integration Tests

Test real page components using ConsoleProvider:
- Loading state
- Error state
- Success state with data display

### Performance Tests

Verify QueryClient is not recreated on re-renders (useState caching).

## Edge Cases

| Scenario | Handling |
|----------|----------|
| ErrorBoundary itself fails | Use `componentDidCatch` to show basic error info |
| QueryClient init fails | ErrorBoundary catches, shows "Failed to initialize" |
| i18n loading fails | Fallback to displaying key itself (i18next default) |
| Theme switching fails | Log warning but don't break app, keep current theme |

## Migration Plan

### Phase 1: Implementation
1. Create `ErrorBoundary` component
2. Create `SuspenseBoundary` component
3. Create `config/providers.ts`
4. Implement `ConsoleProvider` component
5. Write all tests

### Phase 2: Migration (One page as example)
1. Migrate `DashboardPage` to use `ConsoleProvider`
2. Verify functionality
3. Document migration pattern

### Phase 3: Bulk Migration
1. Migrate all dashboard pages to use `ConsoleProvider`
2. Remove old Provider wrapper code
3. Update documentation

## Files to Create

1. `web/src/components/providers/ConsoleProvider.tsx` (main component)
2. `web/src/components/providers/ErrorBoundary.tsx`
3. `web/src/components/providers/SuspenseBoundary.tsx`
4. `web/src/config/providers.ts`
5. `web/src/components/providers/__tests__/ConsoleProvider.test.tsx`
6. `web/src/components/providers/__tests__/ErrorBoundary.test.tsx`
7. `web/src/components/providers/__tests__/SuspenseBoundary.test.tsx`
8. `web/src/config/__tests__/providers.config.test.ts`

## Files to Modify

1. All dashboard page components (replace Provider wrappers with `<ConsoleProvider>`)
2. `web/src/i18n/locales/en.json` (add error messages)
3. `web/src/i18n/locales/zh-CN.json` (添加错误消息)

## Success Criteria

- ✅ All dashboard pages use `<ConsoleProvider>` instead of individual providers
- ✅ Error boundaries catch and display errors gracefully
- ✅ Loading states are consistent across all pages
- ✅ All tests pass (unit + integration)
- ✅ No performance degradation (QueryClient cached)
- ✅ Code reduced by ~200 lines across all page components
