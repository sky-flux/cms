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
    await expect(page.getByText(/welcome to sky flux/i)).toBeVisible();
  });

  test('complete 3-step installation wizard', async ({ page }) => {
    await page.goto('/setup');

    // Step 1: Admin account
    await page.getByLabel(/display name/i).fill(TEST_SUPER.displayName);
    await page.getByLabel(/admin email/i).fill(TEST_SUPER.email);
    await page.getByLabel('Password', { exact: true }).fill(TEST_SUPER.password);
    await page.getByLabel(/confirm/i).fill(TEST_SUPER.password);
    await page.getByRole('button', { name: /next/i }).click();

    // Step 2: Site info
    await page.getByLabel(/site name/i).fill(TEST_SITE.name);
    await page.getByLabel(/site slug/i).fill(TEST_SITE.slug);
    await page.getByLabel(/site url/i).fill(TEST_SITE.url);
    await page.getByRole('button', { name: /next/i }).click();

    // Step 3: Review & Install
    await expect(page.getByText(TEST_SUPER.email)).toBeVisible();
    await expect(page.getByText(TEST_SITE.name)).toBeVisible();
    await page.getByRole('button', { name: /install/i }).click();

    // Should redirect to setup complete page
    await expect(page).toHaveURL(/\/setup\/complete/, { timeout: 15_000 });
  });

  test('setup check now returns installed', async ({ request }) => {
    const resp = await request.post(`${API_BASE}/api/v1/setup/check`);
    const json = await resp.json();
    expect(json.data.installed).toBe(true);
  });

  test('repeat installation attempt is rejected', async ({ page }) => {
    await page.goto('/setup');
    // InstallationGuard middleware should redirect away
    await expect(page).not.toHaveURL(/\/setup$/);
  });
});
