import { type Page, expect } from '@playwright/test';
import { API_BASE } from './constants';

export async function loginViaUI(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.locator('#email').fill(email);
  await page.locator('#password').fill(password);
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 });
}

export async function loginViaAPI(page: Page, email: string, password: string): Promise<string> {
  const resp = await page.request.post(`${API_BASE}/api/v1/auth/login`, {
    data: { email, password },
  });
  expect(resp.ok()).toBeTruthy();
  const json = await resp.json();
  const token = json.data.access_token;

  await page.context().addCookies([
    {
      name: 'access_token',
      value: token,
      domain: 'localhost',
      path: '/',
    },
  ]);

  return token;
}
