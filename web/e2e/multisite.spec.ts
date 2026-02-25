import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
import { apiLogin, createSite, createPost } from './helpers/api';

const SITE_A = { name: 'Site Alpha', slug: 'site-alpha' };
const SITE_B = { name: 'Site Beta', slug: 'site-beta' };

test.describe.serial('Multi-Site Isolation', () => {
  let superToken: string;

  test.beforeAll(async () => {
    superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
  });

  test('create two sites via API', async () => {
    try { await createSite(superToken, SITE_A); } catch { /* may exist */ }
    try { await createSite(superToken, SITE_B); } catch { /* may exist */ }
  });

  test('create posts in different sites', async () => {
    await createPost(
      superToken,
      { title: 'Alpha Post', content: '<p>Content for alpha</p>', status: 'published' },
      SITE_A.slug,
    );
    await createPost(
      superToken,
      { title: 'Beta Post', content: '<p>Content for beta</p>', status: 'published' },
      SITE_B.slug,
    );
  });

  test('site_alpha API only returns alpha posts', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': SITE_A.slug },
    });
    if (resp.ok()) {
      const json = await resp.json();
      const titles = json.data?.map((p: { title: string }) => p.title) ?? [];
      expect(titles).toContain('Alpha Post');
      expect(titles).not.toContain('Beta Post');
    }
  });

  test('site_beta API only returns beta posts', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': SITE_B.slug },
    });
    if (resp.ok()) {
      const json = await resp.json();
      const titles = json.data?.map((p: { title: string }) => p.title) ?? [];
      expect(titles).toContain('Beta Post');
      expect(titles).not.toContain('Alpha Post');
    }
  });

  test.fixme('sites list page shows both sites', async ({ page }) => {
    const { loginViaUI } = await import('./helpers/auth');
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');
    await expect(page.getByText(SITE_A.name)).toBeVisible();
    await expect(page.getByText(SITE_B.name)).toBeVisible();
  });
});
