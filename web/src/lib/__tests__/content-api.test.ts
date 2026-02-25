import { describe, it, expect, vi, beforeEach } from 'vitest';

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
