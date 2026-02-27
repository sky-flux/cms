# Web to Admin Migration Implementation Plan - Feature-Based Architecture

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.
> **IMPORTANT:** This plan uses Feature-Based Architecture where each feature is a self-contained module.

**Goal:** Migrate all management dashboard functionality from Astro SSR (@web/) to TanStack Start (@admin/) with Feature-Based Architecture, TanStack Query hooks, TanStack Router, and Paraglide i18n.

**Architecture:** Feature-Based - each feature module is self-contained with its own components, hooks, types, and tests.

**Feature Module Structure:**
```
features/{feature-name}/
├── components/      # React components specific to this feature
├── hooks/          # TanStack Query hooks specific to this feature
├── types/          # TypeScript types specific to this feature
└── index.ts       # Barrel exports
```

**Tech Stack:** TanStack Start, TanStack Query v5, TanStack Router, Paraglide, Vitest

---

## Feature 1: shared - Core Infrastructure

**Description:** ConsoleProvider, ErrorBoundary, SuspenseBoundary, ThemeProvider, QueryProvider, api-client

**Target Path:** `admin/src/features/shared/`

### Task 1.1: Create api-client

**Files:**
- Create: `admin/src/features/shared/types/api-client.ts`
- Create: `admin/src/features/shared/api-client.ts`
- Create: `admin/src/features/shared/index.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/__tests__/api-client.test.ts
import { describe, it, expect, vi } from 'vitest';

describe('apiClient', () => {
  it('should export ApiError class', () => {
    expect(() => {
      const { ApiError } = require('../api-client');
      new ApiError('test', 500);
    }).not.toThrow();
  });

  it('should have get, post, put, patch, delete methods', () => {
    const { apiClient } = require('../api-client');
    expect(typeof apiClient.get).toBe('function');
    expect(typeof apiClient.post).toBe('function');
    expect(typeof apiClient.put).toBe('function');
    expect(typeof apiClient.patch).toBe('function');
    expect(typeof apiClient.delete).toBe('function');
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/__tests__/api-client.test.ts
```
Expected: FAIL (module not found)

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/types/api-client.ts
export class ApiError extends Error {
  constructor(
    message: string,
    public status: number,
    public code?: string
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

export interface RequestOptions extends RequestInit {
  params?: Record<string, string | number | boolean>;
}

export interface ListResponse<T> {
  data: T[];
  meta: {
    page: number;
    pageSize: number;
    total: number;
    totalPages: number;
  };
}
```

```typescript
// admin/src/features/shared/api-client.ts
import { ApiError, type RequestOptions } from './types/api-client';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  private buildURL(endpoint: string, params?: Record<string, string | number | boolean>): string {
    const url = new URL(`${this.baseURL}${endpoint}`);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, String(value));
      });
    }
    return url.toString();
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        errorData.message || `HTTP error ${response.status}`,
        response.status,
        errorData.code
      );
    }
    return response.json();
  }

  async get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'GET',
      credentials: 'include',
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async post<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async put<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async patch<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'PATCH',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'DELETE',
      credentials: 'include',
      ...options,
    });
    return this.handleResponse<T>(response);
  }
}

export const apiClient = new ApiClient();
export { ApiError };
```

```typescript
// admin/src/features/shared/index.ts
export { apiClient, ApiError } from './api-client';
export type { RequestOptions, ListResponse } from './types/api-client';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/__tests__/api-client.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/
git commit -m "feat(admin): add api-client to shared feature"
```

---

### Task 1.2: Create ErrorBoundary

**Files:**
- Create: `admin/src/features/shared/components/ErrorBoundary.tsx`
- Create: `admin/src/features/shared/components/__tests__/ErrorBoundary.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/components/__tests__/ErrorBoundary.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ErrorBoundary } from '../ErrorBoundary';
import { Component, type ErrorInfo } from 'react';

const ThrowError = () => {
  throw new Error('Test error');
};

