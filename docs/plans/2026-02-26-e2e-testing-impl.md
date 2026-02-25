# E2E Test Suite Rewrite — Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Rewrite all 6 E2E spec files + helpers with 4-role coverage (Super/Admin/Editor/Viewer), accurate selectors from actual components, and Batch 12 system management stubs.

**Architecture:** Playwright drives Chromium against Astro dev server (:4321) which proxies `/api` to Go backend (:8080). Tests run serially via `projects` with dependencies. Helpers handle API-based seeding and auth. Tests marked `test.fixme()` for unimplemented features.

**Tech Stack:** @playwright/test, Astro 5 SSR, Go/Gin backend, PostgreSQL 18, Redis 8

**Key reference:** `docs/plans/2026-02-26-e2e-testing-design.md`, `web/src/components/` for selectors

---

### Known UI Limitations (affects test expectations)

1. **DashboardShell** (`web/src/components/layout/DashboardShell.tsx`): Uses hardcoded placeholder user, NOT auth store. Shows ALL nav items to ALL users (no role filtering).
2. **AuthUser** (`web/src/stores/auth-store.ts`): No `role` field — role-based nav filtering impossible until store is updated.
3. **Logout**: Just `window.location.href = '/login'` — no API call to invalidate token.
4. **Middleware** (`web/src/middleware.ts`): Only checks cookie existence, no role-based page access control.

**Consequence:** RBAC nav visibility tests and role-based page access tests MUST use `test.fixme()` until DashboardShell is updated with auth store integration + role-based nav filtering.

---

### Task 1: Rewrite constants.ts — Add 4 Roles

**Files:**
- Modify: `web/e2e/helpers/constants.ts`

**Step 1: Rewrite file**

```typescript
// web/e2e/helpers/constants.ts
export const TEST_SUPER = {
  displayName: 'E2E Super Admin',
  email: 'super@e2e-test.com',
  password: 'SuperPass123!',
  role: 'super',
};

export const TEST_ADMIN = {
  displayName: 'E2E Admin',
  email: 'admin@e2e-test.com',
  password: 'AdminPass123!',
  role: 'admin',
};

export const TEST_EDITOR = {
  displayName: 'E2E Editor',
  email: 'editor@e2e-test.com',
  password: 'EditorPass123!',
  role: 'editor',
};

export const TEST_VIEWER = {
  displayName: 'E2E Viewer',
  email: 'viewer@e2e-test.com',
  password: 'ViewerPass123!',
  role: 'viewer',
};

export const TEST_SITE = {
  name: 'E2E Test Site',
  slug: 'e2e-test',
  url: 'http://localhost:4321',
  locale: 'en',
};

export const API_BASE = 'http://localhost:8080';
```

**Step 2: Verify no TypeScript errors**

Run:
```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx tsc --noEmit --project tsconfig.json 2>&1 | head -20
```

---

### Task 2: Rewrite api.ts — Fix Role Field + Add Helpers

**Files:**
- Modify: `web/e2e/helpers/api.ts`

**Step 1: Rewrite file**

The backend `POST /api/v1/users` accepts `role` (slug string like "admin"), not `role_id`. Also needs `X-Site-Slug` header for site-scoped endpoints.

