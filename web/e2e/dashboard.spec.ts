import { test, expect } from '@playwright/test';
import { TEST_ADMIN } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe('Dashboard Statistics', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
  });

  test('dashboard page loads successfully', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page.getByRole('heading', { name: /dashboard/i })).toBeVisible();
  });

  test('displays posts statistics', async ({ page }) => {
    await page.goto('/dashboard');

    // Wait for content to load (not skeleton)
    await expect(page.getByText('Posts')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Published')).toBeVisible();
    await expect(page.getByText('Drafts')).toBeVisible();
    await expect(page.getByText('Scheduled')).toBeVisible();
  });

  test('displays users statistics', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page.getByText('Users')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Active')).toBeVisible();
    await expect(page.getByText('Inactive')).toBeVisible();
  });

  test('displays comments statistics', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page.getByText('Comments')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Pending')).toBeVisible();
    await expect(page.getByText('Approved')).toBeVisible();
    await expect(page.getByText('Spam')).toBeVisible();
  });

  test('displays media statistics', async ({ page }) => {
    await page.goto('/dashboard');

    await expect(page.getByText('Media Files')).toBeVisible({ timeout: 10_000 });
    await expect(page.getByText('Storage Used')).toBeVisible();
  });

  test('statistics show numeric values', async ({ page }) => {
    await page.goto('/dashboard');

    // Wait for Posts label to appear
    await expect(page.getByText('Posts')).toBeVisible({ timeout: 10_000 });

    // Verify page contains numeric values (not just loading skeletons)
    const bodyText = await page.locator('body').textContent();
    expect(bodyText).toMatch(/\d+/);
  });
});
