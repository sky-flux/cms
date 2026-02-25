import { test, expect } from '@playwright/test';
import { TEST_SUPER } from './helpers/constants';
import { loginViaUI } from './helpers/auth';

test.describe.serial('Authentication Flows', () => {
  test('successful login redirects to dashboard', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await expect(page.getByText(/dashboard/i)).toBeVisible();
  });

  test('wrong password shows error message', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel(/email/i).fill(TEST_SUPER.email);
    await page.getByLabel(/password/i).fill('WrongPassword123!');
    await page.getByRole('button', { name: /sign in/i }).click();

    // Should show toast error, not redirect
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test('unauthenticated access to /dashboard redirects to /login', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('user display name visible after login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    // Header user menu trigger contains avatar with initials
    await expect(page.getByLabel('User menu')).toBeVisible();
  });

  test('logout invalidates session', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);

    // Open user menu and click logout
    await page.getByLabel('User menu').click();
    await page.getByRole('menuitem', { name: /logout/i }).click();

    // Should redirect to login
    await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });

    // Try accessing dashboard — should redirect to login
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('forgot password flow shows success message', async ({ page }) => {
    await page.goto('/forgot-password');
    await page.getByLabel(/email/i).fill(TEST_SUPER.email);
    await page.getByRole('button', { name: /send|reset/i }).click();

    // Anti-enumeration: always shows success page
    await expect(page).toHaveURL(/\/forgot-password\/sent/, { timeout: 5_000 });
  });

  test('invalid email format shows validation error', async ({ page }) => {
    await page.goto('/login');
    await page.getByLabel(/email/i).fill('not-an-email');
    await page.getByLabel(/password/i).fill('SomePassword123!');
    await page.getByRole('button', { name: /sign in/i }).click();

    // Zod validation should prevent submission
    await expect(page).toHaveURL(/\/login/);
  });
});