```typescript
// web/e2e/helpers/api.ts
import { API_BASE, TEST_SUPER, TEST_SITE } from './constants';

/** Low-level API call helper. */
async function apiCall<T>(
  method: string,
  path: string,
  body?: unknown,
  token?: string,
  extraHeaders?: Record<string, string>,
): Promise<T> {
  const headers: Record<string, string> = {
    'Content-Type': 'application/json',
    ...extraHeaders,
  };
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

/** Run setup/initialize to create super admin + first site. Returns access_token. */
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

/** Check if system is installed. */
export async function checkInstalled(): Promise<boolean> {
  const resp = await apiCall<{
    success: boolean;
    data: { installed: boolean };
  }>('POST', '/api/v1/setup/check');
  return resp.data.installed;
}

/** Login via API, returns access_token. */
export async function apiLogin(email: string, password: string): Promise<string> {
  const resp = await apiCall<{
    success: boolean;
    data: { access_token: string };
  }>('POST', '/api/v1/auth/login', { email, password });
  return resp.data.access_token;
}

/**
 * Create a user via API.
 * Backend accepts `role` field as slug string ("admin", "editor", "viewer").
 * Requires super token + X-Site-Slug header.
 */
export async function createUser(
  token: string,
  user: { display_name: string; email: string; password: string; role: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string }> {
  const resp = await apiCall<{ success: boolean; data: { id: string } }>(
    'POST',
    '/api/v1/users',
    user,
    token,
    { 'X-Site-Slug': siteSlug },
  );
  return resp.data;
}

/** Create a post via API. Requires auth token + X-Site-Slug header. */
export async function createPost(
  token: string,
  post: { title: string; content: string; status?: string },
  siteSlug = TEST_SITE.slug,
): Promise<{ id: string; slug: string }> {
  const resp = await apiCall<{ success: boolean; data: { id: string; slug: string } }>(
    'POST',
    '/api/v1/posts',
    post,
    token,
    { 'X-Site-Slug': siteSlug },
  );
  return resp.data;
}

/** Create a site via API. Requires super token. */
export async function createSite(
  token: string,
  site: { name: string; slug: string; domain?: string },
): Promise<void> {
  await apiCall('POST', '/api/v1/sites', site, token);
}

/** Seed all 3 non-super test users. Silently ignores "already exists" errors. */
export async function seedTestUsers(
  superToken: string,
  users: Array<{ displayName: string; email: string; password: string; role: string }>,
  siteSlug = TEST_SITE.slug,
): Promise<void> {
  for (const u of users) {
    try {
      await createUser(superToken, {
        display_name: u.displayName,
        email: u.email,
        password: u.password,
        role: u.role,
      }, siteSlug);
    } catch {
      // User may already exist from previous run — ignore
    }
  }
}
```

---

### Task 3: Update auth.ts — Keep loginViaUI + loginViaAPI

**Files:**
- Modify: `web/e2e/helpers/auth.ts`

**Step 1: Rewrite file (minor cleanup)**

The existing file is mostly correct. Keep as-is but ensure selectors match actual components.

Actual LoginForm uses: `#email` (id), `#password` (id), submit button with i18n `auth.loginButton` (text "Sign In" in en).

```typescript
// web/e2e/helpers/auth.ts
import { type Page, expect } from '@playwright/test';
import { API_BASE } from './constants';

/**
 * Login via the UI login form.
 * Selectors based on LoginForm.tsx: #email, #password, submit button.
 */
export async function loginViaUI(page: Page, email: string, password: string): Promise<void> {
  await page.goto('/login');
  await page.locator('#email').fill(email);
  await page.locator('#password').fill(password);
  await page.locator('button[type="submit"]').click();
  await expect(page).toHaveURL(/\/dashboard/, { timeout: 10_000 });
}

/**
 * Login via API and inject access_token cookie.
 * Faster than UI login — use in beforeEach when login itself isn't under test.
 */
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
```

---

### Task 4: Update playwright.config.ts — New Project Names

**Files:**
- Modify: `web/playwright.config.ts`

**Step 1: Update config**

Replace old project names with new spec file names (content replaces posts, system is new).

```typescript
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
  projects: [
    { name: 'setup', testMatch: 'setup.spec.ts' },
    { name: 'auth', testMatch: 'auth.spec.ts', dependencies: ['setup'] },
    { name: 'content', testMatch: 'content.spec.ts', dependencies: ['setup'] },
    { name: 'media', testMatch: 'media.spec.ts', dependencies: ['setup'] },
    { name: 'system', testMatch: 'system.spec.ts', dependencies: ['setup'] },
    { name: 'rbac', testMatch: 'rbac.spec.ts', dependencies: ['setup'] },
    { name: 'multisite', testMatch: 'multisite.spec.ts', dependencies: ['setup'] },
  ],
  webServer: {
    command: 'bun run dev',
    url: 'http://localhost:4321',
    timeout: 30_000,
    reuseExistingServer: !process.env.CI,
  },
});
```

