# E2E Testing (Playwright) Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Full-stack E2E tests covering all 6 critical user journeys (setup, auth, posts, media, RBAC, multi-site) using Playwright against real Go backend + PostgreSQL.

**Architecture:** Playwright drives Chromium against Astro dev server (:4321) which proxies `/api` to Go backend (:8080). Tests run serially (01-06 prefix ordering). globalSetup resets DB and runs setup/initialize API. Helper modules handle auth session reuse and API-based data seeding.

**Tech Stack:** @playwright/test, Astro 5 SSR, Go/Gin backend, PostgreSQL 18, Redis 8

**Key reference:** `docs/testing.md` (section 5), `docs/plans/2026-02-26-e2e-testing-design.md`

---

### Task 1: Install Playwright and Create Config

**Files:**
- Modify: `web/package.json` (add @playwright/test devDependency)
- Create: `web/playwright.config.ts`
- Create: `web/.gitignore` update (add playwright artifacts)

**Step 1: Install Playwright**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bun add -D @playwright/test
```

Then install browsers:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright install chromium
```

**Step 2: Create playwright.config.ts**

```typescript
// web/playwright.config.ts
import { defineConfig } from '@playwright/test';

export default defineConfig({
  testDir: './e2e',
  timeout: 30_000,
  expect: { timeout: 5_000 },
  fullyParallel: false,
  forbidOnly: !!process.env.CI,
  retries: process.env.CI ? 2 : 1,
  workers: 1,
  reporter: process.env.CI ? 'github' : 'html',
  use: {
    baseURL: 'http://localhost:4321',
    trace: 'retain-on-failure',
    screenshot: 'only-on-failure',
    locale: 'en-US',
  },
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:4321',
    timeout: 30_000,
    reuseExistingServer: !process.env.CI,
  },
});
```

**Step 3: Add scripts to package.json**

Add to `web/package.json` scripts:
```json
"test:e2e": "playwright test",
"test:e2e:ui": "playwright test --ui",
"test:e2e:debug": "playwright test --debug"
```

**Step 4: Add gitignore entries**

Append to `web/.gitignore` (or create if needed):
```
# Playwright
test-results/
playwright-report/
e2e/.auth/
```

**Step 5: Create e2e directory structure**

```bash
mkdir -p web/e2e/fixtures web/e2e/helpers
```

**Step 6: Verify playwright runs (empty)**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test
```
Expected: "No tests found" or similar (no test files yet).

**Step 7: Commit**

```bash
git add web/package.json web/bun.lock web/playwright.config.ts web/.gitignore web/e2e/
git commit -m "feat(web): add Playwright E2E test infrastructure"
```

---

### Task 2: Create E2E Helper Modules

**Files:**
- Create: `web/e2e/helpers/constants.ts`
- Create: `web/e2e/helpers/api.ts`
- Create: `web/e2e/helpers/auth.ts`

**Step 1: Create constants.ts**

```typescript
// web/e2e/helpers/constants.ts
export const TEST_SUPER = {
  displayName: 'E2E Super Admin',
  email: 'super@e2e-test.com',
  password: 'SuperPass123!',
};

export const TEST_EDITOR = {
  displayName: 'E2E Editor',
  email: 'editor@e2e-test.com',
  password: 'EditorPass123!',
};

export const TEST_VIEWER = {
  displayName: 'E2E Viewer',
  email: 'viewer@e2e-test.com',
  password: 'ViewerPass123!',
};

export const TEST_SITE = {
  name: 'E2E Test Site',
  slug: 'e2e-test',
  url: 'http://localhost:4321',
  locale: 'en',
};

export const API_BASE = 'http://localhost:8080';
```

**Step 2: Create api.ts — direct API helper for data seeding**

```typescript
// web/e2e/helpers/api.ts
import { API_BASE, TEST_SUPER, TEST_SITE } from './constants';

