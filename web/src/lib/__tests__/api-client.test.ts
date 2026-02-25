import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';
import { ApiError, api } from '@/lib/api-client';

// ── ApiError class ──────────────────────────────────────────────

describe('ApiError', () => {
  it('creates an error with status, message and data', () => {
    const data = { field: 'email', reason: 'invalid' };
    const err = new ApiError(422, 'Validation failed', data);

    expect(err).toBeInstanceOf(Error);
    expect(err).toBeInstanceOf(ApiError);
    expect(err.name).toBe('ApiError');
    expect(err.status).toBe(422);
    expect(err.message).toBe('Validation failed');
    expect(err.data).toEqual(data);
  });

  it('works without optional data', () => {
    const err = new ApiError(500, 'Internal Server Error');

    expect(err.status).toBe(500);
    expect(err.message).toBe('Internal Server Error');
    expect(err.data).toBeUndefined();
  });
});

// ── api.get / post / put / patch / delete ───────────────────────

describe('api', () => {
  const originalFetch = globalThis.fetch;

  beforeEach(() => {
    vi.stubGlobal('fetch', vi.fn());
  });

  afterEach(() => {
    vi.restoreAllMocks();
    globalThis.fetch = originalFetch;
  });

  function mockFetchResponse(body: unknown, init: ResponseInit = {}) {
    const status = init.status ?? 200;
    const response = new Response(JSON.stringify(body), {
      status,
      headers: { 'Content-Type': 'application/json', ...init.headers },
    });
    vi.mocked(fetch).mockResolvedValue(response);
    return response;
  }

  // ── Success responses ─────────────────────────────────────────

  it('GET returns parsed JSON on success', async () => {
    const payload = { id: 1, title: 'Hello' };
    mockFetchResponse(payload);

    const result = await api.get<typeof payload>('/posts/1');

    expect(result).toEqual(payload);
    expect(fetch).toHaveBeenCalledWith(
      '/api/posts/1',
      expect.objectContaining({
        method: 'GET',
        credentials: 'include',
      }),
    );
  });

  it('POST sends JSON body', async () => {
    const body = { title: 'New post' };
    mockFetchResponse({ id: 2, ...body });

    await api.post('/posts', body);

    expect(fetch).toHaveBeenCalledWith(
      '/api/posts',
      expect.objectContaining({
        method: 'POST',
        body: JSON.stringify(body),
      }),
    );
  });

  it('PUT sends JSON body', async () => {
    mockFetchResponse({ ok: true });

    await api.put('/posts/1', { title: 'Updated' });

    expect(fetch).toHaveBeenCalledWith(
      '/api/posts/1',
      expect.objectContaining({ method: 'PUT' }),
    );
  });

  it('PATCH sends JSON body', async () => {
    mockFetchResponse({ ok: true });

    await api.patch('/posts/1', { status: 'published' });

    expect(fetch).toHaveBeenCalledWith(
      '/api/posts/1',
      expect.objectContaining({ method: 'PATCH' }),
    );
  });

  it('DELETE does not send body', async () => {
    mockFetchResponse({ ok: true });

    await api.delete('/posts/1');

    expect(fetch).toHaveBeenCalledWith(
      '/api/posts/1',
      expect.objectContaining({
        method: 'DELETE',
        body: undefined,
      }),
    );
  });

  // ── Headers ───────────────────────────────────────────────────

  it('sets Content-Type: application/json by default', async () => {
    mockFetchResponse({});

    await api.get('/test');

    const callArgs = vi.mocked(fetch).mock.calls[0][1] as RequestInit;
    expect(callArgs.headers).toEqual(
      expect.objectContaining({ 'Content-Type': 'application/json' }),
    );
  });

  it('allows custom headers alongside defaults', async () => {
    mockFetchResponse({});

    await api.get('/test', { headers: { 'X-Custom': 'value' } });

    const callArgs = vi.mocked(fetch).mock.calls[0][1] as RequestInit;
    expect(callArgs.headers).toEqual(
      expect.objectContaining({
        'Content-Type': 'application/json',
        'X-Custom': 'value',
      }),
    );
  });

  it('sends credentials: include for httpOnly cookies', async () => {
    mockFetchResponse({});

    await api.get('/auth/me');

    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ credentials: 'include' }),
    );
  });

  // ── 204 No Content ────────────────────────────────────────────

  it('returns undefined for 204 responses', async () => {
    const response = new Response(null, { status: 204 });
    vi.mocked(fetch).mockResolvedValue(response);

    const result = await api.delete('/posts/1');

    expect(result).toBeUndefined();
  });

  // ── Error handling ────────────────────────────────────────────

  it('throws ApiError with status and message on server error', async () => {
    const errorBody = { message: 'Not Found' };
    vi.mocked(fetch).mockResolvedValue(
      new Response(JSON.stringify(errorBody), {
        status: 404,
        headers: { 'Content-Type': 'application/json' },
      }),
    );

    try {
      await api.get('/posts/999');
      expect.unreachable('Should have thrown');
    } catch (e) {
      expect(e).toBeInstanceOf(ApiError);
      const err = e as ApiError;
      expect(err.status).toBe(404);
      expect(err.message).toBe('Not Found');
      expect(err.data).toEqual(errorBody);
    }
  });

  it('falls back to statusText when error body is not JSON', async () => {
    const response = new Response('Internal Server Error', {
      status: 500,
      statusText: 'Internal Server Error',
    });
    vi.mocked(fetch).mockResolvedValue(response);

    await expect(api.get('/broken')).rejects.toThrow(ApiError);

    try {
      await api.get('/broken');
    } catch (e) {
      const err = e as ApiError;
      expect(err.status).toBe(500);
      expect(err.message).toBe('Internal Server Error');
    }
  });

  it('propagates network errors as-is', async () => {
    vi.mocked(fetch).mockRejectedValue(new TypeError('Failed to fetch'));

    await expect(api.get('/offline')).rejects.toThrow(TypeError);
  });

  // ── AbortSignal ───────────────────────────────────────────────

  it('passes AbortSignal to fetch', async () => {
    mockFetchResponse({});
    const controller = new AbortController();

    await api.get('/test', { signal: controller.signal });

    expect(fetch).toHaveBeenCalledWith(
      expect.any(String),
      expect.objectContaining({ signal: controller.signal }),
    );
  });
});