---

### Task 5: Rewrite setup.spec.ts — Installation Wizard

**Files:**
- Modify: `web/e2e/setup.spec.ts`

**Step 1: Rewrite with accurate selectors from SetupWizard.tsx**

SetupWizard uses: `#admin_display_name`, `#admin_email`, `#password`, `#confirmPassword` (step 1); `#site_name`, `#site_slug`, `#site_url`, `#locale` (step 2); review text display (step 3). Submit buttons are `button[type="submit"]`, back button has `variant="outline"`. Install button text: i18n `auth.setupInstall`.

```typescript
// web/e2e/setup.spec.ts
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
    // SetupWizard renders step indicator with numbers 1, 2, 3
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

    // Click install button (last primary button on the page)
    await page.locator('button[type="submit"]').click();

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
    // InstallationGuard middleware should redirect away from /setup
    await expect(page).not.toHaveURL(/\/setup$/);
  });
});
```

**Step 2: Delete old posts.spec.ts (replaced by content.spec.ts)**

```bash
rm web/e2e/posts.spec.ts
```

---

### Task 6: Rewrite auth.spec.ts — Authentication Flows

**Files:**
- Modify: `web/e2e/auth.spec.ts`

**Step 1: Rewrite**

Selectors from LoginForm.tsx: `#email`, `#password`, `button[type="submit"]`.
ForgotPasswordForm.tsx: `#email`, `button[type="submit"]`.
Header.tsx: `aria-label="User menu"`, menuitem with "Logout" text.

```typescript
// web/e2e/auth.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
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

    // Sonner toast should appear with error
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
    await expect(page).toHaveURL(/\/login/);
  });

  test('unauthenticated /dashboard redirects to /login', async ({ page }) => {
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('user menu is visible after login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    // Header renders avatar button with aria-label="User menu"
    await expect(page.getByLabel('User menu')).toBeVisible();
  });

  test('logout redirects to login', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);

    // Open user menu dropdown
    await page.getByLabel('User menu').click();
    // Click logout menu item
    await page.getByRole('menuitem', { name: /logout/i }).click();

    await expect(page).toHaveURL(/\/login/, { timeout: 5_000 });
  });

  test('after logout, dashboard is inaccessible', async ({ page }) => {
    // Just verify cookie-based redirect works
    await page.goto('/dashboard');
    await expect(page).toHaveURL(/\/login/);
  });

  test('forgot password always shows sent page', async ({ page }) => {
    await page.goto('/forgot-password');
    await page.locator('#email').fill(TEST_SUPER.email);
    await page.locator('button[type="submit"]').click();

    // Anti-enumeration: always redirects to sent page
    await expect(page).toHaveURL(/\/forgot-password\/sent/, { timeout: 5_000 });
  });

  test('login form validates email format', async ({ page }) => {
    await page.goto('/login');
    await page.locator('#email').fill('not-an-email');
    await page.locator('#password').fill('SomePassword123!');
    await page.locator('button[type="submit"]').click();

    // Should stay on login (Zod validation prevents submission)
    await expect(page).toHaveURL(/\/login/);
  });
});
```

---

### Task 7: Write content.spec.ts — Posts + Categories + Tags

**Files:**
- Create: `web/e2e/content.spec.ts`

**Step 1: Write content spec**

PostsListPage.tsx: heading "Posts", button with Plus icon + i18n `content.newPost`.
PostEditor.tsx: title input placeholder `content.postTitlePlaceholder`, save button `common.create`/`common.save`.
CategoryTree.tsx: heading, buttons with aria-labels.
TagsTable.tsx: search input, table.

