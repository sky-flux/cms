import { test, expect } from '@playwright/test';
import path from 'node:path';
import { TEST_SUPER } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Media Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password);
  });

  test('media library page loads', async ({ page }) => {
    await page.goto('/dashboard/media');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test('upload an image file', async ({ page }) => {
    await page.goto('/dashboard/media');

    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(
      path.resolve(import.meta.dirname, 'fixtures/test-image.png'),
    );

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 10_000 });
  });

  test('uploaded image appears in library', async ({ page }) => {
    await page.goto('/dashboard/media');
    await expect(page.getByText(/test-image/i)).toBeVisible({ timeout: 5_000 });
  });

  test('media detail dialog shows file info', async ({ page }) => {
    await page.goto('/dashboard/media');
    await page.getByText(/test-image/i).click();

    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible({ timeout: 3_000 });
    await expect(dialog.getByText(/png/i)).toBeVisible();
  });

  test('delete media file', async ({ page }) => {
    await page.goto('/dashboard/media');
    await page.getByText(/test-image/i).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByRole('button', { name: /delete/i }).click();

    const confirm = page.getByRole('alertdialog');
    if (await confirm.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await confirm.getByRole('button', { name: /confirm|delete/i }).click();
    }

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });
});
