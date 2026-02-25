import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_EDITOR, TEST_VIEWER, API_BASE } from './helpers/constants';
import { loginViaUI } from './helpers/auth';
import { apiLogin, createUser } from './helpers/api';

test.describe.serial('Role-Based Access Control', () => {
  test.beforeAll(async () => {
    // Seed editor and viewer users via API
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);

    try {
      await createUser(superToken, {
        display_name: TEST_EDITOR.displayName,
        email: TEST_EDITOR.email,
        password: TEST_EDITOR.password,
      });
    } catch {
      // User may already exist from previous run
    }

    try {
      await createUser(superToken, {
        display_name: TEST_VIEWER.displayName,
        email: TEST_VIEWER.email,
        password: TEST_VIEWER.password,
      });
    } catch {
      // User may already exist
    }
  });

  test('super admin sees all navigation items', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);

    // Super should see system navigation items
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /roles/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
  });

  test('editor can access posts page', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/posts');
    await expect(page.getByRole('heading', { name: /posts/i })).toBeVisible();
  });

  test('editor cannot access user management', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/users');

    // Should see forbidden/redirect or no access
    const forbidden = page.getByText(/forbidden|permission|access denied/i);
    const loginRedirect = page.url().includes('/login');
    expect(await forbidden.isVisible() || loginRedirect).toBeTruthy();
  });

  test('viewer cannot see new post button', async ({ page }) => {
    await loginViaUI(page, TEST_VIEWER.email, TEST_VIEWER.password);
    await page.goto('/dashboard/posts');

    // Viewer should not see create/new post button
    await expect(
      page.getByRole('button', { name: /new post|create/i }),
    ).not.toBeVisible();
  });

  test('API returns 403 for unauthorized role', async ({ request }) => {
    // Editor tries to access super-only endpoint
    const editorToken = await apiLogin(TEST_EDITOR.email, TEST_EDITOR.password);

    const resp = await request.get(`${API_BASE}/api/v1/sites`, {
      headers: { Authorization: `Bearer ${editorToken}` },
    });
    expect(resp.status()).toBe(403);
  });
});
