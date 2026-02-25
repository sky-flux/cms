import { test, expect } from '@playwright/test';
import { TEST_SUPER } from './helpers/constants';
import { loginViaUI } from './helpers/auth';

test.describe.serial('Authentication Flows', () => {
  test('successful login redirects to dashboard', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await expect(page).toHaveURL(/\/dashboard/);
  });

  test('wrong password shows error toast', async ({ page }) => {
    await page.goto('/login');
    await page.locator('#email').fill(TEST_SUPER.email);
    await page.locator('#password').fill('WrongPassword123!');
    await page.locator('button[type="submit"]').click();

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test('unauthenticated /dashboard redirects to /login', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('user menu is visible after login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await expect(page.getByLabel('User menu')).toBeVisible();
  });

  test('logout redirects to login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.getByLabel('User menu').click();
    await page.getByRole('menuitem', { name: /logout/i }).click();
    await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });
  });

  test('after logout, dashboard is inaccessible', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('forgot password always shows sent page', async ({ page }) => {
    await page.goto('/forgot-password');
    await page.locator('#email').fill(TEST_SUPER.email);
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/\/forgot-password\/sent/, { timeout: 5_000 });
  });

  test('login form validates email format', async ({ page }) => {
    await page.goto('/login');
    await page.locator('#email').fill('not-an-email');
    await page.locator('#password').fill('SomePassword123!');
    await page.locator('button[type="submit"]').click();
    await expect(page).toHaveURL(/\/login/);
  });
});
