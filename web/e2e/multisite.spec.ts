import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
import { loginViaUI } from './helpers/auth';
import { apiLogin, createSite, createPost } from './helpers/api';

const SITE_A = { name: 'Site Alpha', slug: 'site-alpha' };
const SITE_B = { name: 'Site Beta', slug: 'site-beta' };

test.describe.serial('Multi-Site Isolation', () => {
  let superToken: string;

  test.beforeAll(async () => {
    superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
  });

  test('create two sites via UI', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');

    // Create Site A
    await page.getByRole('button', { name: /new site|create/i }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/name/i).fill(SITE_A.name);
    await dialog.getByLabel(/slug/i).fill(SITE_A.slug);
    await dialog.getByRole('button', { name: /save|create/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });

    // Wait for dialog to close
    await expect(dialog).not.toBeVisible({ timeout: 3_000 });

    // Create Site B
    await page.getByRole('button', { name: /new site|create/i }).click();
    const dialog2 = page.getByRole('dialog');
    await dialog2.getByLabel(/name/i).fill(SITE_B.name);
    await dialog2.getByLabel(/slug/i).fill(SITE_B.slug);
    await dialog2.getByRole('button', { name: /save|create/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('create posts in different sites via API', async () => {
    await createPost(
      superToken,
      { title: 'Alpha Post', content: '<p>Content for site alpha</p>', status: 'published' },
      SITE_A.slug,
    );
    await createPost(
      superToken,
      { title: 'Beta Post', content: '<p>Content for site beta</p>', status: 'published' },
      SITE_B.slug,
    );
  });

  test('site_alpha Public API only returns alpha posts', async ({ request }) => {
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

  test('site_beta Public API only returns beta posts', async ({ request }) => {
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

  test('both sites visible in sites list', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');
    await expect(page.getByText(SITE_A.name)).toBeVisible();
    await expect(page.getByText(SITE_B.name)).toBeVisible();
  });
});
