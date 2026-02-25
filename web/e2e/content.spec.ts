import { test, expect } from '@playwright/test';
import { TEST_SUPER } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Content Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password);
  });

  // --- Posts ---

  test('posts list page loads', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('create a draft post', async ({ page }) => {
    await page.goto('/dashboard/posts/new');
    await expect(page.getByPlaceholder(/title/i)).toBeVisible({ timeout: 5_000 });

    await page.getByPlaceholder(/title/i).fill('E2E Draft Post');

    const editable = page.locator('[contenteditable="true"]').first();
    if (await editable.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await editable.click();
      await page.keyboard.type('E2E test content paragraph.');
    }

    await page.getByRole('button', { name: /create|save/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('draft post appears in posts list', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await expect(page.getByText('E2E Draft Post')).toBeVisible({ timeout: 5_000 });
  });

  test('edit and publish a post', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await page.getByRole('link', { name: 'E2E Draft Post' }).click();
    await expect(page.getByPlaceholder(/title/i)).toBeVisible({ timeout: 5_000 });

    await page.getByRole('button', { name: /publish/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('delete a post via list actions', async ({ page }) => {
    await page.goto('/dashboard/posts');

    const row = page.getByText('E2E Draft Post').locator('ancestor::tr');
    const actionsBtn = row.locator('button').last();
    await actionsBtn.click();

    await page.getByRole('menuitem', { name: /delete/i }).click();

    const dialog = page.getByRole('alertdialog');
    await expect(dialog).toBeVisible({ timeout: 3_000 });
    await dialog.getByRole('button', { name: /confirm|delete/i }).click();

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  // --- Categories ---

  test('categories page loads', async ({ page }) => {
    await page.goto('/dashboard/categories');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create a category', async ({ page }) => {
    await page.goto('/dashboard/categories');
    await page.getByRole('button', { name: /new|create|add/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByLabel(/name/i).fill('E2E Category');
    await dialog.getByRole('button', { name: /save|create/i }).click();

    await expect(page.getByText('E2E Category')).toBeVisible({ timeout: 5_000 });
  });

  // --- Tags ---

  test('tags page loads', async ({ page }) => {
    await page.goto('/dashboard/tags');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create a tag', async ({ page }) => {
    await page.goto('/dashboard/tags');
    await page.getByRole('button', { name: /new|create|add/i }).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await dialog.getByLabel(/name/i).fill('e2e-tag');
    await dialog.getByRole('button', { name: /save|create/i }).click();

    await expect(page.getByText('e2e-tag')).toBeVisible({ timeout: 5_000 });
  });

  test.fixme('search for a tag', async ({ page }) => {
    await page.goto('/dashboard/tags');
    await page.getByPlaceholder(/search/i).fill('e2e');
    await expect(page.getByText('e2e-tag')).toBeVisible({ timeout: 5_000 });
  });
});
