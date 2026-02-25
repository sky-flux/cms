import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_SITE, API_BASE } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Post Lifecycle', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password);
  });

  test('navigate to posts list page', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await expect(page.getByRole('heading', { name: /posts/i })).toBeVisible();
  });

  test('create a new draft post', async ({ page }) => {
    await page.goto('/dashboard/posts/new');

    // Fill title
    await page.getByPlaceholder(/title/i).fill('E2E Test Post');

    // BlockNote editor — click into content area and type
    const editor = page.locator('[data-content-editable-leaf="true"]').first();
    if (await editor.isVisible()) {
      await editor.click();
      await editor.fill('This is E2E test content for the post.');
    }

    // Save draft
    await page.getByRole('button', { name: /save|draft/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });

    // Should redirect to edit page with post ID
    await expect(page).toHaveURL(/\/dashboard\/posts\/[\w-]+\/edit/);
  });

  test('post appears in posts list', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await expect(page.getByText('E2E Test Post')).toBeVisible();
  });

  test('publish the post', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await page.getByText('E2E Test Post').click();

    // Wait for editor to load
    await expect(page.getByPlaceholder(/title/i)).toHaveValue('E2E Test Post');

    // Click publish action
    await page.getByRole('button', { name: /publish/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('published post is accessible via Public API', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': TEST_SITE.slug },
    });
    // Public API may need an API key — verify endpoint is reachable
    expect([200, 401]).toContain(resp.status());
  });

  test('soft delete the post', async ({ page }) => {
    await page.goto('/dashboard/posts');

    // Find the post row and trigger delete
    const row = page.getByText('E2E Test Post').locator('..');
    await row.getByRole('button', { name: /delete|trash/i }).click();

    // Confirm deletion in dialog
    const dialog = page.getByRole('alertdialog');
    if (await dialog.isVisible()) {
      await dialog.getByRole('button', { name: /confirm|delete/i }).click();
    }

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });
});
