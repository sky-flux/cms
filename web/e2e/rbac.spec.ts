import { test, expect } from '@playwright/test';
import {
  TEST_SUPER, TEST_ADMIN, TEST_EDITOR, TEST_VIEWER, API_BASE,
} from './helpers/constants';
import { loginViaUI } from './helpers/auth';
import { apiLogin, seedTestUsers } from './helpers/api';

test.describe.serial('Role-Based Access Control', () => {
  test.beforeAll(async () => {
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
    await seedTestUsers(superToken, [TEST_ADMIN, TEST_EDITOR, TEST_VIEWER]);
  });

  // --- Navigation Visibility (fixme: DashboardShell has no role filtering) ---

  test.fixme('super sees all navigation items including sites', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /roles/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
  });

  test.fixme('admin sees system nav but NOT sites', async ({ page }) => {
    await loginViaUI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).not.toBeVisible();
  });

  test.fixme('editor sees content nav only', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await expect(page.getByRole('link', { name: /posts/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /media/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /users/i })).not.toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).not.toBeVisible();
  });

  test.fixme('viewer sees content nav but no create buttons', async ({ page }) => {
    await loginViaUI(page, TEST_VIEWER.email, TEST_VIEWER.password);
    await page.goto('/dashboard/posts');
    await expect(
      page.getByRole('button', { name: /new post|create/i }),
    ).not.toBeVisible();
  });

  // --- Page Access (fixme: middleware has no role-based page guard) ---

  test.fixme('editor cannot access /dashboard/users', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/users');
    const forbidden = page.getByText(/forbidden|permission|access denied/i);
    await expect(forbidden).toBeVisible({ timeout: 5_000 });
  });

  // --- API-Level RBAC (should pass — backend enforces) ---

  test('API returns 403 when editor accesses sites endpoint', async ({ request }) => {
    const editorToken = await apiLogin(TEST_EDITOR.email, TEST_EDITOR.password);
    const resp = await request.get(`${API_BASE}/api/v1/sites`, {
      headers: { Authorization: `Bearer ${editorToken}` },
    });
    expect(resp.status()).toBe(403);
  });

  test('API returns 403 when viewer creates a post', async ({ request }) => {
    const viewerToken = await apiLogin(TEST_VIEWER.email, TEST_VIEWER.password);
    const resp = await request.post(`${API_BASE}/api/v1/posts`, {
      headers: {
        Authorization: `Bearer ${viewerToken}`,
        'Content-Type': 'application/json',
        'X-Site-Slug': 'e2e-test',
      },
      data: { title: 'Unauthorized Post', content: '<p>test</p>' },
    });
    expect(resp.status()).toBe(403);
  });
});
