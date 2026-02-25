import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_ADMIN } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';
import { apiLogin, seedTestUsers } from './helpers/api';

test.describe.serial('System Management', () => {
  test.beforeAll(async () => {
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
    await seedTestUsers(superToken, [TEST_ADMIN]);
  });

  // --- Users ---

  test.fixme('users list page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/users');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create a new user with editor role', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/users');
    await page.getByRole('button', { name: /new|create|add/i }).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/email/i).fill('new-user@e2e-test.com');
    await dialog.getByLabel(/display name/i).fill('New E2E User');
    await dialog.getByLabel(/password/i).fill('NewUser123!');
    await dialog.getByLabel(/role/i).click();
    await page.getByRole('option', { name: /editor/i }).click();
    await dialog.getByRole('button', { name: /save|create/i }).click();

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  // --- Roles ---

  test.fixme('roles list page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/roles');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    await expect(page.getByText(/super/i)).toBeVisible();
    await expect(page.getByText(/admin/i)).toBeVisible();
    await expect(page.getByText(/editor/i)).toBeVisible();
    await expect(page.getByText(/viewer/i)).toBeVisible();
  });

  test.fixme('view role permissions', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/roles');
    await page.getByText(/editor/i).click();
    await expect(page.getByText(/permissions/i)).toBeVisible();
  });

  // --- Settings ---

  test.fixme('settings page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/settings');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('update a site setting', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/settings');
    const input = page.getByLabel(/site name|site title/i);
    await input.clear();
    await input.fill('Updated Site Name');
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  // --- API Keys ---

  test.fixme('api keys page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/api-keys');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create an API key', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/api-keys');
    await page.getByRole('button', { name: /new|create|generate/i }).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/name|description/i).fill('E2E Test Key');
    await dialog.getByRole('button', { name: /create|generate/i }).click();

    await expect(page.getByText(/key.*created|copy.*key/i)).toBeVisible({ timeout: 5_000 });
  });

  // --- Audit Logs ---

  test.fixme('audit logs page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/audit');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  // --- Comments ---

  test.fixme('comments page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/comments');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  // --- Menus ---

  test.fixme('menus page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/menus');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  // --- Redirects ---

  test.fixme('redirects page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/redirects');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });
});