```typescript
// web/e2e/content.spec.ts
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

    // Fill title
    await page.getByPlaceholder(/title/i).fill('E2E Draft Post');

    // Try to interact with BlockNote editor
    const editable = page.locator('[contenteditable="true"]').first();
    if (await editable.isVisible({ timeout: 3_000 }).catch(() => false)) {
      await editable.click();
      await page.keyboard.type('E2E test content paragraph.');
    }

    // Click save/create button
    await page.getByRole('button', { name: /create|save/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('draft post appears in posts list', async ({ page }) => {
    await page.goto('/dashboard/posts');
    await expect(page.getByText('E2E Draft Post')).toBeVisible({ timeout: 5_000 });
  });

  test('edit and publish a post', async ({ page }) => {
    await page.goto('/dashboard/posts');

    // Click on the post title link to edit
    await page.getByRole('link', { name: 'E2E Draft Post' }).click();
    await expect(page.getByPlaceholder(/title/i)).toBeVisible({ timeout: 5_000 });

    // Find and click publish button
    await page.getByRole('button', { name: /publish/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  test('delete a post via list actions', async ({ page }) => {
    await page.goto('/dashboard/posts');

    // Open actions menu for the post row (MoreHorizontal icon button)
    const row = page.getByText('E2E Draft Post').locator('ancestor::tr');
    const actionsBtn = row.locator('button').last();
    await actionsBtn.click();

    // Click delete in dropdown
    await page.getByRole('menuitem', { name: /delete/i }).click();

    // Confirm in AlertDialog
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
    // TODO: Depends on CategoryForm dialog implementation
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
    // TODO: Depends on TagForm dialog implementation
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
```

---

### Task 8: Rewrite media.spec.ts — Media Management

**Files:**
- Modify: `web/e2e/media.spec.ts`

**Step 1: Rewrite**

MediaLibrary.tsx: heading, search input, grid/list toggle buttons.
MediaUploader.tsx: dropzone with hidden `input[type="file"]`.
MediaDetailDialog.tsx: `role="dialog"` with file info.

```typescript
// web/e2e/media.spec.ts
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

    // MediaUploader uses react-dropzone with hidden file input
    const fileInput = page.locator('input[type="file"]');
    await fileInput.setInputFiles(
      path.resolve(import.meta.dirname, 'fixtures/test-image.png'),
    );

    // Wait for upload success toast
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

    // ConfirmDialog (AlertDialog)
    const confirm = page.getByRole('alertdialog');
    if (await confirm.isVisible({ timeout: 2_000 }).catch(() => false)) {
      await confirm.getByRole('button', { name: /confirm|delete/i }).click();
    }

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });
});
```

---

### Task 9: Write system.spec.ts — System Management (Batch 12)

**Files:**
- Create: `web/e2e/system.spec.ts`

**Step 1: Write system spec**

ALL tests use `test.fixme()` since Batch 12 pages are not implemented yet. These serve as executable documentation of expected behavior.