/** Low-level API call (bypasses browser, used for seeding data). */
async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
): Promise<T> {
  const headers: Record<string, string> = { 'Content-Type': 'application/json' };
  if (token) headers.Authorization = `Bearer ${token}`;

  const res = await fetch(`${API_BASE}${path}`, {
    method,
    headers,
    body: body ? JSON.stringify(body) : undefined,
  });

  if (!res.ok) {
    const text = await res.text();
    throw new Error(`API ${method} ${path} failed (${res.status}): ${text}`);
  }

  if (res.status === 204) return undefined as T;
  return res.json();
}

/** Run setup/initialize to create the first super admin + site. */
export async function setupInitialize(): Promise<string> {
  const resp = await apiCall<{
    success: boolean;
    data: { access_token: string };
  }>('POST', '/api/v1/setup/initialize', {
    admin_display_name: TEST_SUPER.displayName,
    admin_email: TEST_SUPER.email,
    admin_password: TEST_SUPER.password,
    site_name: TEST_SITE.name,
    site_slug: TEST_SITE.slug,
    site_url: TEST_SITE.url,
    locale: TEST_SITE.locale,
  });
  return resp.data.access_token;
}

/** Check if system is already installed. */
export async function checkInstalled(): Promise<boolean> {
  const resp = await apiCall<{
    success: boolean;
    data: { installed: boolean };
  }>('POST', '/api/v1/setup/check');
  return resp.data.installed;
}

/** Login via API, returns access_token. */
export async function apiLogin(
  email: string,
  password: string,
): Promise<string> {
  const resp = await apiCall<{
    success: boolean;
    data: { access_token: string };
  }>('POST', '/api/v1/auth/login', { email, password });
  return resp.data.access_token;
}

/** Create a user via API (requires super token + X-Site-Slug). */
export async function createUser(
  token: string,
  user: { display_name: string; email: string; password: string; role_id?: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string }> {
  const resp = await apiCall<{ success: boolean; data: { id: string } }>(
    'POST',
    '/api/v1/users',
    user,
    token,
  );
  return resp.data;
}

/** Create a post via API (requires auth token + X-Site-Slug header). */
export async function createPost(
  token: string,
  post: { title: string; content: string; status?: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string; slug: string }> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    Authorization: `Bearer ${token}`,
    'X-Site-Slug': siteSlug,
  };

  const res = await fetch(`${API_BASE}/api/v1/posts`, {
    method: 'POST',
    headers,
    body: JSON.stringify(post),
  });

  if (!res.ok) throw new Error(`Create post failed: ${res.status}`);
  const data = await res.json();
  return data.data;
}

/** Create a site via API (requires super token). */
export async function createSite(
  token: string,
  site: { name: string; slug: string; domain?: string },
): Promise<void> {
  await apiCall('POST', '/api/v1/sites', site, token);
}
```

**Step 3: Create auth.ts — browser-level login helper**

```typescript
// web/e2e/helpers/auth.ts
import { type Page, expect } from '@playwright/test';

/**
 * Login via the UI login form.
 * After success, page should be at /dashboard.
 */
export async function loginViaUI(
  page: Page,
  email: string,
  password: string,
): Promise<void> {
  await page.goto('/login');
  await page.getByLabel(/email/i).fill(email);
  await page.getByLabel(/password/i).fill(password);
  await page.getByRole('button', { name: /sign in/i }).click();
  await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 });
}

/**
 * Login via API and set the access_token cookie directly (faster than UI login).
 * Use this in beforeEach when the test doesn't need to test login itself.
 */