describe('ErrorBoundary', () => {
  it('renders children when no error', () => {
    render(
      <ErrorBoundary>
        <div>Normal Content</div>
      </ErrorBoundary>
    );
    expect(screen.getByText('Normal Content')).toBeInTheDocument();
  });

  it('shows fallback when error occurs', () => {
    const fallback = <div data-testid="fallback">Error Occurred</div>;
    render(
      <ErrorBoundary fallback={fallback}>
        <ThrowError />
      </ErrorBoundary>
    );
    expect(screen.getByTestId('fallback')).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ErrorBoundary.test.tsx
```
Expected: FAIL (component not found)

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/components/ErrorBoundary.tsx
import { Component, type ReactNode } from 'react';

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

export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = { hasError: false };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return { hasError: true, error };
  }

  componentDidCatch(error: Error, errorInfo: ErrorInfo) {
    this.props.onError?.(error, errorInfo);
  }

  render() {
    if (this.state.hasError) {
      return this.props.fallback ?? (
        <div style={{ padding: '2rem', textAlign: 'center' }}>
          <h1>Something went wrong</h1>
          {this.props.showErrorDetails && this.state.error && (
            <pre style={{ marginTop: '1rem', textAlign: 'left' }}>
              {this.state.error.message}
            </pre>
          )}
        </div>
      );
    }
    return this.props.children;
  }
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ErrorBoundary.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/components/
git commit -m "feat(admin): add ErrorBoundary to shared feature"
```

---

### Task 1.3: Create SuspenseBoundary

**Files:**
- Create: `admin/src/features/shared/components/SuspenseBoundary.tsx`
- Create: `admin/src/features/shared/components/__tests__/SuspenseBoundary.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/components/__tests__/SuspenseBoundary.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import { SuspenseBoundary } from '../SuspenseBoundary';
import { Suspense } from 'react';

describe('SuspenseBoundary', () => {
  it('renders children when resolved', async () => {
    render(
      <SuspenseBoundary>
        <div>Loaded</div>
      </SuspenseBoundary>
    );
    await waitFor(() => {
      expect(screen.getByText('Loaded')).toBeInTheDocument();
    });
  });

  it('shows fallback during suspense', async () => {
    const Loading = () => {
      throw new Promise(() => {}); // Never resolves
    };

    render(
      <SuspenseBoundary fallback={<div>Loading...</div>}>
        <Suspense fallback={<div>Loading...</div>}>
          <Loading />
        </Suspense>
      </SuspenseBoundary>
    );

    await waitFor(() => {
      expect(screen.getByText('Loading...')).toBeInTheDocument();
    });
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/SuspenseBoundary.test.tsx
```
Expected: FAIL (component not found)

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/components/SuspenseBoundary.tsx
import { Suspense, type ReactNode } from 'react';

interface SuspenseBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

export function SuspenseBoundary({
  children,
  fallback = <div style={{ padding: '2rem', textAlign: 'center' }}>Loading...</div>
}: SuspenseBoundaryProps): ReactNode {
  return <Suspense fallback={fallback}>{children}</Suspense>;
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/SuspenseBoundary.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/components/
git commit -m "feat(admin): add SuspenseBoundary to shared feature"
```

---

### Task 1.4: Create ThemeProvider

**Files:**
- Create: `admin/src/features/shared/components/ThemeProvider.tsx`
- Create: `admin/src/features/shared/components/useTheme.ts`
- Create: `admin/src/features/shared/components/__tests__/ThemeProvider.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/components/__tests__/ThemeProvider.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ThemeProvider, useTheme } from '../ThemeProvider';

const TestComponent = () => {
  const { theme } = useTheme();
  return <div data-testid="theme">{theme}</div>;
};

describe('ThemeProvider', () => {
  it('provides theme context', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );
    expect(screen.getByTestId('theme')).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ThemeProvider.test.tsx
```
Expected: FAIL (component not found)

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/components/ThemeProvider.tsx
'use client';

import { createContext, useContext, useState, useEffect, type ReactNode } from 'react';

type Theme = 'light' | 'dark' | 'system';

interface ThemeContextValue {
  theme: Theme;
  setTheme: (theme: Theme) => void;
  resolvedTheme: 'light' | 'dark';
}

const ThemeContext = createContext<ThemeContextValue | null>(null);

export function useTheme() {
  const context = useContext(ThemeContext);
  if (!context) throw new Error('useTheme must be used within ThemeProvider');
  return context;
}

interface ThemeProviderProps {
  children: ReactNode;
  defaultTheme?: Theme;
}

export function ThemeProvider({
  children,
  defaultTheme = 'system'
}: ThemeProviderProps): ReactNode {
  const [theme, setTheme] = useState<Theme>(defaultTheme);
  const [resolvedTheme, setResolvedTheme] = useState<'light' | 'dark'>('light');

  useEffect(() => {
    const root = document.documentElement;
    root.classList.remove('light', 'dark');
    root.classList.add(resolvedTheme);
  }, [resolvedTheme]);

  useEffect(() => {
    if (theme === 'system') {
      const mediaQuery = window.matchMedia('(prefers-color-scheme: dark)');
      setResolvedTheme(mediaQuery.matches ? 'dark' : 'light');
      const handler = (e: MediaQueryListEvent) => {
        setResolvedTheme(e.matches ? 'dark' : 'light');
      };
      mediaQuery.addEventListener('change', handler);
      return () => mediaQuery.removeEventListener('change', handler);
    } else {
      setResolvedTheme(theme);
    }
  }, [theme]);

  return (
    <ThemeContext.Provider value={{ theme, setTheme, resolvedTheme }}>
      {children}
    </ThemeContext.Provider>
  );
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ThemeProvider.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/components/
git commit -m "feat(admin): add ThemeProvider to shared feature"
```

---

### Task 1.5: Create QueryProvider

**Files:**
- Create: `admin/src/features/shared/components/QueryProvider.tsx`
- Create: `admin/src/features/shared/components/__tests__/QueryProvider.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/components/__tests__/QueryProvider.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryProvider } from '../QueryProvider';
import { useQuery, QueryClient, QueryClientProvider } from '@tanstack/react-query';

const TestQuery = () => {
  const { data } = useQuery({
    queryKey: ['test'],
    queryFn: () => Promise.resolve('test-data'),
  });
  return <div data-testid="data">{data}</div>;
};

describe('QueryProvider', () => {
  it('renders with query client', () => {
    render(
      <QueryProvider>
        <TestQuery />
      </QueryProvider>
    );
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/QueryProvider.test.tsx
```
Expected: FAIL

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/components/QueryProvider.tsx
import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

export interface QueryProviderProps {
  children: ReactNode;
}

const defaultQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 3,
    },
  },
});

export function QueryProvider({ children }: QueryProviderProps): ReactNode {
  return (
    <QueryClientProvider client={defaultQueryClient}>
      {children}
    </QueryClientProvider>
  );
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/QueryProvider.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/components/
git commit -m "feat(admin): add QueryProvider to shared feature"
```

---

### Task 1.6: Create ConsoleProvider (Composition Root)

**Files:**
- Create: `admin/src/features/shared/components/ConsoleProvider.tsx`
- Create: `admin/src/features/shared/components/__tests__/ConsoleProvider.test.tsx`
- Update: `admin/src/features/shared/index.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/shared/components/__tests__/ConsoleProvider.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConsoleProvider } from '../ConsoleProvider';

describe('ConsoleProvider', () => {
  it('renders children', () => {
    render(
      <ConsoleProvider>
        <div data-testid="child">Test Content</div>
      </ConsoleProvider>
    );
    expect(screen.getByTestId('child')).toHaveTextContent('Test Content');
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ConsoleProvider.test.tsx
```
Expected: FAIL

**Step 3: Write minimal implementation**

```typescript
// admin/src/features/shared/components/ConsoleProvider.tsx
import type { ReactNode } from 'react';
import { QueryProvider } from './QueryProvider';
import { ThemeProvider } from './ThemeProvider';
import { ErrorBoundary } from './ErrorBoundary';
import { SuspenseBoundary } from './SuspenseBoundary';

export interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
  showErrorDetails?: boolean;
}

export function ConsoleProvider({
  children,
  defaultTheme = 'system',
  showErrorDetails = import.meta.env.DEV
}: ConsoleProviderProps): ReactNode {
  return (
    <ErrorBoundary showErrorDetails={showErrorDetails}>
      <QueryProvider>
        <ThemeProvider defaultTheme={defaultTheme}>
          <SuspenseBoundary>
            {children}
          </SuspenseBoundary>
        </ThemeProvider>
      </QueryProvider>
    </ErrorBoundary>
  );
}
```

```typescript
// admin/src/features/shared/index.ts (追加 export)
export { ConsoleProvider } from './components/ConsoleProvider';
export { ErrorBoundary } from './components/ErrorBoundary';
export { SuspenseBoundary } from './components/SuspenseBoundary';
export { ThemeProvider, useTheme } from './components/ThemeProvider';
export { QueryProvider } from './components/QueryProvider';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/shared/components/__tests__/ConsoleProvider.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/shared/
git commit -m "feat(admin): add ConsoleProvider composition root to shared feature"
```

---

## Feature 2: auth - Authentication

**Description:** LoginForm, TwoFactorForm, ForgotPasswordForm, useLogin, useLogout, useMe

**Target Path:** `admin/src/features/auth/`

### Task 2.1: Create auth hooks

**Files:**
- Create: `admin/src/features/auth/types/auth.ts`
- Create: `admin/src/features/auth/hooks/useLogin.ts`
- Create: `admin/src/features/auth/hooks/useLogout.ts`
- Create: `admin/src/features/auth/hooks/useMe.ts`
- Create: `admin/src/features/auth/hooks/useForgotPassword.ts`
- Create: `admin/src/features/auth/hooks/useResetPassword.ts`
- Create: `admin/src/features/auth/hooks/index.ts`
- Create: `admin/src/features/auth/__tests__/auth-hooks.test.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/auth/__tests__/auth-hooks.test.ts
import { describe, it, expect } from 'vitest';

describe('auth hooks', () => {
  it('exports useLogin', () => {
    const { useLogin } = require('../hooks');
    expect(useLogin).toBeDefined();
  });

  it('exports useLogout', () => {
    const { useLogout } = require('../hooks');
    expect(useLogout).toBeDefined();
  });

  it('exports useMe', () => {
    const { useMe } = require('../hooks');
    expect(useMe).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/auth/__tests__/auth-hooks.test.ts
```
Expected: FAIL (module not found)

**Step 3: Write implementation**

```typescript
// admin/src/features/auth/types/auth.ts
export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
  requires?: 'totp';
  tempToken?: string;
}

export interface VerifyTOTPRequest {
  tempToken: string;
  code: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface MeResponse {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  role: string;
  siteIds: string[];
}
```

```typescript
// admin/src/features/auth/hooks/useLogin.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { LoginRequest, LoginResponse } from '../types/auth';

export function useLogin() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: LoginRequest): Promise<LoginResponse> => {
      const response = await apiClient.post<LoginResponse>('/auth/login', data);
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['me'] });
    },
  });
}
```

```typescript
// admin/src/features/auth/hooks/useLogout.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useLogout() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async () => {
      await apiClient.post('/auth/logout');
    },
    onSuccess: () => {
      queryClient.clear();
    },
  });
}
```

```typescript
// admin/src/features/auth/hooks/useMe.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { MeResponse } from '../types/auth';