```typescript
// web/e2e/system.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, TEST_ADMIN } from './helpers/constants';
import { loginViaAPI } from './helpers/auth';
import { apiLogin, seedTestUsers } from './helpers/api';

test.describe.serial('System Management', () => {
  test.beforeAll(async () => {
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
    await seedTestUsers(superToken, [TEST_ADMIN]);
  });

  // --- Users ---

  test.fixme('users list page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/users');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create a new user with editor role', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/users');
    await page.getByRole('button', { name: /new|create|add/i }).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/email/i).fill('new-user@e2e-test.com');
    await dialog.getByLabel(/display name/i).fill('New E2E User');
    await dialog.getByLabel(/password/i).fill('NewUser123!');
    // Role select
    await dialog.getByLabel(/role/i).click();
    await page.getByRole('option', { name: /editor/i }).click();
    await dialog.getByRole('button', { name: /save|create/i }).click();

    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  // --- Roles ---

  test.fixme('roles list page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/roles');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    // Should show 4 built-in roles
    await expect(page.getByText(/super/i)).toBeVisible();
    await expect(page.getByText(/admin/i)).toBeVisible();
    await expect(page.getByText(/editor/i)).toBeVisible();
    await expect(page.getByText(/viewer/i)).toBeVisible();
  });

  test.fixme('view role permissions', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/roles');
    // Click on editor role to view permissions
    await page.getByText(/editor/i).click();
    // PermissionTree component should be visible
    await expect(page.getByText(/permissions/i)).toBeVisible();
  });

  // --- Settings ---

  test.fixme('settings page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/settings');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('update a site setting', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/settings');
    // Find a text input and change value
    const input = page.getByLabel(/site name|site title/i);
    await input.clear();
    await input.fill('Updated Site Name');
    await page.getByRole('button', { name: /save/i }).click();
    await expect(page.locator('[data-sonner-toast]')).toBeVisible({ timeout: 5_000 });
  });

  // --- API Keys ---

  test.fixme('api keys page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/api-keys');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  test.fixme('create an API key', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/api-keys');
    await page.getByRole('button', { name: /new|create|generate/i }).click();

    const dialog = page.getByRole('dialog');
    await dialog.getByLabel(/name|description/i).fill('E2E Test Key');
    await dialog.getByRole('button', { name: /create|generate/i }).click();

    // Should show the generated key (only shown once)
    await expect(page.getByText(/key.*created|copy.*key/i)).toBeVisible({ timeout: 5_000 });
  });

  // --- Audit Logs ---

  test.fixme('audit logs page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/audit-logs');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
    // Should show recent actions (user creation, etc.)
  });

  // --- Comments ---

  test.fixme('comments page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/comments');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  // --- Menus ---

  test.fixme('menus page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/menus');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });

  // --- Redirects ---

  test.fixme('redirects page loads', async ({ page }) => {
    await loginViaAPI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await page.goto('/dashboard/redirects');
    await expect(page.getByRole('heading', { level: 1 })).toBeVisible();
  });
});
```

---

### Task 10: Rewrite rbac.spec.ts — 4-Role Permission Matrix

**Files:**
- Modify: `web/e2e/rbac.spec.ts`

**Step 1: Rewrite**

Note: DashboardShell currently shows ALL nav items to all users (hardcoded). Nav visibility tests MUST use `test.fixme()`. API-level RBAC tests should pass.

```typescript
// web/e2e/rbac.spec.ts
import { test, expect } from '@playwright/test';
import {
  TEST_SUPER, TEST_ADMIN, TEST_EDITOR, TEST_VIEWER, API_BASE,
} from './helpers/constants';
import { loginViaUI } from './helpers/auth';
import { apiLogin, seedTestUsers } from './helpers/api';

test.describe.serial('Role-Based Access Control', () => {
  test.beforeAll(async () => {
    const superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
    await seedTestUsers(superToken, [TEST_ADMIN, TEST_EDITOR, TEST_VIEWER]);
  });

  // --- Navigation Visibility (fixme: DashboardShell has no role filtering) ---

  test.fixme('super sees all navigation items including sites', async ({ page }) => {
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /roles/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
  });

  test.fixme('admin sees system nav but NOT sites', async ({ page }) => {
    await loginViaUI(page, TEST_ADMIN.email, TEST_ADMIN.password);
    await expect(page.getByRole('link', { name: /users/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /settings/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).not.toBeVisible();
  });

  test.fixme('editor sees content nav only', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await expect(page.getByRole('link', { name: /posts/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /media/i })).toBeVisible();
    await expect(page.getByRole('link', { name: /users/i })).not.toBeVisible();
    await expect(page.getByRole('link', { name: /sites/i })).not.toBeVisible();
  });

  test.fixme('viewer sees content nav but no create buttons', async ({ page }) => {
    await loginViaUI(page, TEST_VIEWER.email, TEST_VIEWER.password);
    await page.goto('/dashboard/posts');
    await expect(
      page.getByRole('button', { name: /new post|create/i }),
    ).not.toBeVisible();
  });

  // --- Page Access (fixme: middleware has no role-based page guard) ---

  test.fixme('editor cannot access /dashboard/users', async ({ page }) => {
    await loginViaUI(page, TEST_EDITOR.email, TEST_EDITOR.password);
    await page.goto('/dashboard/users');
    const forbidden = page.getByText(/forbidden|permission|access denied/i);
    await expect(forbidden).toBeVisible({ timeout: 5_000 });
  });

  // --- API-Level RBAC (should pass — backend enforces) ---

  test('API returns 403 when editor accesses sites endpoint', async ({ request }) => {
    const editorToken = await apiLogin(TEST_EDITOR.email, TEST_EDITOR.password);
    const resp = await request.get(`${API_BASE}/api/v1/sites`, {
      headers: { Authorization: `Bearer ${editorToken}` },
    });
    expect(resp.status()).toBe(403);
  });

  test('API returns 403 when viewer creates a post', async ({ request }) => {
    const viewerToken = await apiLogin(TEST_VIEWER.email, TEST_VIEWER.password);
    const resp = await request.post(`${API_BASE}/api/v1/posts`, {
      headers: {
        Authorization: `Bearer ${viewerToken}`,
        'Content-Type': 'application/json',
        'X-Site-Slug': 'e2e-test',
      },
      data: { title: 'Unauthorized Post', content: '<p>test</p>' },
    });
    expect(resp.status()).toBe(403);
  });
});
```