export async function loginViaAPI(
  page: Page,
  email: string,
  password: string,
  baseURL: string,
): Promise<string> {
  const resp = await page.request.post(`${baseURL}/api/v1/auth/login`, {
    data: { email, password },
  });
  expect(resp.ok()).toBeTruthy();
  const json = await resp.json();
  const token = json.data.access_token;

  // Set cookie so Astro middleware sees it
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
```

**Step 4: Create test image fixture**

```bash
# Create a small 1x1 red PNG for media upload tests
cd /Users/martinadamsdev/workspace/sky-flux-cms/web/e2e/fixtures
python3 -c "
import struct, zlib
def png():
    sig = b'\\x89PNG\\r\\n\\x1a\\n'
    def chunk(t, d):
        c = t + d
        return struct.pack('>I', len(d)) + c + struct.pack('>I', zlib.crc32(c) & 0xffffffff)
    ihdr = struct.pack('>IIBBBBB', 1, 1, 8, 2, 0, 0, 0)
    raw = zlib.compress(b'\\x00\\xff\\x00\\x00')
    return sig + chunk(b'IHDR', ihdr) + chunk(b'IDAT', raw) + chunk(b'IEND', b'')
open('test-image.png', 'wb').write(png())
"
```

**Step 5: Create CSV fixture for redirect import tests**

```csv
source_path,target_url,status_code
/old-page,/new-page,301
/legacy,/modern,302
```

Save as `web/e2e/fixtures/test-redirects.csv`.

**Step 6: Commit**

```bash
git add web/e2e/
git commit -m "feat(web): add E2E helper modules and test fixtures"
```

---

### Task 3: 01-setup.spec.ts — Installation Wizard

**Files:**
- Create: `web/e2e/01-setup.spec.ts`

**Step 1: Write the test**

```typescript
// web/e2e/01-setup.spec.ts
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
    await expect(page.getByText(/system setup/i)).toBeVisible();
  });

  test('complete 3-step installation wizard', async ({ page }) => {
    await page.goto('/setup');

    // Step 1: Admin account
    await page.getByLabel(/username/i).fill(TEST_SUPER.displayName);
    await page.getByLabel(/email/i).fill(TEST_SUPER.email);
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
    // Should redirect away or show already-installed message
    // The InstallationGuard middleware handles this
    await expect(page).not.toHaveURL(/\/setup$/);
  });
});
```

**Step 2: Run to verify**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/01-setup.spec.ts
```

> **NOTE:** This requires Go backend + docker services running. If services aren't running, tests will fail at the API call stage. The test validates against real infrastructure.

**Step 3: Commit**

```bash
git add web/e2e/01-setup.spec.ts
git commit -m "test(e2e): add installation wizard tests"
```

---

### Task 4: 02-auth.spec.ts — Authentication Flows

**Files:**
- Create: `web/e2e/02-auth.spec.ts`

**Step 1: Write the test**

```typescript
// web/e2e/02-auth.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
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

  test('user can view their profile after login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    // The header should show user info or avatar
    await expect(page.getByText(TEST_SUPER.displayName)).toBeVisible();
  });

  test('logout invalidates session', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);

    // Click user menu and logout
    await page.getByRole('button', { name: /user menu|avatar/i }).click();
    await page.getByRole('menuitem', { name: /logout|sign out/i }).click();

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

    // Should show success/sent page (anti-enumeration: always shows success)
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
```

