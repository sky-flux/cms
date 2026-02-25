# Batch 11: Content Management Pages Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build all content management frontend pages (posts, categories, tags, media) with full TDD coverage.

**Architecture:** Astro 5 SSR pages wrapping React 19 Islands via `client:load`. DashboardLayout provides sidebar+header shell. TanStack Query manages server state, Zustand for local UI state. All API calls routed through typed wrapper in `content-api.ts`.

**Tech Stack:** Astro 5 + React 19 + BlockNote (editor) + @tanstack/react-table + @dnd-kit (category reorder) + react-dropzone (media upload) + shadcn/ui + TanStack Query v5 + Vitest + RTL

---

## Agent Division

| Agent | Tasks | Dependencies |
|-------|-------|-------------|
| Agent 1 (infra) | Tasks 1–6 | None |
| Agent 2 (posts-list) | Tasks 7–10 | Agent 1 done |
| Agent 3 (post-editor) | Tasks 11–15 | Agent 1 done |
| Agent 4 (taxonomy-media) | Tasks 16–23 | Agent 1 done |

---

## Agent 1: Shared Infrastructure

### Task 1: Install new dependencies

**Step 1: Install runtime packages**

Run in `web/`:
```bash
bun add @blocknote/core @blocknote/react @blocknote/shadcn @dnd-kit/core @dnd-kit/sortable @dnd-kit/utilities react-dropzone @tanstack/react-table
```

**Step 2: Commit**

```bash
git add package.json bun.lock
git commit -m "deps: add BlockNote, dnd-kit, react-dropzone, react-table"
```

---

### Task 2: Add FormData upload support to api-client + content-api.ts

**Files:**
- Modify: `web/src/lib/api-client.ts`
- Create: `web/src/lib/content-api.ts`
- Create: `web/src/lib/__tests__/content-api.test.ts`

**Step 1: Write content-api tests**

Create `web/src/lib/__tests__/content-api.test.ts`:

```typescript
import { describe, it, expect, vi, beforeEach } from 'vitest';

// Mock api-client
vi.mock('../api-client', () => ({
  api: {
    get: vi.fn(),
    post: vi.fn(),
    put: vi.fn(),
    delete: vi.fn(),
  },
  ApiError: class ApiError extends Error {
    constructor(public status: number, message: string, public data?: unknown) {
      super(message);
    }
  },
}));

// Mock requestFormData
vi.mock('../api-client', async (importOriginal) => {
  const mod = await importOriginal<typeof import('../api-client')>();
  return {
    ...mod,
    api: {
      get: vi.fn(),
      post: vi.fn(),
      put: vi.fn(),
      delete: vi.fn(),
    },
    requestFormData: vi.fn(),
  };
});

import { api, requestFormData } from '../api-client';
import { postsApi, categoriesApi, tagsApi, mediaApi } from '../content-api';

beforeEach(() => {
  vi.clearAllMocks();
});

describe('postsApi', () => {
  it('list calls GET with query params', async () => {
    const mockRes = { success: true, data: [], pagination: {} };
    vi.mocked(api.get).mockResolvedValue(mockRes);

    const result = await postsApi.list({ page: 1, per_page: 20, status: 'draft' });

    expect(api.get).toHaveBeenCalledWith(
      '/api/v1/site/posts?page=1&per_page=20&status=draft',
      undefined,
    );
    expect(result).toEqual(mockRes);
  });

  it('list omits undefined params', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: [] });

    await postsApi.list({ page: 1, per_page: 20 });

    expect(api.get).toHaveBeenCalledWith(
      '/api/v1/site/posts?page=1&per_page=20',
      undefined,
    );
  });

  it('get calls GET with id', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: {} });
    await postsApi.get('abc-123');
    expect(api.get).toHaveBeenCalledWith('/api/v1/site/posts/abc-123', undefined);
  });

  it('create calls POST', async () => {
    const body = { title: 'Test', content: '<p>hi</p>' };
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });
    await postsApi.create(body);
    expect(api.post).toHaveBeenCalledWith('/api/v1/site/posts', body, undefined);
  });

  it('update calls PUT with version', async () => {
    const body = { title: 'Updated', version: 2 };
    vi.mocked(api.put).mockResolvedValue({ success: true, data: {} });
    await postsApi.update('abc-123', body);
    expect(api.put).toHaveBeenCalledWith('/api/v1/site/posts/abc-123', body, undefined);
  });

  it('publish calls POST to publish endpoint', async () => {
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });
    await postsApi.publish('abc-123');
    expect(api.post).toHaveBeenCalledWith('/api/v1/site/posts/abc-123/publish', undefined, undefined);
  });

  it('getRevisions calls GET', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: [] });
    await postsApi.getRevisions('abc-123');
    expect(api.get).toHaveBeenCalledWith('/api/v1/site/posts/abc-123/revisions', undefined);
  });

  it('rollback calls POST with revision id', async () => {
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });
    await postsApi.rollback('abc-123', 'rev-456');
    expect(api.post).toHaveBeenCalledWith(
      '/api/v1/site/posts/abc-123/revisions/rev-456/rollback',
      undefined,
      undefined,
    );
  });
});

describe('categoriesApi', () => {
  it('tree calls GET', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: [] });
    await categoriesApi.tree();
    expect(api.get).toHaveBeenCalledWith('/api/v1/site/categories', undefined);
  });

  it('reorder calls PUT', async () => {
    const orders = [{ id: 'a', sort_order: 1 }];
    vi.mocked(api.put).mockResolvedValue({ success: true });
    await categoriesApi.reorder(orders);
    expect(api.put).toHaveBeenCalledWith(
      '/api/v1/site/categories/reorder',
      { orders },
      undefined,
    );
  });

  it('delete calls DELETE with id', async () => {
    vi.mocked(api.delete).mockResolvedValue({ success: true });
    await categoriesApi.delete('cat-1');
    expect(api.delete).toHaveBeenCalledWith('/api/v1/site/categories/cat-1', undefined);
  });
});

describe('tagsApi', () => {
  it('list calls GET with query params', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: [] });
    await tagsApi.list({ page: 1, per_page: 20, q: 'go' });
    expect(api.get).toHaveBeenCalledWith(
      '/api/v1/site/tags?page=1&per_page=20&q=go',
      undefined,
    );
  });

  it('suggest calls GET with query', async () => {
    vi.mocked(api.get).mockResolvedValue({ success: true, data: [] });
    await tagsApi.suggest('go');
    expect(api.get).toHaveBeenCalledWith('/api/v1/site/tags/suggest?q=go', undefined);
  });
});

describe('mediaApi', () => {
  it('upload calls requestFormData', async () => {
    const file = new File(['content'], 'test.png', { type: 'image/png' });
    vi.mocked(requestFormData).mockResolvedValue({ success: true, data: {} });
    await mediaApi.upload(file, 'alt text');
    expect(requestFormData).toHaveBeenCalledWith(
      'POST',
      '/api/v1/site/media',
      expect.any(FormData),
    );
  });

  it('delete calls DELETE with force param', async () => {
    vi.mocked(api.delete).mockResolvedValue({ success: true });
    await mediaApi.delete('m-1', true);
    expect(api.delete).toHaveBeenCalledWith('/api/v1/site/media/m-1?force=true', undefined);
  });

  it('batchDelete calls DELETE with body', async () => {
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });
    await mediaApi.batchDelete(['m-1', 'm-2']);
    expect(api.post).toHaveBeenCalledWith(
      '/api/v1/site/media/batch-delete',
      { ids: ['m-1', 'm-2'] },
      undefined,
    );
  });

  it('updateMeta calls PUT', async () => {
    vi.mocked(api.put).mockResolvedValue({ success: true, data: {} });
    await mediaApi.updateMeta('m-1', { alt_text: 'new alt' });
    expect(api.put).toHaveBeenCalledWith('/api/v1/site/media/m-1', { alt_text: 'new alt' }, undefined);
  });
});
```