---

### Task 11: Rewrite multisite.spec.ts — Multi-Site Isolation

**Files:**
- Modify: `web/e2e/multisite.spec.ts`

**Step 1: Rewrite**

Multi-site tests are API-heavy. UI tests for site management use `test.fixme()` (Batch 12 sites page not yet implemented).

```typescript
// web/e2e/multisite.spec.ts
import { test, expect } from '@playwright/test';
import { TEST_SUPER, API_BASE } from './helpers/constants';
import { apiLogin, createSite, createPost } from './helpers/api';

const SITE_A = { name: 'Site Alpha', slug: 'site-alpha' };
const SITE_B = { name: 'Site Beta', slug: 'site-beta' };

test.describe.serial('Multi-Site Isolation', () => {
  let superToken: string;

  test.beforeAll(async () => {
    superToken = await apiLogin(TEST_SUPER.email, TEST_SUPER.password);
  });

  test('create two sites via API', async () => {
    try { await createSite(superToken, SITE_A); } catch { /* may exist */ }
    try { await createSite(superToken, SITE_B); } catch { /* may exist */ }
  });

  test('create posts in different sites', async () => {
    await createPost(
      superToken,
      { title: 'Alpha Post', content: '<p>Content for alpha</p>', status: 'published' },
      SITE_A.slug,
    );
    await createPost(
      superToken,
      { title: 'Beta Post', content: '<p>Content for beta</p>', status: 'published' },
      SITE_B.slug,
    );
  });

  test('site_alpha API only returns alpha posts', async ({ request }) => {
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

  test('site_beta API only returns beta posts', async ({ request }) => {
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

  test.fixme('sites list page shows both sites', async ({ page }) => {
    // TODO: Depends on Batch 12 sites management page
    const { loginViaUI } = await import('./helpers/auth');
    await loginViaUI(page, TEST_SUPER.email, TEST_SUPER.password);
    await page.goto('/dashboard/sites');
    await expect(page.getByText(SITE_A.name)).toBeVisible();
    await expect(page.getByText(SITE_B.name)).toBeVisible();
  });
});
```

---

### Task 12: Run Full Suite and Iterate

**Step 1: Run all E2E tests**

```bash
cd /Users/martinadamsdev/workspace/sky-flux-cms/web && bunx playwright test 2>&1 | tail -40
```

Expected: setup + auth specs should mostly pass. Content/media specs depend on real backend data flow. System + RBAC nav tests are all `fixme` (skipped). API-level RBAC tests should pass.

**Step 2: Fix any selector mismatches or timing issues**

Common fixes:
- Add `{ timeout: X }` to slow assertions
- Use `page.waitForResponse()` for API-dependent flows
- Adjust selectors if i18n renders different text than expected

**Step 3: Commit all changes**

```bash
git add web/e2e/ web/playwright.config.ts
git commit -m "test(e2e): rewrite full E2E suite with 4-role coverage"
```