export function useMe() {
  return createQuery({
    queryKey: ['me'],
    queryFn: async (): Promise<MeResponse> => {
      const response = await apiClient.get<MeResponse>('/auth/me');
      return response;
    },
    staleTime: 1000 * 60 * 5, // 5 minutes
  });
}
```

```typescript
// admin/src/features/auth/hooks/useForgotPassword.ts
import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { ForgotPasswordRequest } from '../types/auth';

export function useForgotPassword() {
  return createMutation({
    mutationFn: async (data: ForgotPasswordRequest) => {
      await apiClient.post('/auth/forgot-password', data);
    },
  });
}
```

```typescript
// admin/src/features/auth/hooks/useResetPassword.ts
import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { ResetPasswordRequest } from '../types/auth';

export function useResetPassword() {
  return createMutation({
    mutationFn: async (data: ResetPasswordRequest) => {
      await apiClient.post('/auth/reset-password', data);
    },
  });
}
```

```typescript
// admin/src/features/auth/hooks/useVerifyTOTP.ts
import { createMutation } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { VerifyTOTPRequest, LoginResponse } from '../types/auth';

export function useVerifyTOTP() {
  return createMutation({
    mutationFn: async (data: VerifyTOTPRequest): Promise<LoginResponse> => {
      const response = await apiClient.post<LoginResponse>('/auth/verify-totp', data);
      return response;
    },
  });
}
```

```typescript
// admin/src/features/auth/hooks/index.ts
export { useLogin } from './useLogin';
export { useLogout } from './useLogout';
export { useMe } from './useMe';
export { useForgotPassword } from './useForgotPassword';
export { useResetPassword } from './useResetPassword';
export { useVerifyTOTP } from './useVerifyTOTP';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/auth/__tests__/auth-hooks.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/auth/
git commit -m "feat(admin): add auth hooks to auth feature"
```

---

### Task 2.2: Create LoginForm component

**Files:**
- Create: `admin/src/features/auth/components/LoginForm.tsx`
- Create: `admin/src/features/auth/components/__tests__/LoginForm.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/auth/components/__tests__/LoginForm.test.tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { LoginForm } from '../LoginForm';

