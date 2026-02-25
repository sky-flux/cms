import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_SITE, API_BASE } from './helpers/constants';

test.describe.serial('Installation Wizard', () => {
  test('setup check returns not installed', async ({ request }) => {
    const resp = await request.post(`${API_BASE}/api/v1/setup/check`);
    expect(resp.ok()).toBeTruthy();
    const json = await resp.json();
    expect(json.data.installed).toBe(false);
  });

  test('navigate to /setup shows wizard', async ({ page }) => {
    await page.goto('/setup');
    await expect(page.locator('#admin_display_name')).toBeVisible();
  });

  test('complete 3-step installation wizard', async ({ page }) => {
    await page.goto('/setup');

    // Step 1: Admin account
    await page.locator('#admin_display_name').fill(TEST_SUPER.displayName);
    await page.locator('#admin_email').fill(TEST_SUPER.email);
    await page.locator('#password').fill(TEST_SUPER.password);
    await page.locator('#confirmPassword').fill(TEST_SUPER.password);
    await page.locator('button[type="submit"]').click();

    // Step 2: Site info
    await expect(page.locator('#site_name')).toBeVisible({ timeout: 5_000 });
    await page.locator('#site_name').fill(TEST_SITE.name);
    await page.locator('#site_slug').fill(TEST_SITE.slug);
    await page.locator('#site_url').fill(TEST_SITE.url);
    await page.locator('button[type="submit"]').click();

    // Step 3: Review & Install
    await expect(page.getByText(TEST_SUPER.email)).toBeVisible({ timeout: 5_000 });
    await expect(page.getByText(TEST_SITE.name)).toBeVisible();
    await page.locator('button[type="submit"]').click();

    await expect(page).toHaveURL(/\/setup\/complete/, { timeout: 15_000 });
  });

  test('setup check now returns installed', async ({ request }) => {
    const resp = await request.post(`${API_BASE}/api/v1/setup/check`);
    const json = await resp.json();
    expect(json.data.installed).toBe(true);
  });

  test('repeat installation attempt is rejected', async ({ page }) => {
    await page.goto('/setup');
    await expect(page).not.toHaveURL(/\/setup$/);
  });
});
