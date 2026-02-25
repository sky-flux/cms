import { test, expect } from '@playwright/test';
import path from 'node:path';
import { TEST_SUPER } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Media Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password);
  });

  test('navigate to media library', async ({ page }) => {
    await page.goto('/dashboard/media');
    await expect(page.getByRole('heading', { name: /media/i })).toBeVisible();
  });

  test('upload an image file', async ({ page }) => {
    await page.goto('/dashboard/media');

    // react-dropzone creates a hidden file input
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(
      path.resolve(import.meta.dirname, 'fixtures/test-image.png'),
    );

    // Wait for upload success toast
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 10_000 });

    // Image should appear in the library
    await expect(page.getByText(/test-image/i)).toBeVisible();
  });

  test('media detail dialog shows file info', async ({ page }) => {
    await page.goto('/dashboard/media');

    // Click on the uploaded image
    await page.getByText(/test-image/i).click();

    // Dialog should show file details
    const dialog = page.getByRole('dialog');
    await expect(dialog).toBeVisible();
    await expect(dialog.getByText(/png/i)).toBeVisible();
  });

  test('delete media file', async ({ page }) => {
    await page.goto('/dashboard/media');

    // Select the media item
    await page.getByText(/test-image/i).click();

    // Click delete in the detail dialog
    const dialog = page.getByRole('dialog');
    await dialog.getByRole('button', { name: /delete/i }).click();

    // Confirm deletion
    const confirmDialog = page.getByRole('alertdialog');
    if (await confirmDialog.isVisible()) {
      await confirmDialog.getByRole('button', { name: /confirm|delete/i }).click();
    }

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });
});