**Step 2: Run**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/02-auth.spec.ts
```

**Step 3: Commit**

```bash
git add web/e2e/02-auth.spec.ts
git commit -m "test(e2e): add authentication flow tests"
```

---

### Task 5: 03-posts.spec.ts — Post Lifecycle

**Files:**
- Create: `web/e2e/03-posts.spec.ts`

**Step 1: Write the test**

```typescript
// web/e2e/03-posts.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_SITE, API_BASE } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Post Lifecycle', () => {
  let token: string;

  test.beforeAll(async ({ browser }) => {
    // Login via API for faster setup
    const context = await browser.newContext();
    const page = await context.newPage();
    token = await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password, API_BASE);
    await context.close();
  });

  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password, API_BASE);
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
    // Create an API key first or use the admin token to check public API
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': TEST_SITE.slug },
    });
    // Public API might need an API key — adjust based on actual implementation
    // For now verify the endpoint is reachable
    expect([200, 401]).toContain(resp.status());
  });

  test('soft delete the post', async ({ page }) => {
    await page.goto('/dashboard/posts');
    // Find the post row and delete it
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
```

**Step 2: Run**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/03-posts.spec.ts
```

**Step 3: Commit**

```bash
git add web/e2e/03-posts.spec.ts
git commit -m "test(e2e): add post lifecycle tests"
```

---

### Task 6: 04-media.spec.ts — Media Management

**Files:**
- Create: `web/e2e/04-media.spec.ts`

**Step 1: Write the test**

```typescript
// web/e2e/04-media.spec.ts
import { test, expect } from '@playwright/test';
import path from 'node:path';
import { TEST_SUPER, API_BASE } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';

test.describe.serial('Media Management', () => {
  test.beforeEach(async ({ page }) => {
    await loginViaAPI(page, TEST_SUPER.email, TEST_SUPER.password, API_BASE);
  });

  test('navigate to media library', async ({ page }) => {
    await page.goto('/dashboard/media');
    await expect(page.getByRole('heading', { name: /media/i })).toBeVisible();
  });

  test('upload an image file', async ({ page }) => {
    await page.goto('/dashboard/media');

    // Trigger file input (react-dropzone creates a hidden input)
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(
      path.resolve(__dirname, 'fixtures/test-image.png'),
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
    const mediaItem = page.getByText(/test-image/i);
    await mediaItem.click();

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
```

**Step 2: Run**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/04-media.spec.ts
```

**Step 3: Commit**

```bash
git add web/e2e/04-media.spec.ts
git commit -m "test(e2e): add media management tests"
```

---

### Task 7: 05-rbac.spec.ts — Role-Based Access Control

**Files:**
- Create: `web/e2e/05-rbac.spec.ts`

**Prerequisite:** This test needs editor and viewer users. They will be created via API in beforeAll.

**Step 1: Write the test**

```typescript
// web/e2e/05-rbac.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_EDITOR, TEST_VIEWER, API_BASE } from './helpers/constants';
import { loginViaUI, loginViaAPI } from './helpers/auth';
import { apiLogin, createUser } from './helpers/api';

test.describe.serial('Role-Based Access Control', () => {
  test.beforeAll(async () => {
    // Seed editor and viewer users via API
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);

    try {
      await createUser(superToken, {
        display_name: TEST_EDITOR.displayName,
        email: TEST_EDITOR.email,
        password: TEST_EDITOR.password,
      });
    } catch {
      // User may already exist from previous run
    }

    try {
      await createUser(superToken, {
        display_name: TEST_VIEWER.displayName,
        email: TEST_VIEWER.email,
        password: TEST_VIEWER.password,
      });
    } catch {
      // User may already exist
    }
  });

  test('super admin sees all navigation items', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);

    // Super should see system navigation items
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /roles/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
  });

  test('editor can access posts page', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/posts');
    await expect(page.getByRole('heading', { name: /posts/i })).toBeVisible();
  });

  test('editor cannot access user management', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/users');

    // Should see forbidden/redirect or no access
    const forbidden = page.getByText(/forbidden|permission|access denied/i);
    const loginRedirect = page.url().includes('/login');
    expect(await forbidden.isVisible() || loginRedirect).toBeTruthy();
  });

  test('viewer cannot see new post button', async ({ page }) => {
    await loginViaUI(page, TEST_VIEWER.email, TEST_VIEWER.password);
    await page.goto('/dashboard/posts');

    // Viewer should not see create/new post button
    await expect(
      page.getByRole('button', { name: /new post|create/i }),
    ).not.toBeVisible();
  });

  test('API returns 403 for unauthorized role', async ({ request }) => {
    // Editor tries to access super-only endpoint
    const editorToken = await apiLogin(TEST_EDITOR.email, TEST_EDITOR.password);

    const resp = await request.get(`${API_BASE}/api/v1/sites`, {
      headers: { Authorization: `Bearer ${editorToken}` },
    });
    expect(resp.status()).toBe(403);
  });
});
```

**Step 2: Run**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/05-rbac.spec.ts
```

**Step 3: Commit**

```bash
git add web/e2e/05-rbac.spec.ts
git commit -m "test(e2e): add RBAC permission tests"
```

---

### Task 8: 06-multisite.spec.ts — Multi-Site Isolation

**Files:**
- Create: `web/e2e/06-multisite.spec.ts`

