import { API_BASE, TEST_SUPER, TEST_SITE } from './constants';

async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
  extraHeaders?: Record<string, string>,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...extraHeaders,
  };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${method} ${path} failed (${res.status}): ${text}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

export async function setupInitialize(): Promise<string> {
  const resp = await apiCall<{
    success: boolean;
    data: { access_token: string };
  }>('POST', '/api/v1/setup/initialize', {
    admin_display_name: TEST_SUPER.displayName,
    admin_email: TEST_SUPER.email,
    admin_password: TEST_SUPER.password,
    site_name: TEST_SITE.name,
    site_slug: TEST_SITE.slug,
    site_url: TEST_SITE.url,
    locale: TEST_SITE.locale,
  });
  return resp.data.access_token;
}

export async function checkInstalled(): Promise<boolean> {
  const resp = await apiCall<{
    success: boolean;
    data: { installed: boolean };
  }>('POST', '/api/v1/setup/check');
  return resp.data.installed;
}

export async function apiLogin(email: string, password: string): Promise<string> {
  const resp = await apiCall<{
    success: boolean;
    data: { access_token: string };
  }>('POST', '/api/v1/auth/login', { email, password });
  return resp.data.access_token;
}

export async function createUser(
  token: string,
  user: { display_name: string; email: string; password: string; role: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string }> {
  const resp = await apiCall<{ success: boolean; data: { id: string } }>(
    'POST',
    '/api/v1/users',
    user,
    token,
    { 'X-Site-Slug': siteSlug },
  );
  return resp.data;
}

export async function createPost(
  token: string,
  post: { title: string; content: string; status?: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string; slug: string }> {
  const resp = await apiCall<{ success: boolean; data: { id: string; slug: string } }>(
    'POST',
    '/api/v1/posts',
    post,
    token,
    { 'X-Site-Slug': siteSlug },
  );
  return resp.data;
}

export async function createSite(
  token: string,
  site: { name: string; slug: string; domain?: string },
): Promise<void> {
  await apiCall('POST', '/api/v1/sites', site, token);
}

export async function seedTestUsers(
  superToken: string,
  users: Array<{ displayName: string; email: string; password: string; role: string }>,
  siteSlug = TEST_SITE.slug,
): Promise<void> {
  for (const u of users) {
    try {
      await createUser(superToken, {
        display_name: u.displayName,
        email: u.email,
        password: u.password,
        role: u.role,
      }, siteSlug);
    } catch {
      // User may already exist from previous run
    }
  }
}
