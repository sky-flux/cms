// Detect if running in Docker production environment
const isDockerProduction = () => {
  try {
    // Check if we're in a container by looking for Docker indicators
    return typeof window === 'undefined' &&
           process.env.NODE_ENV === 'production';
  } catch {
    return false;
  }
};

// Use internal API URL for server-side calls in Docker, public URL for browser
const getApiBase = () => {
  // Server-side in Docker production: use internal API URL
  if (isDockerProduction()) {
    return 'http://api:8080'; // Docker container-to-container communication
  }
  // Server-side in local dev or client-side: use public URL
  return import.meta.env.PUBLIC_API_URL || '/api';
};

const API_BASE = getApiBase();

export type RequestOptions = {
  headers?: Record<string, string>;
  signal?: AbortSignal;
};

export class ApiError extends Error {
  constructor(
    public status: number,
    message: string,
    public data?: unknown,
  ) {
    super(message);
    this.name = 'ApiError';
  }
}

async function request<T>(
  method: string,
  path: string,
  body?: unknown,
  opts?: RequestOptions,
  retry = true,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...opts?.headers,
  };

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
    credentials: 'include',
    signal: opts?.signal,
  });

  // If unauthorized and retry is enabled, try to refresh token
  if (res.status === 401 && retry) {
    try {
      // Call refresh endpoint - it will set new cookies automatically
      await fetch(`${API_BASE}/v1/auth/refresh`, {
        method: 'POST',
        credentials: 'include',
      });
      // Retry the original request
      return request<T>(method, path, body, opts, false);
    } catch {
      // Refresh failed, redirect to login
      window.location.href = '/login';
      throw new ApiError(401, 'Session expired');
    }
  }

  if (!res.ok) {
    const error = await res.json().catch(() => ({ message: res.statusText }));
    throw new ApiError(res.status, error.message || res.statusText, error);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

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

export const api = {
  get: <T>(path: string, opts?: RequestOptions) =>
    request<T>('GET', path, undefined, opts),
  post: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('POST', path, body, opts),
  put: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PUT', path, body, opts),
  patch: <T>(path: string, body?: unknown, opts?: RequestOptions) =>
    request<T>('PATCH', path, body, opts),
  delete: <T>(path: string, opts?: RequestOptions) =>
    request<T>('DELETE', path, undefined, opts),
};