**Step 2: Run tests to verify they fail**

```bash
cd web && bun run vitest run src/lib/__tests__/content-api.test.ts
```

Expected: FAIL — `content-api` module does not exist.

**Step 3: Add requestFormData to api-client.ts**

Add to `web/src/lib/api-client.ts` after the `request` function (before `export const api`):

```typescript
export async function requestFormData<T>(
  method: string,
  path: string,
  formData: FormData,
  opts?: RequestOptions,
): Promise<T> {
  const headers: Record<string, string> = { ...opts?.headers };

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: formData,
    credentials: 'include',
    signal: opts?.signal,
  });

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }));
    throw new ApiError(res.status, error.message || res.statusText, error);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}
```

**Step 4: Implement content-api.ts**

Create `web/src/lib/content-api.ts`:

```typescript
import { api, requestFormData, type RequestOptions } from './api-client';

// --- Types ---

export interface PaginationMeta {
  page: number;
  per_page: number;
  total: number;
  total_pages: number;
}

export interface PaginatedResponse<T> {
  success: boolean;
  data: T[];
  pagination: PaginationMeta;
}

export interface ApiResponse<T> {
  success: boolean;
  data: T;
}

// Posts
export interface PostSummary {
  id: string;
  title: string;
  slug: string;
  status: string;
  author: { id: string; display_name: string };
  cover_image: { id: string; url: string; thumbnail_urls?: { sm: string; md: string } } | null;
  categories: { id: string; name: string; slug: string }[];
  tags: { id: string; name: string }[];
  view_count: number;
  published_at: string | null;
  created_at: string;
  updated_at: string;
}

export interface Post extends PostSummary {
  content: string;
  content_json: unknown;
  excerpt: string;
  seo: { meta_title: string; meta_description: string; og_image_url: string } | null;
  extra_fields: Record<string, unknown> | null;
  scheduled_at: string | null;
  version: number;
}

export interface PostListParams {
  page: number;
  per_page: number;
  status?: string;
  q?: string;
  category_id?: string;
  tag_id?: string;
  author_id?: string;
  sort?: string;
  include_deleted?: boolean;
}

export interface CreatePostDTO {
  title: string;
  slug?: string;
  content?: string;
  content_json?: unknown;
  excerpt?: string;
  status?: string;
  scheduled_at?: string;
  cover_image_id?: string;
  category_ids?: string[];
  primary_category_id?: string;
  tag_ids?: string[];
  meta_title?: string;
  meta_description?: string;
  og_image_url?: string;
}

export interface UpdatePostDTO extends Partial<CreatePostDTO> {
  version: number;
}

export interface Revision {
  id: string;
  version: number;
  editor: { id: string; display_name: string };
  diff_summary: string;
  created_at: string;
}

// Categories
export interface CategoryNode {
  id: string;
  name: string;
  slug: string;
  path: string;
  description?: string;
  parent_id: string | null;
  post_count: number;
  sort_order: number;
  children: CategoryNode[];
  created_at?: string;
  updated_at?: string;
}

export interface CreateCategoryDTO {
  name: string;
  slug?: string;
  parent_id?: string | null;
  description?: string;
  sort_order?: number;
}

export interface UpdateCategoryDTO extends Partial<CreateCategoryDTO> {}

export interface ReorderItem {
  id: string;
  sort_order: number;
}

// Tags
export interface Tag {
  id: string;
  name: string;
  slug: string;
  post_count: number;
  created_at: string;
  updated_at?: string;
}

export interface TagListParams {
  page: number;
  per_page: number;
  q?: string;
  sort?: string;
}

export interface CreateTagDTO {
  name: string;
  slug?: string;
}

export interface UpdateTagDTO extends Partial<CreateTagDTO> {}

// Media
export interface MediaFile {
  id: string;
  file_name: string;
  original_name: string;
  mime_type: string;
  media_type: string;
  file_size: number;
  public_url: string;
  webp_url?: string;
  thumbnail_urls?: { sm: string; md: string };
  reference_count: number;
  created_at: string;
  updated_at: string;
}

export interface MediaFileDetail extends MediaFile {
  width: number;
  height: number;
  alt_text: string;
  title: string;
  referencing_posts: { id: string; title: string }[];
}

export interface MediaListParams {
  page: number;
  per_page: number;
  q?: string;
  media_type?: string;
}

export interface UpdateMediaDTO {
  alt_text?: string;
  title?: string;
}

export interface BatchDeleteResult {
  deleted_count: number;
  skipped: { id: string; reason: string; reference_count: number }[];
}

// --- Helpers ---

function buildQuery(params: Record<string, unknown>): string {
  const entries = Object.entries(params).filter(([, v]) => v !== undefined && v !== null);
  if (entries.length === 0) return '';
  return '?' + entries.map(([k, v]) => `${k}=${encodeURIComponent(String(v))}`).join('&');
}

// --- API Wrappers ---

export const postsApi = {
  list: (params: PostListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<PostSummary>>(
      `/api/v1/site/posts${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Post>>(`/api/v1/site/posts/${id}`, opts),

  create: (data: CreatePostDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>('/api/v1/site/posts', data, opts),

  update: (id: string, data: UpdatePostDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Post>>(`/api/v1/site/posts/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/posts/${id}`, opts),

  publish: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/api/v1/site/posts/${id}/publish`, undefined, opts),

  unpublish: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/api/v1/site/posts/${id}/unpublish`, undefined, opts),

  revertToDraft: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/api/v1/site/posts/${id}/revert-to-draft`, undefined, opts),

  restore: (id: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(`/api/v1/site/posts/${id}/restore`, undefined, opts),

  getRevisions: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Revision[]>>(`/api/v1/site/posts/${id}/revisions`, opts),

  rollback: (id: string, revisionId: string, opts?: RequestOptions) =>
    api.post<ApiResponse<Post>>(
      `/api/v1/site/posts/${id}/revisions/${revisionId}/rollback`,
      undefined,
      opts,
    ),
};

export const categoriesApi = {
  tree: (opts?: RequestOptions) =>
    api.get<ApiResponse<CategoryNode[]>>('/api/v1/site/categories', opts),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<CategoryNode>>(`/api/v1/site/categories/${id}`, opts),

  create: (data: CreateCategoryDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<CategoryNode>>('/api/v1/site/categories', data, opts),

  update: (id: string, data: UpdateCategoryDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<CategoryNode>>(`/api/v1/site/categories/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/categories/${id}`, opts),

  reorder: (orders: ReorderItem[], opts?: RequestOptions) =>
    api.put<{ success: boolean }>('/api/v1/site/categories/reorder', { orders }, opts),
};

export const tagsApi = {
  list: (params: TagListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<Tag>>(
      `/api/v1/site/tags${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Tag>>(`/api/v1/site/tags/${id}`, opts),

  create: (data: CreateTagDTO, opts?: RequestOptions) =>
    api.post<ApiResponse<Tag>>('/api/v1/site/tags', data, opts),

  update: (id: string, data: UpdateTagDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<Tag>>(`/api/v1/site/tags/${id}`, data, opts),

  delete: (id: string, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(`/api/v1/site/tags/${id}`, opts),

  suggest: (q: string, opts?: RequestOptions) =>
    api.get<ApiResponse<Tag[]>>(`/api/v1/site/tags/suggest?q=${encodeURIComponent(q)}`, opts),
};

export const mediaApi = {
  list: (params: MediaListParams, opts?: RequestOptions) =>
    api.get<PaginatedResponse<MediaFile>>(
      `/api/v1/site/media${buildQuery(params)}`,
      opts,
    ),

  get: (id: string, opts?: RequestOptions) =>
    api.get<ApiResponse<MediaFileDetail>>(`/api/v1/site/media/${id}`, opts),

  upload: (file: File, altText?: string) => {
    const formData = new FormData();
    formData.append('file', file);
    if (altText) formData.append('alt_text', altText);
    return requestFormData<ApiResponse<MediaFile>>('POST', '/api/v1/site/media', formData);
  },

  updateMeta: (id: string, data: UpdateMediaDTO, opts?: RequestOptions) =>
    api.put<ApiResponse<MediaFile>>(`/api/v1/site/media/${id}`, data, opts),

  delete: (id: string, force?: boolean, opts?: RequestOptions) =>
    api.delete<{ success: boolean }>(
      `/api/v1/site/media/${id}${force ? '?force=true' : ''}`,
      opts,
    ),

  batchDelete: (ids: string[], force?: boolean, opts?: RequestOptions) =>
    api.post<ApiResponse<BatchDeleteResult>>(
      `/api/v1/site/media/batch-delete${force ? '?force=true' : ''}`,
      { ids },
      opts,
    ),
};
```

**Step 5: Export RequestOptions from api-client**

In `web/src/lib/api-client.ts`, change the type definition to be exported:

```typescript
export type RequestOptions = {
  headers?: Record<string, string>;
  signal?: AbortSignal;
};
```

**Step 6: Run tests**

```bash
cd web && bun run vitest run src/lib/__tests__/content-api.test.ts
```

Expected: All 14 tests PASS.

**Step 7: Commit**

```bash
git add web/src/lib/api-client.ts web/src/lib/content-api.ts web/src/lib/__tests__/content-api.test.ts
git commit -m "feat(web): add content API wrapper with FormData upload support"
```

---

### Task 3: Shared hooks (use-pagination, use-debounce)

**Files:**
- Create: `web/src/hooks/use-pagination.ts`
- Create: `web/src/hooks/use-debounce.ts`
- Create: `web/src/hooks/__tests__/use-pagination.test.ts`
- Create: `web/src/hooks/__tests__/use-debounce.test.ts`

**Step 1: Write use-debounce tests**

Create `web/src/hooks/__tests__/use-debounce.test.ts`:

```typescript
import { describe, it, expect, vi } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { useDebounce } from '../use-debounce';

describe('useDebounce', () => {
  it('returns initial value immediately', () => {
    const { result } = renderHook(() => useDebounce('hello', 300));
    expect(result.current).toBe('hello');
  });

  it('debounces value changes', async () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(
      ({ value, delay }) => useDebounce(value, delay),
      { initialProps: { value: 'hello', delay: 300 } },
    );

    rerender({ value: 'world', delay: 300 });
    expect(result.current).toBe('hello');

    act(() => { vi.advanceTimersByTime(300); });
    expect(result.current).toBe('world');

    vi.useRealTimers();
  });

  it('resets timer on rapid changes', () => {
    vi.useFakeTimers();
    const { result, rerender } = renderHook(
      ({ value }) => useDebounce(value, 300),
      { initialProps: { value: 'a' } },
    );

    rerender({ value: 'b' });
    act(() => { vi.advanceTimersByTime(200); });
    rerender({ value: 'c' });
    act(() => { vi.advanceTimersByTime(200); });
    expect(result.current).toBe('a');

    act(() => { vi.advanceTimersByTime(100); });
    expect(result.current).toBe('c');

    vi.useRealTimers();
  });
});
```

**Step 2: Implement use-debounce**

Create `web/src/hooks/use-debounce.ts`:

```typescript
import { useState, useEffect } from 'react';

export function useDebounce<T>(value: T, delay: number): T {
  const [debouncedValue, setDebouncedValue] = useState(value);

  useEffect(() => {
    const timer = setTimeout(() => setDebouncedValue(value), delay);
    return () => clearTimeout(timer);
  }, [value, delay]);

  return debouncedValue;
}
```

**Step 3: Write use-pagination tests**

Create `web/src/hooks/__tests__/use-pagination.test.ts`:

```typescript
import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePagination } from '../use-pagination';

describe('usePagination', () => {
  it('starts with default values', () => {
    const { result } = renderHook(() => usePagination());
    expect(result.current.page).toBe(1);
    expect(result.current.perPage).toBe(20);
  });

  it('accepts initial values', () => {
    const { result } = renderHook(() => usePagination({ page: 2, perPage: 50 }));
    expect(result.current.page).toBe(2);
    expect(result.current.perPage).toBe(50);
  });

  it('setPage updates page', () => {
    const { result } = renderHook(() => usePagination());
    act(() => { result.current.setPage(3); });
    expect(result.current.page).toBe(3);
  });

  it('setPerPage updates perPage and resets page to 1', () => {
    const { result } = renderHook(() => usePagination({ page: 3 }));
    act(() => { result.current.setPerPage(50); });
    expect(result.current.perPage).toBe(50);
    expect(result.current.page).toBe(1);
  });

  it('resetPage sets page to 1', () => {
    const { result } = renderHook(() => usePagination({ page: 5 }));
    act(() => { result.current.resetPage(); });
    expect(result.current.page).toBe(1);
  });
});
```

**Step 4: Implement use-pagination**

Create `web/src/hooks/use-pagination.ts`:

```typescript
import { useState, useCallback } from 'react';

interface UsePaginationOptions {
  page?: number;
  perPage?: number;
}

export function usePagination(options?: UsePaginationOptions) {
  const [page, setPage] = useState(options?.page ?? 1);
  const [perPage, setPerPageValue] = useState(options?.perPage ?? 20);

  const setPerPage = useCallback((newPerPage: number) => {
    setPerPageValue(newPerPage);
    setPage(1);
  }, []);

  const resetPage = useCallback(() => setPage(1), []);

  return { page, perPage, setPage, setPerPage, resetPage };
}
```

**Step 5: Run tests**

```bash
cd web && bun run vitest run src/hooks/__tests__/
```

Expected: All 8 tests PASS.

**Step 6: Commit**

```bash
git add web/src/hooks/
git commit -m "feat(web): add use-pagination and use-debounce hooks"
```

---

### Task 4: Shared components (StatusBadge, ConfirmDialog, DataTable)

**Files:**
- Create: `web/src/components/shared/StatusBadge.tsx`
- Create: `web/src/components/shared/ConfirmDialog.tsx`
- Create: `web/src/components/shared/DataTable.tsx`
- Create: `web/src/components/shared/__tests__/StatusBadge.test.tsx`
- Create: `web/src/components/shared/__tests__/ConfirmDialog.test.tsx`
- Create: `web/src/components/shared/__tests__/DataTable.test.tsx`

**Step 1: Write StatusBadge tests**

Create `web/src/components/shared/__tests__/StatusBadge.test.tsx`:

```tsx
import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { StatusBadge } from '../StatusBadge';

describe('StatusBadge', () => {
  it('renders draft status', () => {
    render(<StatusBadge status="draft" />);
    expect(screen.getByText('Draft')).toBeInTheDocument();
  });

  it('renders published status', () => {
    render(<StatusBadge status="published" />);
    expect(screen.getByText('Published')).toBeInTheDocument();
  });

  it('renders scheduled status', () => {
    render(<StatusBadge status="scheduled" />);
    expect(screen.getByText('Scheduled')).toBeInTheDocument();
  });

  it('renders archived status', () => {
    render(<StatusBadge status="archived" />);
    expect(screen.getByText('Archived')).toBeInTheDocument();
  });

  it('applies variant styling via className', () => {
    const { container } = render(<StatusBadge status="published" />);
    const badge = container.firstChild as HTMLElement;
    expect(badge.className).toContain('bg-green');
  });
});
```

**Step 2: Implement StatusBadge**

Create `web/src/components/shared/StatusBadge.tsx`:

```tsx
import { Badge } from '@/components/ui/badge';

const statusConfig: Record<string, { label: string; className: string }> = {
  draft: { label: 'Draft', className: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300' },
  published: { label: 'Published', className: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300' },
  scheduled: { label: 'Scheduled', className: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300' },
  archived: { label: 'Archived', className: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300' },
};

interface StatusBadgeProps {
  status: string;
}

export function StatusBadge({ status }: StatusBadgeProps) {
  const config = statusConfig[status] ?? { label: status, className: '' };
  return (
    <Badge variant="outline" className={config.className}>
      {config.label}
    </Badge>
  );
}
```

**Step 3: Write ConfirmDialog tests**

Create `web/src/components/shared/__tests__/ConfirmDialog.test.tsx`:

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { ConfirmDialog } from '../ConfirmDialog';

describe('ConfirmDialog', () => {
  it('renders when open', () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="Delete item?"
        description="This cannot be undone."
        onConfirm={() => {}}
      />,
    );
    expect(screen.getByText('Delete item?')).toBeInTheDocument();
    expect(screen.getByText('This cannot be undone.')).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    render(
      <ConfirmDialog
        open={false}
        onOpenChange={() => {}}
        title="Delete item?"
        description="This cannot be undone."
        onConfirm={() => {}}
      />,
    );
    expect(screen.queryByText('Delete item?')).not.toBeInTheDocument();
  });

  it('calls onConfirm when confirm button clicked', async () => {
    const onConfirm = vi.fn();
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="Delete?"
        description="Sure?"
        onConfirm={onConfirm}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: /confirm|delete/i }));
    expect(onConfirm).toHaveBeenCalledOnce();
  });

  it('calls onOpenChange when cancel clicked', async () => {
    const onOpenChange = vi.fn();
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={onOpenChange}
        title="Delete?"
        description="Sure?"
        onConfirm={() => {}}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: /cancel/i }));
    expect(onOpenChange).toHaveBeenCalledWith(false);
  });

  it('shows loading state on confirm button', () => {
    render(
      <ConfirmDialog
        open={true}
        onOpenChange={() => {}}
        title="Delete?"
        description="Sure?"
        onConfirm={() => {}}
        loading={true}
      />,
    );
    const btn = screen.getByRole('button', { name: /confirm|delete|loading/i });
    expect(btn).toBeDisabled();
  });
});
```

**Step 4: Implement ConfirmDialog**

Create `web/src/components/shared/ConfirmDialog.tsx`:

```tsx
import {
  AlertDialog,
  AlertDialogAction,
  AlertDialogCancel,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
} from '@/components/ui/alert-dialog';
import { useTranslation } from 'react-i18next';

interface ConfirmDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  title: string;
  description: string;
  onConfirm: () => void;
  loading?: boolean;
  variant?: 'danger' | 'warning';
  confirmLabel?: string;
}

export function ConfirmDialog({
  open,
  onOpenChange,
  title,
  description,
  onConfirm,
  loading = false,
  variant = 'danger',
  confirmLabel,
}: ConfirmDialogProps) {
  const { t } = useTranslation();

  return (
    <AlertDialog open={open} onOpenChange={onOpenChange}>
      <AlertDialogContent>
        <AlertDialogHeader>
          <AlertDialogTitle>{title}</AlertDialogTitle>
          <AlertDialogDescription>{description}</AlertDialogDescription>
        </AlertDialogHeader>
        <AlertDialogFooter>
          <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
          <AlertDialogAction
            onClick={onConfirm}
            disabled={loading}
            className={variant === 'danger' ? 'bg-destructive text-white hover:bg-destructive/90' : ''}
          >
            {loading ? t('common.loading') : (confirmLabel ?? t('common.confirm'))}
          </AlertDialogAction>
        </AlertDialogFooter>
      </AlertDialogContent>
    </AlertDialog>
  );
}
```

**Step 5: Write DataTable tests**

Create `web/src/components/shared/__tests__/DataTable.test.tsx`:

```tsx
import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DataTable } from '../DataTable';
import type { ColumnDef } from '@tanstack/react-table';

interface TestRow {
  id: string;
  name: string;
}

const columns: ColumnDef<TestRow>[] = [
  { accessorKey: 'name', header: 'Name' },
];

const data: TestRow[] = [
  { id: '1', name: 'Alpha' },
  { id: '2', name: 'Beta' },
];

describe('DataTable', () => {
  it('renders column headers', () => {
    render(<DataTable columns={columns} data={data} />);
    expect(screen.getByText('Name')).toBeInTheDocument();
  });

  it('renders row data', () => {
    render(<DataTable columns={columns} data={data} />);
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('shows empty message when no data', () => {
    render(<DataTable columns={columns} data={[]} emptyMessage="No items" />);
    expect(screen.getByText('No items')).toBeInTheDocument();
  });

  it('shows loading skeleton', () => {
    const { container } = render(<DataTable columns={columns} data={[]} loading={true} />);
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });

  it('renders pagination when provided', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('calls onPageChange when next page clicked', async () => {
    const onPageChange = vi.fn();
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={onPageChange}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: /next/i }));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  it('disables prev button on first page', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: /prev/i })).toBeDisabled();
  });

  it('disables next button on last page', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 3, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: /next/i })).toBeDisabled();
  });
});
```

**Step 6: Implement DataTable**

Create `web/src/components/shared/DataTable.tsx`:

```tsx
import {
  flexRender,
  getCoreRowModel,
  useReactTable,
  type ColumnDef,
} from '@tanstack/react-table';
import { Button } from '@/components/ui/button';
import { Skeleton } from '@/components/ui/skeleton';
import { ChevronLeft, ChevronRight } from 'lucide-react';

interface DataTableProps<T> {
  columns: ColumnDef<T, unknown>[];
  data: T[];
  loading?: boolean;
  emptyMessage?: string;
  pagination?: { page: number; totalPages: number };
  onPageChange?: (page: number) => void;
}

export function DataTable<T>({
  columns,
  data,
  loading = false,
  emptyMessage = 'No results.',
  pagination,
  onPageChange,
}: DataTableProps<T>) {
  const table = useReactTable({
    data,
    columns,
    getCoreRowModel: getCoreRowModel(),
  });

  return (
    <div>
      <div className="rounded-md border">
        <table className="w-full caption-bottom text-sm">
          <thead className="border-b">
            {table.getHeaderGroups().map((headerGroup) => (
              <tr key={headerGroup.id}>
                {headerGroup.headers.map((header) => (
                  <th
                    key={header.id}
                    className="h-10 px-4 text-left align-middle font-medium text-muted-foreground"
                  >
                    {header.isPlaceholder
                      ? null
                      : flexRender(header.column.columnDef.header, header.getContext())}
                  </th>
                ))}
              </tr>
            ))}
          </thead>
          <tbody>
            {loading ? (
              Array.from({ length: 5 }).map((_, i) => (
                <tr key={`skeleton-${i}`} className="border-b">
                  {columns.map((_, j) => (
                    <td key={`skeleton-${i}-${j}`} className="p-4">
                      <Skeleton className="h-4 w-full" />
                    </td>
                  ))}
                </tr>
              ))
            ) : table.getRowModel().rows.length === 0 ? (
              <tr>
                <td colSpan={columns.length} className="h-24 text-center text-muted-foreground">
                  {emptyMessage}
                </td>
              </tr>
            ) : (
              table.getRowModel().rows.map((row) => (
                <tr key={row.id} className="border-b transition-colors hover:bg-muted/50">
                  {row.getVisibleCells().map((cell) => (
                    <td key={cell.id} className="p-4 align-middle">
                      {flexRender(cell.column.columnDef.cell, cell.getContext())}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {pagination && onPageChange && (
        <div className="flex items-center justify-end gap-2 py-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(pagination.page - 1)}
            disabled={pagination.page <= 1}
            aria-label="Previous page"
          >
            <ChevronLeft className="h-4 w-4" />
            <span className="sr-only">Prev</span>
          </Button>
          <span className="text-sm text-muted-foreground">
            {pagination.page} / {pagination.totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(pagination.page + 1)}
            disabled={pagination.page >= pagination.totalPages}
            aria-label="Next page"
          >
            <ChevronRight className="h-4 w-4" />
            <span className="sr-only">Next</span>
          </Button>
        </div>
      )}
    </div>
  );
}
```

**Step 7: Run tests**

```bash
cd web && bun run vitest run src/components/shared/__tests__/
```

Expected: All 18 tests PASS.

**Step 8: Commit**

```bash
git add web/src/components/shared/
git commit -m "feat(web): add StatusBadge, ConfirmDialog, DataTable shared components"
```

---

### Task 5: Extend i18n with content management keys

**Files:**
- Modify: `web/src/i18n/locales/en.json`
- Modify: `web/src/i18n/locales/zh-CN.json`

**Step 1: Add English content keys**

Add a `"content"` section to `en.json` after the `"messages"` section. See the full key list in the design document `docs/plans/2026-02-26-batch11-content-pages-design.md` under "i18n" section. All keys live under `content.*` namespace.

**Step 2: Add Chinese content keys**

Mirror all `content.*` keys in `zh-CN.json` with Chinese translations:
- `content.posts` → "文章"
- `content.newPost` → "新建文章"
- `content.categories` → "分类"
- `content.tags` → "标签"
- `content.media` → "媒体库"
- (and all other keys from design doc)

**Step 3: Commit**

```bash
git add web/src/i18n/locales/
git commit -m "feat(web): add i18n keys for content management pages"
```

---

### Task 6: Install shadcn table component

**Step 1: Install table component via shadcn CLI**

```bash
cd web && bunx shadcn@latest add table
```

This creates `web/src/components/ui/table.tsx`.

**Step 2: Commit**

```bash
git add web/src/components/ui/table.tsx
git commit -m "feat(web): add shadcn table component"
```

---

## Agent 2: Posts List

### Task 7: PostsTable component

**Files:**
- Create: `web/src/components/content/PostsTable.tsx`
- Create: `web/src/components/content/__tests__/PostsTable.test.tsx`

**Key behaviors to test:**
1. Renders column headers (title, status, author, published_at, actions)
2. Renders post rows with StatusBadge
3. Shows loading skeletons
4. Shows empty state with CTA
5. Filter bar: status Select triggers callback
6. Search input triggers debounced callback
7. Click "New Post" triggers navigation callback
8. Checkbox selection + bulk action toolbar appears
9. Pagination renders from pagination prop

**Implementation notes:**
- Uses `DataTable` internally
- Filter bar at top with: status `Select` (all/draft/published/scheduled/archived), search `Input`, "New Post" `Button`
- Columns: checkbox (optional), title (link), status (StatusBadge), author, categories (badges), published_at (formatted), actions (DropdownMenu with edit/delete)
- Props: `posts`, `pagination`, `loading`, `onPageChange`, `onStatusFilter`, `onSearch`, `onDelete`
- TanStack Query: the Astro page wraps PostsTable inside a React island that calls `useQuery` with `postsApi.list()`

**Commit:** `feat(web): add PostsTable component with filters and pagination`

---

### Task 8: PostStatusActions component

**Files:**
- Create: `web/src/components/content/PostStatusActions.tsx`
- Create: `web/src/components/content/__tests__/PostStatusActions.test.tsx`

**Key behaviors:**
1. Shows "Publish" button for draft posts
2. Shows "Unpublish" + "Revert to Draft" for published posts
3. Shows "Publish Now" + "Revert to Draft" for scheduled posts
4. Shows "Republish" + "Revert to Draft" for archived posts
5. Publish action calls `postsApi.publish()`
6. Loading state disables buttons
7. Schedule button opens date picker (for draft → scheduled transition)

**Implementation notes:**
- Props: `postId`, `status`, `onStatusChange` callback
- Each button uses `useMutation` from TanStack Query
- Schedule: secondary button opens a popover with a datetime input
- On success: calls `onStatusChange()` to trigger re-fetch

**Commit:** `feat(web): add PostStatusActions with contextual status transitions`

---

### Task 9: Posts list Astro page

**Files:**
- Create: `web/src/pages/dashboard/posts/index.astro`
- Create: `web/src/components/content/PostsListPage.tsx` (React island wrapper)

**Implementation notes:**
- Astro page uses DashboardLayout, renders `<PostsListPage client:load />`
- PostsListPage React component:
  - Uses `useQuery` with `postsApi.list()` (key: `['posts', filters]`)
  - Manages filter state: status, search (debounced), page/perPage
  - Passes data to PostsTable
  - Delete: `useMutation` → `postsApi.delete()` → invalidate query
  - Navigation: `window.location.href` for new/edit links

**Commit:** `feat(web): add posts list page with filtering and pagination`

---

### Task 10: Run posts list tests + integration check

```bash
cd web && bun run vitest run src/components/content/__tests__/PostsTable.test.tsx src/components/content/__tests__/PostStatusActions.test.tsx
```

Expected: All tests PASS.

**Commit:** All Agent 2 work should be committed by this point.

---

## Agent 3: Post Editor

### Task 11: PostEditor component (BlockNote integration)

**Files:**
- Create: `web/src/components/content/PostEditor.tsx`
- Create: `web/src/components/content/__tests__/PostEditor.test.tsx`

**Key behaviors:**
1. Renders title input
2. Renders BlockNote editor (may need to mock in tests — BlockNote depends on DOM APIs)
3. Save button calls `postsApi.update()` with content_json + version
4. New post: create button calls `postsApi.create()`
5. Metadata panel: categories multi-select, tags autocomplete, cover image, excerpt, SEO
6. Auto-save triggers after 30s for drafts (use setInterval or useEffect + debounce)
7. Version conflict (409) shows toast with refresh suggestion
8. Unsaved changes: tracks dirty state

**Implementation notes:**
- BlockNote: `useCreateBlockNote()` hook creates editor, `BlockNoteView` renders it
- BlockNote output: `editor.document` → JSON (content_json), `editor.domElement?.innerHTML` → HTML (content)
- Two-column layout: `grid grid-cols-[1fr_320px]`
- Right panel sections: each in a collapsible Card
- Auto-save: `useEffect` with `setTimeout(30000)`, only for status === 'draft'
- Editor store: `useEditorStore.saveDraft(postId, JSON.stringify(content_json))` for local backup
- Cover image: opens MediaLibrary in dialog mode (shared picker component — simplified: just an Input for URL in V1)

**Testing notes:**
- BlockNote editor requires DOM — mock `@blocknote/react` in tests:
  ```typescript
  vi.mock('@blocknote/react', () => ({
    useCreateBlockNote: () => ({ document: [], domElement: { innerHTML: '' } }),
    BlockNoteView: ({ children }: any) => <div data-testid="blocknote-editor">{children}</div>,
  }));
  vi.mock('@blocknote/shadcn', () => ({
    ShadCNDefaultComponents: {},
  }));
  ```

**Commit:** `feat(web): add PostEditor with BlockNote rich text and metadata panel`

---

### Task 12: RevisionHistory component

**Files:**
- Create: `web/src/components/content/RevisionHistory.tsx`
- Create: `web/src/components/content/__tests__/RevisionHistory.test.tsx`

**Key behaviors:**
1. Renders list of revisions with version, editor, diff_summary, date
2. Current version (highest) is highlighted
3. Rollback button opens ConfirmDialog
4. Rollback calls `postsApi.rollback(postId, revisionId)`
5. Empty state when no revisions
6. Loading state

**Implementation notes:**
- Props: `postId`
- Uses `useQuery` with `postsApi.getRevisions(postId)` (key: `['revisions', postId]`)
- Rollback: `useMutation` → on success, navigate back to edit page or invalidate
- Timeline layout: vertical list with version badges

**Commit:** `feat(web): add RevisionHistory with rollback support`

---

### Task 13: Post new/edit/revisions Astro pages

**Files:**
- Create: `web/src/pages/dashboard/posts/new.astro`
- Create: `web/src/pages/dashboard/posts/[id]/edit.astro`
- Create: `web/src/pages/dashboard/posts/[id]/revisions.astro`
- Create: `web/src/components/content/PostEditorPage.tsx` (React island wrapper for new+edit)
- Create: `web/src/components/content/RevisionsPage.tsx` (React island wrapper)

**Implementation notes:**
- `new.astro`: DashboardLayout → `<PostEditorPage client:load mode="create" />`
- `edit.astro`: DashboardLayout → `<PostEditorPage client:load mode="edit" postId={id} />`
  - Extract `id` from `Astro.params.id`
- `revisions.astro`: DashboardLayout → `<RevisionsPage client:load postId={id} />`
- PostEditorPage:
  - mode="create": empty editor, on first save → `postsApi.create()` → redirect to edit URL
  - mode="edit": `useQuery` → `postsApi.get(postId)` → populate editor
- RevisionsPage: wrapper around RevisionHistory component

**Commit:** `feat(web): add post new, edit, and revisions Astro pages`

---

### Task 14: Auto-save + version conflict handling

**Files:**
- Modify: `web/src/components/content/PostEditor.tsx` (add auto-save logic)

**Key behaviors:**
1. Auto-save interval (30s) only for draft posts
2. Saves to both local (editor-store) and API (PUT)
3. On 409 VERSION_CONFLICT: show toast "This post was modified by another user" + "Refresh" button
4. `beforeunload` event warns on unsaved changes
5. Ctrl+S / Cmd+S keyboard shortcut for manual save

**Implementation notes:**
- `useEffect` with `setInterval(30000)` — clear on unmount or status change
- Dirty tracking: compare current content_json with last saved version
- 409 handling: catch `ApiError`, check `status === 409`, show Sonner toast
- `beforeunload`: `useEffect` that adds/removes event listener based on `isDirty`

**Commit:** `feat(web): add auto-save and version conflict handling to PostEditor`

---

### Task 15: Run editor tests + integration check

```bash
cd web && bun run vitest run src/components/content/__tests__/PostEditor.test.tsx src/components/content/__tests__/RevisionHistory.test.tsx
```

Expected: All tests PASS.

---

## Agent 4: Taxonomy + Media

### Task 16: CategoryTree component

**Files:**
- Create: `web/src/components/content/CategoryTree.tsx`
- Create: `web/src/components/content/__tests__/CategoryTree.test.tsx`

**Key behaviors:**
1. Renders nested tree from CategoryNode[] data
2. Expand/collapse toggle per node
3. Shows post_count badge
4. Action buttons: edit, add child, delete
5. Empty state message
6. Drag-and-drop reorder within same level (calls onReorder)

**Implementation notes:**
- Recursive `TreeNode` component renders each CategoryNode
- @dnd-kit: `DndContext` + `SortableContext` + `useSortable` for each tree level
- On drag end: compute new `sort_order` values, call `categoriesApi.reorder()`
- Tree expand state managed via `Set<string>` in component state
- No checkbox multi-select needed — single item actions only

**Testing notes:**
- Mock @dnd-kit in tests (drag-drop testing is complex in jsdom):
  ```typescript
  vi.mock('@dnd-kit/core', () => ({
    DndContext: ({ children }: any) => <div>{children}</div>,
    closestCenter: vi.fn(),
    // ...
  }));
  ```
- Focus tests on rendering tree structure, expand/collapse, action callbacks

**Commit:** `feat(web): add CategoryTree with nested display and reorder`

---

### Task 17: CategoryForm component

**Files:**
- Create: `web/src/components/content/CategoryForm.tsx`
- Create: `web/src/components/content/__tests__/CategoryForm.test.tsx`

**Key behaviors:**
1. Create mode: empty form
2. Edit mode: pre-filled with category data
3. Name field generates slug automatically
4. Parent select dropdown (flattened tree)
5. Validation: name required, slug format
6. Submit calls `categoriesApi.create()` or `.update()`
7. Dialog closes on success

**Implementation notes:**
- Uses shadcn Dialog + react-hook-form + zod
- Props: `open`, `onOpenChange`, `category?` (undefined = create mode), `parentOptions`
- Slug auto-generation: on name change, slugify (lowercase, replace spaces with hyphens)
- Parent select: flat list with indentation showing depth

**Commit:** `feat(web): add CategoryForm dialog with create/edit modes`

---

### Task 18: Categories Astro page

**Files:**
- Create: `web/src/pages/dashboard/categories/index.astro`
- Create: `web/src/components/content/CategoriesPage.tsx` (React island)

**Implementation notes:**
- CategoriesPage:
  - `useQuery` → `categoriesApi.tree()` (key: `['categories']`)
  - "Add Category" button → opens CategoryForm dialog (create mode)
  - Tree actions → edit (opens CategoryForm with data), add child (opens CategoryForm with parent_id), delete (ConfirmDialog → `categoriesApi.delete()`)
  - Reorder: `useMutation` → `categoriesApi.reorder()` → invalidate query
  - Delete error 409: show toast with "Has subcategories" message

**Commit:** `feat(web): add categories management page with tree view`

---

### Task 19: TagsTable + TagForm components

**Files:**
- Create: `web/src/components/content/TagsTable.tsx`
- Create: `web/src/components/content/TagForm.tsx`
- Create: `web/src/components/content/__tests__/TagsTable.test.tsx`
- Create: `web/src/components/content/__tests__/TagForm.test.tsx`

**TagsTable key behaviors:**
1. Renders columns: name, slug, post_count, created_at, actions
2. Search bar with debounce
3. Sort by post_count or name
4. Edit/delete action buttons
5. Uses DataTable internally
6. Pagination

**TagForm key behaviors:**
1. Create mode: empty form, name + slug
2. Edit mode: pre-filled
3. Name auto-generates slug
4. Validation: name required

**Commit:** `feat(web): add TagsTable and TagForm components`

---

### Task 20: Tags Astro page

**Files:**
- Create: `web/src/pages/dashboard/tags/index.astro`
- Create: `web/src/components/content/TagsPage.tsx` (React island)

**Implementation notes:**
- TagsPage:
  - `useQuery` → `tagsApi.list()` (key: `['tags', filters]`)
  - Search, sort, pagination state management
  - "Add Tag" → TagForm dialog (create)
  - Edit/Delete actions → TagForm dialog / ConfirmDialog
  - Invalidate queries on mutation success

**Commit:** `feat(web): add tags management page`

---

### Task 21: MediaLibrary + MediaUploader + MediaDetailDialog

**Files:**
- Create: `web/src/components/content/MediaLibrary.tsx`
- Create: `web/src/components/content/MediaUploader.tsx`
- Create: `web/src/components/content/MediaDetailDialog.tsx`
- Create: `web/src/components/content/__tests__/MediaLibrary.test.tsx`
- Create: `web/src/components/content/__tests__/MediaUploader.test.tsx`
- Create: `web/src/components/content/__tests__/MediaDetailDialog.test.tsx`

**MediaLibrary key behaviors:**
1. Grid/list view toggle
2. Grid: thumbnail cards (4-col)
3. List: DataTable with thumbnail column
4. Filter: media_type select
5. Search with debounce
6. Multi-select with checkboxes
7. Bulk delete toolbar when items selected
8. Click item → opens MediaDetailDialog

**MediaUploader key behaviors:**
1. react-dropzone zone: dashed border, drop text
2. Accepts image/*, video/*, application/pdf
3. Shows upload progress per file
4. Calls `mediaApi.upload()` per file
5. On complete: trigger list refresh
6. Multiple file support

**MediaDetailDialog key behaviors:**
1. Preview: image thumbnail (or icon for non-images)
2. Info display: filename, type, size, dimensions, date
3. Edit: alt_text + title inputs → save button → `mediaApi.updateMeta()`
4. Reference list: shows referencing posts
5. Delete button → ConfirmDialog
6. Force delete option when referenced

**Testing notes:**
- Mock react-dropzone:
  ```typescript
  vi.mock('react-dropzone', () => ({
    useDropzone: () => ({
      getRootProps: () => ({}),
      getInputProps: () => ({}),
      isDragActive: false,
    }),
  }));
  ```

**Commit:** `feat(web): add MediaLibrary, MediaUploader, MediaDetailDialog`

---

### Task 22: Media Astro page

**Files:**
- Create: `web/src/pages/dashboard/media/index.astro`
- Create: `web/src/components/content/MediaPage.tsx` (React island)

**Implementation notes:**
- MediaPage:
  - `useQuery` → `mediaApi.list()` (key: `['media', filters]`)
  - View toggle state (grid/list)
  - Filter: media_type, search (debounced)
  - Upload: `useMutation` → `mediaApi.upload()` → invalidate query
  - Delete: single → `mediaApi.delete()`, batch → `mediaApi.batchDelete()`
  - Detail dialog: `useQuery` → `mediaApi.get(id)` when dialog opens
  - Update meta: `useMutation` → `mediaApi.updateMeta()`

**Commit:** `feat(web): add media library page with upload and management`

---

### Task 23: Run all taxonomy + media tests

```bash
cd web && bun run vitest run src/components/content/__tests__/CategoryTree.test.tsx src/components/content/__tests__/CategoryForm.test.tsx src/components/content/__tests__/TagsTable.test.tsx src/components/content/__tests__/TagForm.test.tsx src/components/content/__tests__/MediaLibrary.test.tsx src/components/content/__tests__/MediaUploader.test.tsx src/components/content/__tests__/MediaDetailDialog.test.tsx
```

Expected: All tests PASS.

---

## Final Integration

### Task 24: Run full test suite

```bash
cd web && bun run vitest run
```

Expected: All existing + new tests PASS (target: 120+ new tests).

### Task 25: Astro check

```bash
cd web && bunx astro check
```

Expected: 0 errors.

### Task 26: Final commit + update v1.0.0.md

Update `docs/v1.0.0.md`:
- Batch 11 status: ✅
- Frontend progress: 30% → 55%
- Frontend test count update
- US-030 content management marked complete

```bash
git add docs/v1.0.0.md
git commit -m "docs: update v1.0.0 with Batch 11 content pages completion"
```