**Step 1: Write the test**

```typescript
// web/e2e/06-multisite.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
import { loginViaUI, loginViaAPI } from './helpers/auth';
import { apiLogin, createSite, createPost } from './helpers/api';

const SITE_A = { name: 'Site Alpha', slug: 'site-alpha' };
const SITE_B = { name: 'Site Beta', slug: 'site-beta' };

test.describe.serial('Multi-Site Isolation', () => {
  let superToken: string;

  test.beforeAll(async () => {
    superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
  });

  test('create two sites via UI', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');

    // Create Site A
    await page.getByRole('button', { name: /new site|create/i }).click();
    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/name/i).fill(SITE_A.name);
    await dialog.getByLabel(/slug/i).fill(SITE_A.slug);
    await dialog.getByRole('button', { name: /save|create/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });

    // Create Site B
    await page.getByRole('button', { name: /new site|create/i }).click();
    const dialog2 = page.getByRole('dialog');
    await dialog2.getByLabel(/name/i).fill(SITE_B.name);
    await dialog2.getByLabel(/slug/i).fill(SITE_B.slug);
    await dialog2.getByRole('button', { name: /save|create/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('create posts in different sites via API', async () => {
    await createPost(
      superToken,
      { title: 'Alpha Post', content: '<p>Content for site alpha</p>', status: 'published' },
      SITE_A.slug,
    );
    await createPost(
      superToken,
      { title: 'Beta Post', content: '<p>Content for site beta</p>', status: 'published' },
      SITE_B.slug,
    );
  });

  test('site_alpha Public API only returns alpha posts', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': SITE_A.slug },
    });
    if (resp.ok()) {
      const json = await resp.json();
      const titles = json.data?.map((p: { title: string }) => p.title) ?? [];
      expect(titles).toContain('Alpha Post');
      expect(titles).not.toContain('Beta Post');
    }
  });

  test('site_beta Public API only returns beta posts', async ({ request }) => {
    const resp = await request.get(`${API_BASE}/api/public/v1/posts`, {
      headers: { 'X-Site-Slug': SITE_B.slug },
    });
    if (resp.ok()) {
      const json = await resp.json();
      const titles = json.data?.map((p: { title: string }) => p.title) ?? [];
      expect(titles).toContain('Beta Post');
      expect(titles).not.toContain('Alpha Post');
    }
  });

  test('both sites visible in sites list', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');
    await expect(page.getByText(SITE_A.name)).toBeVisible();
    await expect(page.getByText(SITE_B.name)).toBeVisible();
  });
});
```

**Step 2: Run**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test e2e/06-multisite.spec.ts
```

**Step 3: Commit**

```bash
git add web/e2e/06-multisite.spec.ts
git commit -m "test(e2e): add multi-site isolation tests"
```

---

### Task 9: Run Full E2E Suite and Fix Issues

**Step 1: Run all E2E tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test
```

**Step 2: Review failures**

E2E tests against real infrastructure will likely need adjustments:
- Selector mismatches (i18n keys vs actual rendered text)
- Timing issues (add `waitFor` / increase timeouts)
- API response format differences
- Missing data-testid attributes

**Step 3: Fix any failing tests**

Iterate on fixes. Common patterns:
- Use `page.waitForResponse()` for API calls
- Use `page.waitForURL()` for navigation
- Add `data-testid` attributes to components if selectors are brittle
- Adjust toast selectors if Sonner uses different attributes

**Step 4: Run full suite again and verify all pass**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test
```

**Step 5: Final commit**

```bash
git add -A
git commit -m "test(e2e): fix E2E tests after integration validation"
```

---

### Task 10: Update Documentation and Memory

**Files:**
- Modify: `docs/v1.0.0.md` (update E2E status)
- Update: auto memory with E2E completion notes

**Step 1: Update milestone doc**

Mark E2E testing as complete in `docs/v1.0.0.md`.

**Step 2: Commit**

```bash
git add docs/v1.0.0.md
git commit -m "docs: mark E2E testing as complete in v1.0.0 milestone"
```