describe('LoginForm', () => {
  it('renders email and password fields', () => {
    render(<LoginForm />);
    expect(screen.getByLabelText(/email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/password/i)).toBeInTheDocument();
  });

  it('shows submit button', () => {
    render(<LoginForm />);
    expect(screen.getByRole('button', { name: /sign in/i })).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/auth/components/__tests__/LoginForm.test.tsx
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/auth/components/LoginForm.tsx
'use client';

import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useLogin } from '../hooks';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Form, FormControl, FormField, FormItem, FormLabel, FormMessage } from '@/components/ui/form';

const loginSchema = z.object({
  email: z.string().email('Invalid email address'),
  password: z.string().min(1, 'Password is required'),
});

type LoginFormData = z.infer<typeof loginSchema>;

interface LoginFormProps {
  onSuccess?: () => void;
}

export function LoginForm({ onSuccess }: LoginFormProps) {
  const [showPassword, setShowPassword] = useState(false);
  const login = useLogin();

  const form = useForm<LoginFormData>({
    resolver: zodResolver(loginSchema),
    defaultValues: {
      email: '',
      password: '',
    },
  });

  const onSubmit = async (data: LoginFormData) => {
    try {
      const result = await login.mutateAsync(data);
      if (result.requires === 'totp') {
        // Handle 2FA redirect
        window.location.href = `/login/2fa?tempToken=${result.tempToken}`;
      } else if (onSuccess) {
        onSuccess();
      }
    } catch (error) {
      // Error handling done by React Query
    }
  };

  return (
    <Form {...form}>
      <form onSubmit={form.handleSubmit(onSubmit)} className="space-y-4">
        <FormField
          control={form.control}
          name="email"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Email</FormLabel>
              <FormControl>
                <Input type="email" placeholder="you@example.com" {...field} />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <FormField
          control={form.control}
          name="password"
          render={({ field }) => (
            <FormItem>
              <FormLabel>Password</FormLabel>
              <FormControl>
                <Input
                  type={showPassword ? 'text' : 'password'}
                  placeholder="Enter your password"
                  {...field}
                />
              </FormControl>
              <FormMessage />
            </FormItem>
          )}
        />
        <Button type="submit" className="w-full" disabled={login.isPending}>
          {login.isPending ? 'Signing in...' : 'Sign In'}
        </Button>
      </form>
    </Form>
  );
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/auth/components/__tests__/LoginForm.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/auth/components/
git commit -m "feat(admin): add LoginForm to auth feature"
```

---

## Feature 3: posts - Posts Management

**Description:** PostsTable, PostEditor, usePosts, useCreatePost

**Target Path:** `admin/src/features/posts/`

### Task 3.1: Create posts hooks

**Files:**
- Create: `admin/src/features/posts/types/posts.ts`
- Create: `admin/src/features/posts/hooks/usePosts.ts`
- Create: `admin/src/features/posts/hooks/usePost.ts`
- Create: `admin/src/features/posts/hooks/useCreatePost.ts`
- Create: `admin/src/features/posts/hooks/useUpdatePost.ts`
- Create: `admin/src/features/posts/hooks/useDeletePost.ts`
- Create: `admin/src/features/posts/hooks/usePublishPost.ts`
- Create: `admin/src/features/posts/hooks/index.ts`
- Create: `admin/src/features/posts/__tests__/posts-hooks.test.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/posts/__tests__/posts-hooks.test.ts
import { describe, it, expect } from 'vitest';

describe('posts hooks', () => {
  it('exports usePosts', () => {
    const { usePosts } = require('../hooks');
    expect(usePosts).toBeDefined();
  });

  it('exports useCreatePost', () => {
    const { useCreatePost } = require('../hooks');
    expect(useCreatePost).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/posts/__tests__/posts-hooks.test.ts
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/posts/types/posts.ts
export interface Post {
  id: string;
  title: string;
  slug: string;
  content: string;
  excerpt?: string;
  status: 'draft' | 'published' | 'scheduled' | 'private';
  authorId: string;
  siteId: string;
  publishedAt?: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreatePostRequest {
  title: string;
  slug: string;
  content?: string;
  excerpt?: string;
  status?: 'draft' | 'published';
  categoryId?: string;
  tagIds?: string[];
}

export interface UpdatePostRequest extends Partial<CreatePostRequest> {
  publishedAt?: string;
}

export interface ListParams {
  page?: number;
  pageSize?: number;
  status?: string;
  categoryId?: string;
  tagIds?: string[];
  authorId?: string;
  search?: string;
}
```

```typescript
// admin/src/features/posts/hooks/usePosts.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Post, ListParams } from '../types/posts';

export function usePosts(siteSlug: string, params: ListParams = {}) {
  return createQuery({
    queryKey: ['posts', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Post>>(
        `/sites/${siteSlug}/posts`,
        { params }
      );
      return response;
    },
  });
}
```

```typescript
// admin/src/features/posts/hooks/usePost.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post } from '../types/posts';

export function usePost(siteSlug: string, postId: string) {
  return createQuery({
    queryKey: ['post', siteSlug, postId],
    queryFn: async () => {
      const response = await apiClient.get<Post>(`/sites/${siteSlug}/posts/${postId}`);
      return response;
    },
    enabled: !!postId,
  });
}
```

```typescript
// admin/src/features/posts/hooks/useCreatePost.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post, CreatePostRequest } from '../types/posts';

export function useCreatePost() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; post: CreatePostRequest }): Promise<Post> => {
      const response = await apiClient.post<Post>(
        `/sites/${data.siteSlug}/posts`,
        data.post
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['posts', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/posts/hooks/useUpdatePost.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post, UpdatePostRequest } from '../types/posts';

export function useUpdatePost() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; postId: string; post: UpdatePostRequest }): Promise<Post> => {
      const response = await apiClient.patch<Post>(
        `/sites/${data.siteSlug}/posts/${data.postId}`,
        data.post
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['posts', variables.siteSlug] });
      queryClient.invalidateQueries({ queryKey: ['post', variables.siteSlug, variables.postId] });
    },
  });
}
```

```typescript
// admin/src/features/posts/hooks/useDeletePost.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeletePost() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; postId: string }) => {
      await apiClient.delete(`/sites/${data.siteSlug}/posts/${data.postId}`);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['posts', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/posts/hooks/usePublishPost.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Post } from '../types/posts';

export function usePublishPost() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; postId: string; status: string }): Promise<Post> => {
      const response = await apiClient.post<Post>(
        `/sites/${data.siteSlug}/posts/${data.postId}/status`,
        { status: data.status }
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['posts', variables.siteSlug] });
      queryClient.invalidateQueries({ queryKey: ['post', variables.siteSlug, variables.postId] });
    },
  });
}
```

```typescript
// admin/src/features/posts/hooks/index.ts
export { usePosts } from './usePosts';
export { usePost } from './usePost';
export { useCreatePost } from './useCreatePost';
export { useUpdatePost } from './useUpdatePost';
export { useDeletePost } from './useDeletePost';
export { usePublishPost } from './usePublishPost';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/posts/__tests__/posts-hooks.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/posts/
git commit -m "feat(admin): add posts hooks to posts feature"
```

---

### Task 3.2: Create PostsTable component

**Files:**
- Create: `admin/src/features/posts/components/PostsTable.tsx`
- Create: `admin/src/features/posts/components/__tests__/PostsTable.test.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/posts/components/__tests__/PostsTable.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { PostsTable } from '../PostsTable';
import type { Post } from '../../types/posts';

const mockPosts: Post[] = [
  {
    id: '1',
    title: 'Test Post 1',
    slug: 'test-post-1',
    content: 'Content 1',
    status: 'published',
    authorId: 'author-1',
    siteId: 'site-1',
    createdAt: '2024-01-01T00:00:00Z',
    updatedAt: '2024-01-01T00:00:00Z',
  },
];

describe('PostsTable', () => {
  it('renders posts data', () => {
    render(<PostsTable posts={mockPosts} siteSlug="test-site" />);
    expect(screen.getByText('Test Post 1')).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/posts/components/__tests__/PostsTable.test.tsx
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/posts/components/PostsTable.tsx
'use client';

import { useNavigate } from '@tanstack/react-router';
import { useDeletePost } from '../hooks';
import type { Post } from '../types/posts';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { DropdownMenu, DropdownMenuContent, DropdownMenuItem, DropdownMenuTrigger } from '@/components/ui/dropdown-menu';
import { MoreHorizontal, Pencil, Trash, Eye } from 'lucide-react';

interface PostsTableProps {
  posts: Post[];
  siteSlug: string;
}

const statusColors: Record<string, string> = {
  draft: 'bg-gray-500',
  published: 'bg-green-500',
  scheduled: 'bg-yellow-500',
  private: 'bg-red-500',
};

export function PostsTable({ posts, siteSlug }: PostsTableProps) {
  const navigate = useNavigate();
  const deletePost = useDeletePost();

  const handleDelete = async (postId: string) => {
    if (confirm('Are you sure you want to delete this post?')) {
      await deletePost.mutateAsync({ siteSlug, postId });
    }
  };

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Title</TableHead>
          <TableHead>Status</TableHead>
          <TableHead>Created</TableHead>
          <TableHead>Updated</TableHead>
          <TableHead className="w-[100px]">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {posts.map((post) => (
          <TableRow key={post.id}>
            <TableCell className="font-medium">{post.title}</TableCell>
            <TableCell>
              <Badge className={statusColors[post.status]}>{post.status}</Badge>
            </TableCell>
            <TableCell>{new Date(post.createdAt).toLocaleDateString()}</TableCell>
            <TableCell>{new Date(post.updatedAt).toLocaleDateString()}</TableCell>
            <TableCell>
              <DropdownMenu>
                <DropdownMenuTrigger asChild>
                  <Button variant="ghost" className="h-8 w-8 p-0">
                    <MoreHorizontal className="h-4 w-4" />
                  </Button>
                </DropdownMenuTrigger>
                <DropdownMenuContent align="end">
                  <DropdownMenuItem onClick={() => window.open(`/preview/${siteSlug}/${post.slug}`)}>
                    <Eye className="mr-2 h-4 w-4" />
                    View
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => navigate({ to: '/posts/$postId', params: { postId: post.id } })}>
                    <Pencil className="mr-2 h-4 w-4" />
                    Edit
                  </DropdownMenuItem>
                  <DropdownMenuItem onClick={() => handleDelete(post.id)} className="text-red-600">
                    <Trash className="mr-2 h-4 w-4" />
                    Delete
                  </DropdownMenuItem>
                </DropdownMenuContent>
              </DropdownMenu>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/posts/components/__tests__/PostsTable.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/posts/components/
git commit -m "feat(admin): add PostsTable to posts feature"
```

---

## Feature 4: categories - Categories Management

**Description:** CategoryTree, CategorySelect, useCategories

**Target Path:** `admin/src/features/categories/`

### Task 4.1: Create categories hooks

**Files:**
- Create: `admin/src/features/categories/types/categories.ts`
- Create: `admin/src/features/categories/hooks/useCategories.ts`
- Create: `admin/src/features/categories/hooks/useCreateCategory.ts`
- Create: `admin/src/features/categories/hooks/useUpdateCategory.ts`
- Create: `admin/src/features/categories/hooks/useDeleteCategory.ts`
- Create: `admin/src/features/categories/hooks/index.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/categories/__tests__/categories-hooks.test.ts
import { describe, it, expect } from 'vitest';

describe('categories hooks', () => {
  it('exports useCategories', () => {
    const { useCategories } = require('../hooks');
    expect(useCategories).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/categories/__tests__/categories-hooks.test.ts
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/categories/types/categories.ts
export interface Category {
  id: string;
  name: string;
  slug: string;
  description?: string;
  parentId?: string;
  siteId: string;
  order: number;
  createdAt: string;
  updatedAt: string;
  children?: Category[];
}

export interface CreateCategoryRequest {
  name: string;
  slug: string;
  description?: string;
  parentId?: string;
}
```

```typescript
// admin/src/features/categories/hooks/useCategories.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category } from '../types/categories';

export function useCategories(siteSlug: string) {
  return createQuery({
    queryKey: ['categories', siteSlug],
    queryFn: async () => {
      const response = await apiClient.get<Category[]>(`/sites/${siteSlug}/categories`);
      return response;
    },
  });
}
```

```typescript
// admin/src/features/categories/hooks/useCreateCategory.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category, CreateCategoryRequest } from '../types/categories';

export function useCreateCategory() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; category: CreateCategoryRequest }): Promise<Category> => {
      const response = await apiClient.post<Category>(
        `/sites/${data.siteSlug}/categories`,
        data.category
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['categories', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/categories/hooks/useUpdateCategory.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Category, CreateCategoryRequest } from '../types/categories';

export function useUpdateCategory() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; categoryId: string; category: Partial<CreateCategoryRequest> }): Promise<Category> => {
      const response = await apiClient.patch<Category>(
        `/sites/${data.siteSlug}/categories/${data.categoryId}`,
        data.category
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['categories', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/categories/hooks/useDeleteCategory.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteCategory() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; categoryId: string }) => {
      await apiClient.delete(`/sites/${data.siteSlug}/categories/${data.categoryId}`);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['categories', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/categories/hooks/index.ts
export { useCategories } from './useCategories';
export { useCreateCategory } from './useCreateCategory';
export { useUpdateCategory } from './useUpdateCategory';
export { useDeleteCategory } from './useDeleteCategory';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/categories/__tests__/categories-hooks.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/categories/
git commit -m "feat(admin): add categories hooks to categories feature"
```

---

### Task 4.2: Create CategoryTree component

**Files:**
- Create: `admin/src/features/categories/components/CategoryTree.tsx`
- Create: `admin/src/features/categories/components/CategorySelect.tsx`

**Step 1: Write the failing test**

```typescript
// admin/src/features/categories/components/__tests__/CategoryTree.test.tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { CategoryTree } from '../CategoryTree';
import type { Category } from '../../types/categories';

const mockCategories: Category[] = [
  { id: '1', name: 'Parent', slug: 'parent', siteId: 'site-1', order: 0, createdAt: '', updatedAt: '' },
  { id: '2', name: 'Child', slug: 'child', parentId: '1', siteId: 'site-1', order: 0, createdAt: '', updatedAt: '' },
];

describe('CategoryTree', () => {
  it('renders categories', () => {
    render(<CategoryTree categories={mockCategories} onEdit={() => {}} onDelete={() => {}} />);
    expect(screen.getByText('Parent')).toBeInTheDocument();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/categories/components/__tests__/CategoryTree.test.tsx
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/categories/components/CategoryTree.tsx
'use client';

import { useState } from 'react';
import type { Category } from '../types/categories';
import { ChevronRight, ChevronDown, Folder, FolderOpen, Pencil, Trash } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface CategoryTreeProps {
  categories: Category[];
  onEdit?: (category: Category) => void;
  onDelete?: (category: Category) => void;
}

interface CategoryNodeProps {
  category: Category;
  children: Category[];
  level: number;
  onEdit?: (category: Category) => void;
  onDelete?: (category: Category) => void;
}

function buildTree(categories: Category[]): Map<string, Category[]> {
  const tree = new Map<string, Category[]>();
  categories.forEach((cat) => {
    const children = tree.get(cat.parentId || '') || [];
    children.push(cat);
    tree.set(cat.parentId || '', children);
  });
  return tree;
}

function CategoryNode({ category, children, level, onEdit, onDelete }: CategoryNodeProps) {
  const [expanded, setExpanded] = useState(false);
  const hasChildren = children.length > 0;

  return (
    <div>
      <div className="flex items-center gap-2 py-1 px-2 hover:bg-gray-50 rounded" style={{ paddingLeft: `${level * 20 + 8}px` }}>
        {hasChildren ? (
          <button onClick={() => setExpanded(!expanded)} className="p-0.5 hover:bg-gray-200 rounded">
            {expanded ? <ChevronDown className="h-4 w-4" /> : <ChevronRight className="h-4 w-4" />}
          </button>
        ) : (
          <span className="w-5" />
        )}
        {expanded ? <FolderOpen className="h-4 w-4 text-yellow-500" /> : <Folder className="h-4 w-4 text-yellow-500" />}
        <span className="flex-1">{category.name}</span>
        {onEdit && (
          <Button variant="ghost" size="sm" onClick={() => onEdit(category)}>
            <Pencil className="h-3 w-3" />
          </Button>
        )}
        {onDelete && (
          <Button variant="ghost" size="sm" onClick={() => onDelete(category)}>
            <Trash className="h-3 w-3" />
          </Button>
        )}
      </div>
      {expanded && children.map((child) => (
        <CategoryNode
          key={child.id}
          category={child}
          children={[]}
          level={level + 1}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}

export function CategoryTree({ categories, onEdit, onDelete }: CategoryTreeProps) {
  const tree = buildTree(categories);
  const rootCategories = tree.get('') || [];

  return (
    <div className="space-y-1">
      {rootCategories.map((category) => (
        <CategoryNode
          key={category.id}
          category={category}
          children={tree.get(category.id) || []}
          level={0}
          onEdit={onEdit}
          onDelete={onDelete}
        />
      ))}
    </div>
  );
}
```

```typescript
// admin/src/features/categories/components/CategorySelect.tsx
'use client';

import { useQuery } from '@tanstack/react-query';
import { useCategories } from '../hooks';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface CategorySelectProps {
  siteSlug: string;
  value?: string;
  onChange: (value: string) => void;
  placeholder?: string;
}

export function CategorySelect({ siteSlug, value, onChange, placeholder = 'Select category' }: CategorySelectProps) {
  const { data: categories } = useCategories(siteSlug);

  return (
    <Select value={value} onValueChange={onChange}>
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {categories?.map((category) => (
          <SelectItem key={category.id} value={category.id}>
            {category.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/categories/components/__tests__/CategoryTree.test.tsx
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/categories/components/
git commit -m "feat(admin): add CategoryTree and CategorySelect to categories feature"
```

---

## Feature 5: tags - Tags Management

**Description:** TagsTable, TagSelect, useTags

**Target Path:** `admin/src/features/tags/`

### Task 5.1: Create tags hooks and components

**Files:**
- Create: `admin/src/features/tags/types/tags.ts`
- Create: `admin/src/features/tags/hooks/useTags.ts`
- Create: `admin/src/features/tags/hooks/useCreateTag.ts`
- Create: `admin/src/features/tags/hooks/useUpdateTag.ts`
- Create: `admin/src/features/tags/hooks/useDeleteTag.ts`
- Create: `admin/src/features/tags/hooks/index.ts`
- Create: `admin/src/features/tags/components/TagsTable.tsx`
- Create: `admin/src/features/tags/components/TagSelect.tsx`
- Create: `admin/src/features/tags/index.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/tags/__tests__/tags-hooks.test.ts
import { describe, it, expect } from 'vitest';

describe('tags hooks', () => {
  it('exports useTags', () => {
    const { useTags } = require('../hooks');
    expect(useTags).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/tags/__tests__/tags-hooks.test.ts
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/tags/types/tags.ts
export interface Tag {
  id: string;
  name: string;
  slug: string;
  siteId: string;
  createdAt: string;
  updatedAt: string;
}

export interface CreateTagRequest {
  name: string;
  slug: string;
}
```

```typescript
// admin/src/features/tags/hooks/useTags.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { Tag } from '../types/tags';

export function useTags(siteSlug: string, params: { page?: number; pageSize?: number } = {}) {
  return createQuery({
    queryKey: ['tags', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<Tag>>(
        `/sites/${siteSlug}/tags`,
        { params }
      );
      return response;
    },
  });
}
```

```typescript
// admin/src/features/tags/hooks/useCreateTag.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Tag, CreateTagRequest } from '../types/tags';

export function useCreateTag() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; tag: CreateTagRequest }): Promise<Tag> => {
      const response = await apiClient.post<Tag>(`/sites/${data.siteSlug}/tags`, data.tag);
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tags', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/tags/hooks/useUpdateTag.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { Tag, CreateTagRequest } from '../types/tags';

export function useUpdateTag() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; tagId: string; tag: Partial<CreateTagRequest> }): Promise<Tag> => {
      const response = await apiClient.patch<Tag>(
        `/sites/${data.siteSlug}/tags/${data.tagId}`,
        data.tag
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tags', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/tags/hooks/useDeleteTag.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteTag() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; tagId: string }) => {
      await apiClient.delete(`/sites/${data.siteSlug}/tags/${data.tagId}`);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['tags', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/tags/hooks/index.ts
export { useTags } from './useTags';
export { useCreateTag } from './useCreateTag';
export { useUpdateTag } from './useUpdateTag';
export { useDeleteTag } from './useDeleteTag';
```

```typescript
// admin/src/features/tags/components/TagsTable.tsx
'use client';

import { useTags, useDeleteTag } from '../hooks';
import type { Tag } from '../types/tags';
import { Table, TableBody, TableCell, TableHead, TableHeader, TableRow } from '@/components/ui/table';
import { Button } from '@/components/ui/button';
import { Pencil, Trash } from 'lucide-react';

interface TagsTableProps {
  siteSlug: string;
  onEdit?: (tag: Tag) => void;
}

export function TagsTable({ siteSlug, onEdit }: TagsTableProps) {
  const { data, isLoading } = useTags(siteSlug);
  const deleteTag = useDeleteTag();

  const handleDelete = async (tagId: string) => {
    if (confirm('Are you sure you want to delete this tag?')) {
      await deleteTag.mutateAsync({ siteSlug, tagId });
    }
  };

  if (isLoading) return <div>Loading...</div>;

  return (
    <Table>
      <TableHeader>
        <TableRow>
          <TableHead>Name</TableHead>
          <TableHead>Slug</TableHead>
          <TableHead>Created</TableHead>
          <TableHead className="w-[100px]">Actions</TableHead>
        </TableRow>
      </TableHeader>
      <TableBody>
        {data?.data.map((tag) => (
          <TableRow key={tag.id}>
            <TableCell className="font-medium">{tag.name}</TableCell>
            <TableCell>{tag.slug}</TableCell>
            <TableCell>{new Date(tag.createdAt).toLocaleDateString()}</TableCell>
            <TableCell>
              <div className="flex gap-2">
                {onEdit && (
                  <Button variant="ghost" size="sm" onClick={() => onEdit(tag)}>
                    <Pencil className="h-3 w-3" />
                  </Button>
                )}
                <Button variant="ghost" size="sm" onClick={() => handleDelete(tag.id)}>
                  <Trash className="h-3 w-3" />
                </Button>
              </div>
            </TableCell>
          </TableRow>
        ))}
      </TableBody>
    </Table>
  );
}
```

```typescript
// admin/src/features/tags/components/TagSelect.tsx
'use client';

import { useTags } from '../hooks';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/ui/select';

interface TagSelectProps {
  siteSlug: string;
  value?: string[];
  onChange: (value: string[]) => void;
  placeholder?: string;
}

export function TagSelect({ siteSlug, value = [], onChange, placeholder = 'Select tags' }: TagSelectProps) {
  const { data } = useTags(siteSlug, { pageSize: 100 });

  return (
    <Select value={value[0]} onValueChange={(val) => onChange([val])}>
      <SelectTrigger>
        <SelectValue placeholder={placeholder} />
      </SelectTrigger>
      <SelectContent>
        {data?.data.map((tag) => (
          <SelectItem key={tag.id} value={tag.id}>
            {tag.name}
          </SelectItem>
        ))}
      </SelectContent>
    </Select>
  );
}
```

```typescript
// admin/src/features/tags/index.ts
export * from './hooks';
export * from './components/TagsTable';
export * from './components/TagSelect';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/tags/__tests__/tags-hooks.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/tags/
git commit -m "feat(admin): add tags feature module"
```

---

## Feature 6: media - Media Management

**Description:** MediaLibrary, MediaUploader, useMediaFiles

**Target Path:** `admin/src/features/media/`

### Task 6.1: Create media hooks and components

**Files:**
- Create: `admin/src/features/media/types/media.ts`
- Create: `admin/src/features/media/hooks/useMediaFiles.ts`
- Create: `admin/src/features/media/hooks/useUploadMedia.ts`
- Create: `admin/src/features/media/hooks/useDeleteMedia.ts`
- Create: `admin/src/features/media/hooks/index.ts`
- Create: `admin/src/features/media/components/MediaLibrary.tsx`
- Create: `admin/src/features/media/components/MediaUploader.tsx`
- Create: `admin/src/features/media/index.ts`

**Step 1: Write the failing test**

```typescript
// admin/src/features/media/__tests__/media-hooks.test.ts
import { describe, it, expect } from 'vitest';

describe('media hooks', () => {
  it('exports useMediaFiles', () => {
    const { useMediaFiles } = require('../hooks');
    expect(useMediaFiles).toBeDefined();
  });
});
```

**Step 2: Run test to verify it fails**

```bash
cd admin && bun run vitest run src/features/media/__tests__/media-hooks.test.ts
```
Expected: FAIL

**Step 3: Write implementation**

```typescript
// admin/src/features/media/types/media.ts
export interface MediaFile {
  id: string;
  filename: string;
  url: string;
  size: number;
  mimeType: string;
  width?: number;
  height?: number;
  alt?: string;
  caption?: string;
  siteId: string;
  folderId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface UploadMediaRequest {
  file: File;
  alt?: string;
  caption?: string;
  folderId?: string;
}
```

```typescript
// admin/src/features/media/hooks/useMediaFiles.ts
import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { MediaFile } from '../types/media';

export function useMediaFiles(siteSlug: string, params: { page?: number; pageSize?: number; folderId?: string } = {}) {
  return createQuery({
    queryKey: ['media', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<MediaFile>>(
        `/sites/${siteSlug}/media`,
        { params }
      );
      return response;
    },
  });
}
```

```typescript
// admin/src/features/media/hooks/useUploadMedia.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { MediaFile, UploadMediaRequest } from '../types/media';

export function useUploadMedia() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; file: File }): Promise<MediaFile> => {
      const formData = new FormData();
      formData.append('file', data.file);
      const response = await apiClient.post<MediaFile>(
        `/sites/${data.siteSlug}/media`,
        formData
      );
      return response;
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['media', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/media/hooks/useDeleteMedia.ts
import { createMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';

export function useDeleteMedia() {
  const queryClient = useQueryClient();
  return createMutation({
    mutationFn: async (data: { siteSlug: string; mediaId: string }) => {
      await apiClient.delete(`/sites/${data.siteSlug}/media/${data.mediaId}`);
    },
    onSuccess: (_, variables) => {
      queryClient.invalidateQueries({ queryKey: ['media', variables.siteSlug] });
    },
  });
}
```

```typescript
// admin/src/features/media/hooks/index.ts
export { useMediaFiles } from './useMediaFiles';
export { useUploadMedia } from './useUploadMedia';
export { useDeleteMedia } from './useDeleteMedia';
```

```typescript
// admin/src/features/media/components/MediaLibrary.tsx
'use client';

import { useMediaFiles, useDeleteMedia } from '../hooks';
import type { MediaFile } from '../types/media';
import { Button } from '@/components/ui/button';
import { Trash, Download } from 'lucide-react';

interface MediaLibraryProps {
  siteSlug: string;
  onSelect?: (media: MediaFile) => void;
}

export function MediaLibrary({ siteSlug, onSelect }: MediaLibraryProps) {
  const { data, isLoading } = useMediaFiles(siteSlug);
  const deleteMedia = useDeleteMedia();

  const handleDelete = async (mediaId: string) => {
    if (confirm('Are you sure you want to delete this file?')) {
      await deleteMedia.mutateAsync({ siteSlug, mediaId });
    }
  };

  if (isLoading) return <div>Loading...</div>;

  return (
    <div className="grid grid-cols-2 md:grid-cols-4 lg:grid-cols-6 gap-4">
      {data?.data.map((media) => (
        <div
          key={media.id}
          className="relative group border rounded-lg overflow-hidden cursor-pointer"
          onClick={() => onSelect?.(media)}
        >
          {media.mimeType.startsWith('image/') ? (
            <img src={media.url} alt={media.alt || media.filename} className="w-full h-32 object-cover" />
          ) : (
            <div className="w-full h-32 flex items-center justify-center bg-gray-100">
              <span className="text-xs">{media.mimeType}</span>
            </div>
          )}
          <div className="p-2">
            <p className="text-xs truncate">{media.filename}</p>
          </div>
          <div className="absolute top-2 right-2 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
            <Button variant="secondary" size="sm" onClick={(e) => { e.stopPropagation(); window.open(media.url); }}>
              <Download className="h-3 w-3" />
            </Button>
            <Button variant="destructive" size="sm" onClick={(e) => { e.stopPropagation(); handleDelete(media.id); }}>
              <Trash className="h-3 w-3" />
            </Button>
          </div>
        </div>
      ))}
    </div>
  );
}
```

```typescript
// admin/src/features/media/components/MediaUploader.tsx
'use client';

import { useCallback } from 'react';
import { useUploadMedia } from '../hooks';
import { useDropzone } from 'react-dropzone';
import { Upload } from 'lucide-react';
import { Button } from '@/components/ui/button';

interface MediaUploaderProps {
  siteSlug: string;
  onSuccess?: () => void;
}

export function MediaUploader({ siteSlug, onSuccess }: MediaUploaderProps) {
  const uploadMedia = useUploadMedia();

  const onDrop = useCallback(async (acceptedFiles: File[]) => {
    for (const file of acceptedFiles) {
      await uploadMedia.mutateAsync({ siteSlug, file });
    }
    onSuccess?.();
  }, [siteSlug, uploadMedia, onSuccess]);

  const { getRootProps, getInputProps, isDragActive } = useDropzone({ onDrop });

  return (
    <div
      {...getRootProps()}
      className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
        isDragActive ? 'border-primary bg-primary/5' : 'border-gray-300'
      }`}
    >
      <input {...getInputProps()} />
      <Upload className="mx-auto h-12 w-12 text-gray-400" />
      <p className="mt-2 text-sm text-gray-600">
        {isDragActive ? 'Drop the files here' : 'Drag and drop files here, or click to select'}
      </p>
      <Button variant="secondary" className="mt-4" disabled={uploadMedia.isPending}>
        {uploadMedia.isPending ? 'Uploading...' : 'Select Files'}
      </Button>
    </div>
  );
}
```

```typescript
// admin/src/features/media/index.ts
export * from './hooks';
export * from './components/MediaLibrary';
export * from './components/MediaUploader';
```

**Step 4: Run test to verify it passes**

```bash
cd admin && bun run vitest run src/features/media/__tests__/media-hooks.test.ts
```
Expected: PASS

**Step 5: Commit**

```bash
git add admin/src/features/media/
git commit -m "feat(admin): add media feature module"
```

---

## Feature 7-15: Additional Features (Summary)

Due to length, here are the remaining features summarized:

### Feature 7: users

**Files to create:**
- `admin/src/features/users/types/users.ts`
- `admin/src/features/users/hooks/useUsers.ts`
- `admin/src/features/users/hooks/useCreateUser.ts`
- `admin/src/features/users/hooks/useUpdateUser.ts`
- `admin/src/features/users/hooks/useDeleteUser.ts`
- `admin/src/features/users/hooks/index.ts`
- `admin/src/features/users/components/UsersTable.tsx`
- `admin/src/features/users/components/UserFormDialog.tsx`
- `admin/src/features/users/index.ts`

---

### Feature 8: roles

**Files to create:**
- `admin/src/features/roles/types/roles.ts`
- `admin/src/features/roles/hooks/useRoles.ts`
- `admin/src/features/roles/hooks/useCreateRole.ts`
- `admin/src/features/roles/hooks/useUpdateRole.ts`
- `admin/src/features/roles/hooks/useDeleteRole.ts`
- `admin/src/features/roles/hooks/index.ts`
- `admin/src/features/roles/components/RolesTable.tsx`
- `admin/src/features/roles/components/RoleFormDialog.tsx`
- `admin/src/features/roles/components/RolePermissions.tsx`
- `admin/src/features/roles/index.ts`

---

### Feature 9: sites

**Files to create:**
- `admin/src/features/sites/types/sites.ts`
- `admin/src/features/sites/hooks/useSites.ts`
- `admin/src/features/sites/hooks/useCreateSite.ts`
- `admin/src/features/sites/hooks/useUpdateSite.ts`
- `admin/src/features/sites/hooks/useDeleteSite.ts`
- `admin/src/features/sites/hooks/index.ts`
- `admin/src/features/sites/components/SitesTable.tsx`
- `admin/src/features/sites/components/SiteFormDialog.tsx`
- `admin/src/features/sites/index.ts`

---

### Feature 10: settings

**Files to create:**
- `admin/src/features/settings/types/settings.ts`
- `admin/src/features/settings/hooks/useSettings.ts`
- `admin/src/features/settings/hooks/useUpdateSettings.ts`
- `admin/src/features/settings/hooks/index.ts`
- `admin/src/features/settings/components/SettingsForm.tsx`
- `admin/src/features/settings/index.ts`

---

### Feature 11: comments

**Files to create:**
- `admin/src/features/comments/types/comments.ts`
- `admin/src/features/comments/hooks/useComments.ts`
- `admin/src/features/comments/hooks/useApproveComment.ts`
- `admin/src/features/comments/hooks/useDeleteComment.ts`
- `admin/src/features/comments/hooks/index.ts`
- `admin/src/features/comments/components/CommentsTable.tsx`
- `admin/src/features/comments/components/CommentDetailDialog.tsx`
- `admin/src/features/comments/index.ts`

---

### Feature 12: menus

**Files to create:**
- `admin/src/features/menus/types/menus.ts`
- `admin/src/features/menus/hooks/useMenus.ts`
- `admin/src/features/menus/hooks/useCreateMenu.ts`
- `admin/src/features/menus/hooks/useUpdateMenu.ts`
- `admin/src/features/menus/hooks/useDeleteMenu.ts`
- `admin/src/features/menus/hooks/index.ts`
- `admin/src/features/menus/components/MenusTable.tsx`
- `admin/src/features/menus/components/MenuFormDialog.tsx`
- `admin/src/features/menus/components/MenuItemsEditor.tsx`
- `admin/src/features/menus/index.ts`

---

### Feature 13: redirects

**Files to create:**
- `admin/src/features/redirects/types/redirects.ts`
- `admin/src/features/redirects/hooks/useRedirects.ts`
- `admin/src/features/redirects/hooks/useCreateRedirect.ts`
- `admin/src/features/redirects/hooks/useUpdateRedirect.ts`
- `admin/src/features/redirects/hooks/useDeleteRedirect.ts`
- `admin/src/features/redirects/hooks/useImportRedirects.ts`
- `admin/src/features/redirects/hooks/index.ts`
- `admin/src/features/redirects/components/RedirectsTable.tsx`
- `admin/src/features/redirects/components/RedirectFormDialog.tsx`
- `admin/src/features/redirects/components/CsvImportDialog.tsx`
- `admin/src/features/redirects/index.ts`

---

### Feature 14: api-keys

**Files to create:**
- `admin/src/features/api-keys/types/api-keys.ts`
- `admin/src/features/api-keys/hooks/useApiKeys.ts`
- `admin/src/features/api-keys/hooks/useCreateApiKey.ts`
- `admin/src/features/api-keys/hooks/useDeleteApiKey.ts`
- `admin/src/features/api-keys/hooks/index.ts`
- `admin/src/features/api-keys/components/ApiKeysTable.tsx`
- `admin/src/features/api-keys/components/CreateApiKeyDialog.tsx`
- `admin/src/features/api-keys/index.ts`

---

### Feature 15: audit

**Files to create:**
- `admin/src/features/audit/types/audit.ts`
- `admin/src/features/audit/hooks/useAuditLogs.ts`
- `admin/src/features/audit/hooks/index.ts`
- `admin/src/features/audit/components/AuditTable.tsx`
- `admin/src/features/audit/index.ts`

---

## Execution Summary

**Total Features:** 15

**Estimated Tasks:** ~45+ individual tasks (3 per feature: types+hooks, components, index)

**TDD Cycle per Task:**
1. Write failing test
2. Verify failure
3. Write minimal implementation
4. Verify pass
5. Commit

**Recommended Execution:**
- Use subagent-driven approach with fresh subagent per feature
- Batch features 1-3 (shared, auth, posts) in first session
- Features 4-6 (categories, tags, media) in second session
- Features 7-15 (users through audit) in remaining sessions

---

## Plan Complete

**This plan provides:**
- Complete Feature-Based architecture structure
- Each feature as self-contained module
- TDD cycle for every task
- Clear file paths and test/implementation examples

**Next Steps:**
1. Start executing Feature 1 (shared) - api-client
2. Continue through all 15 features
3. Create TanStack Router routes that import from features
4. Migrate i18n to Paraglide
5. Rewrite all tests with proper mocks

---

**Two execution options:**

1. **Subagent-Driven (this session)** - Dispatch fresh subagent per task, review between tasks, fast iteration

2. **Parallel Session (separate)** - Open new session with executing-plans, batch execution with checkpoints

**Which approach?**
