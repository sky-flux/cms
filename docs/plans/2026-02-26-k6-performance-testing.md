# k6 Performance Testing Implementation Plan

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**Goal:** Build a modular k6 performance test suite with smoke and full-load modes covering all 4 API layers.

**Architecture:** Shared lib (config/auth/checks) + 4 scenario files + 2 entry points (smoke/full-load). k6 native TypeScript support (v0.54+). Tests hit the Go backend directly at `:8080`.

**Tech Stack:** k6 (TypeScript), Go backend API at `http://localhost:8080`

---

### Task 1: Shared Config Library

**Files:**
- Create: `web/performance/lib/config.ts`

**Step 1: Create config with environment variables and stage templates**

```typescript
// web/performance/lib/config.ts

// ── Environment ──────────────────────────────────────────────────
export const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
export const API_KEY = __ENV.API_KEY || 'cms_live_test_key';
export const ADMIN_EMAIL = __ENV.ADMIN_EMAIL || 'admin@example.com';
export const ADMIN_PASSWORD = __ENV.ADMIN_PASSWORD || 'SecurePass123!';
export const SITE_SLUG = __ENV.SITE_SLUG || 'default';

// ── Thresholds ───────────────────────────────────────────────────
export const DEFAULT_THRESHOLDS = {
  http_req_duration: ['p(99)<200'],   // P99 < 200ms
  http_req_failed: ['rate<0.01'],     // Error rate < 1%
};

// ── Stage Templates ──────────────────────────────────────────────
export function smokeStages() {
  return [
    { duration: '10s', target: 5 },
    { duration: '20s', target: 5 },
    { duration: '5s', target: 0 },
  ];
}

export function loadStages() {
  return [
    { duration: '30s', target: 50 },     // Warm-up
    { duration: '1m', target: 50 },       // Baseline
    { duration: '1m', target: 500 },      // Scale up
    { duration: '2m', target: 500 },      // Medium load
    { duration: '1m', target: 1000 },     // Peak ramp
    { duration: '2m', target: 1000 },     // Peak load
    { duration: '30s', target: 0 },       // Cool-down
  ];
}

// ── Common Headers ───────────────────────────────────────────────
export function apiKeyHeaders() {
  return {
    'X-API-Key': API_KEY,
    'X-Site-Slug': SITE_SLUG,
  };
}

export function siteHeaders() {
  return {
    'X-Site-Slug': SITE_SLUG,
    'Content-Type': 'application/json',
  };
}
```

**Step 2: Verify file exists**

Run: `ls web/performance/lib/config.ts`
Expected: file listed

---

### Task 2: Auth Helper Library

**Files:**
- Create: `web/performance/lib/auth.ts`

**Step 1: Create auth helper with login and header utilities**

```typescript
// web/performance/lib/auth.ts
import http from 'k6/http';
import { check } from 'k6';
import { BASE_URL, ADMIN_EMAIL, ADMIN_PASSWORD, SITE_SLUG } from './config.ts';

export interface AuthData {
  accessToken: string;
  refreshCookie: string;
}

/**
 * Login and return tokens. Call from setup() phase.
 */
export function login(
  email: string = ADMIN_EMAIL,
  password: string = ADMIN_PASSWORD,
): AuthData {
  const res = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email, password }),
    { headers: { 'Content-Type': 'application/json' } },
  );

  check(res, {
    'login status 200': (r) => r.status === 200,
  });

  const body = res.json() as { data: { access_token: string } };
  const cookies = res.cookies;
  const refreshCookie = cookies['refresh_token']
    ? cookies['refresh_token'][0].value
    : '';

  return {
    accessToken: body.data.access_token,
    refreshCookie,
  };
}

/**
 * Build Authorization headers from token.
 */
export function authHeaders(token: string) {
  return {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
    'X-Site-Slug': SITE_SLUG,
  };
}
```

---

### Task 3: Shared Check Utilities

**Files:**
- Create: `web/performance/lib/checks.ts`

**Step 1: Create reusable check functions**

```typescript
// web/performance/lib/checks.ts
import { check } from 'k6';
import { type RefinedResponse, type ResponseType } from 'k6/http';

/**
 * Check HTTP response status and optional latency threshold.
 */
export function checkResponse(
  res: RefinedResponse<ResponseType>,
  name: string,
  maxLatencyMs?: number,
): boolean {
  const checks: Record<string, (r: RefinedResponse<ResponseType>) => boolean> = {
    [`${name} status 2xx`]: (r) => r.status >= 200 && r.status < 300,
  };

  if (maxLatencyMs) {
    checks[`${name} latency < ${maxLatencyMs}ms`] = (r) =>
      r.timings.duration < maxLatencyMs;
  }

  return check(res, checks);
}

/**
 * Check JSON response has success: true.
 */
export function checkJSON(
  res: RefinedResponse<ResponseType>,
  name: string,
): boolean {
  return check(res, {
    [`${name} is JSON`]: (r) => {
      try {
        r.json();
        return true;
      } catch {
        return false;
      }
    },
    [`${name} success: true`]: (r) => {
      try {
        const body = r.json() as { success?: boolean };
        return body.success === true;
      } catch {
        return false;
      }
    },
  });
}

/**
 * Check XML response (for feeds/sitemaps).
 */
export function checkXML(
  res: RefinedResponse<ResponseType>,
  name: string,
): boolean {
  return check(res, {
    [`${name} status 200`]: (r) => r.status === 200,
    [`${name} is XML`]: (r) => {
      const ct = r.headers['Content-Type'] || '';
      return ct.includes('xml') || ct.includes('rss');
    },
  });
}
```

**Step 2: Commit shared libraries**

```bash
git add web/performance/lib/
git commit -m "feat(perf): add k6 shared libraries (config, auth, checks)"
```

---

### Task 4: Public API Scenario

**Files:**
- Create: `web/performance/scenarios/public-api.ts`

**Step 1: Implement public API load test scenario**

```typescript
// web/performance/scenarios/public-api.ts
import http from 'k6/http';
import { sleep } from 'k6';
import { BASE_URL, apiKeyHeaders, DEFAULT_THRESHOLDS, loadStages } from '../lib/config.ts';
import { checkResponse, checkJSON } from '../lib/checks.ts';

export const options = {
  stages: loadStages(),
  thresholds: DEFAULT_THRESHOLDS,
};

const HEADERS = apiKeyHeaders();

export default function () {
  // 1. Post list (paginated)
  const listRes = http.get(`${BASE_URL}/api/public/v1/posts?page=1&per_page=10`, {
    headers: HEADERS,
  });
  checkResponse(listRes, 'public:posts:list', 200);

  sleep(0.3);

  // 2. Post detail (by slug — use a known test slug or first from list)
  const detailRes = http.get(`${BASE_URL}/api/public/v1/posts/test-post`, {
    headers: HEADERS,
  });
  checkResponse(detailRes, 'public:posts:detail', 200);

  sleep(0.3);

  // 3. Categories tree
  const catRes = http.get(`${BASE_URL}/api/public/v1/categories`, {
    headers: HEADERS,
  });
  checkResponse(catRes, 'public:categories', 200);

  sleep(0.2);

  // 4. Tags list
  const tagRes = http.get(`${BASE_URL}/api/public/v1/tags`, {
    headers: HEADERS,
  });
  checkResponse(tagRes, 'public:tags', 200);

  sleep(0.2);

  // 5. Full-text search
  const searchRes = http.get(`${BASE_URL}/api/public/v1/search?q=test`, {
    headers: HEADERS,
  });
  checkResponse(searchRes, 'public:search', 200);

  sleep(0.3);

  // 6. Comments for a post
  const commentsRes = http.get(
    `${BASE_URL}/api/public/v1/posts/test-post/comments`,
    { headers: HEADERS },
  );
  checkResponse(commentsRes, 'public:comments', 200);

  sleep(0.2);

  // 7. Menu by location
  const menuRes = http.get(
    `${BASE_URL}/api/public/v1/menus?location=header`,
    { headers: HEADERS },
  );
  checkResponse(menuRes, 'public:menus', 200);

  sleep(0.5);
}
```

---

### Task 5: Auth Flow Scenario

**Files:**
- Create: `web/performance/scenarios/auth-flow.ts`

**Step 1: Implement auth login/refresh/logout cycle**

```typescript
// web/performance/scenarios/auth-flow.ts
import http from 'k6/http';
import { check, sleep } from 'k6';
import {
  BASE_URL,
  ADMIN_EMAIL,
  ADMIN_PASSWORD,
  DEFAULT_THRESHOLDS,
  loadStages,
} from '../lib/config.ts';
import { checkResponse, checkJSON } from '../lib/checks.ts';

export const options = {
  stages: loadStages(),
  thresholds: DEFAULT_THRESHOLDS,
};

export default function () {
  // 1. Login
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  checkResponse(loginRes, 'auth:login', 200);

  if (loginRes.status !== 200) {
    sleep(1);
    return; // Skip rest if login fails
  }

  const loginBody = loginRes.json() as { data: { access_token: string } };
  const token = loginBody.data.access_token;
  const bearerHeaders = {
    Authorization: `Bearer ${token}`,
    'Content-Type': 'application/json',
  };

  sleep(0.3);

  // 2. Get current user
  const meRes = http.get(`${BASE_URL}/api/v1/auth/me`, {
    headers: bearerHeaders,
  });
  checkResponse(meRes, 'auth:me', 200);

  sleep(0.3);

  // 3. Refresh token
  const refreshRes = http.post(`${BASE_URL}/api/v1/auth/refresh`, null, {
    headers: { 'Content-Type': 'application/json' },
  });
  // Refresh may fail without proper cookie in k6, check gracefully
  check(refreshRes, {
    'auth:refresh status ok': (r) => r.status === 200 || r.status === 401,
  });

  sleep(0.3);

  // 4. Logout
  const logoutRes = http.post(`${BASE_URL}/api/v1/auth/logout`, null, {
    headers: bearerHeaders,
  });
  checkResponse(logoutRes, 'auth:logout', 200);

  sleep(0.5);
}
```

**Step 2: Commit scenario files**

```bash
git add web/performance/scenarios/
git commit -m "feat(perf): add public-api and auth-flow k6 scenarios"
```

---

### Task 6: Content Management Scenario

**Files:**
- Create: `web/performance/scenarios/content-mgmt.ts`

**Step 1: Implement content management operations**

```typescript
// web/performance/scenarios/content-mgmt.ts
import http from 'k6/http';
import { sleep } from 'k6';
import { BASE_URL, DEFAULT_THRESHOLDS, loadStages } from '../lib/config.ts';
import { login, authHeaders } from '../lib/auth.ts';
import { checkResponse } from '../lib/checks.ts';

export const options = {
  stages: loadStages(),
  thresholds: DEFAULT_THRESHOLDS,
};

export interface SetupData {
  token: string;
}

// Login once in setup phase, share token across VUs
export function setup(): SetupData {
  const { accessToken } = login();
  return { token: accessToken };
}

export default function (data: SetupData) {
  const headers = authHeaders(data.token);

  // ── Posts (read-heavy) ────────────────────────────────
  // 1. List posts
  const listRes = http.get(`${BASE_URL}/api/v1/site/posts?page=1&per_page=20`, {
    headers,
  });
  checkResponse(listRes, 'mgmt:posts:list', 200);
  sleep(0.3);

  // 2. Create a draft post
  const createRes = http.post(
    `${BASE_URL}/api/v1/site/posts`,
    JSON.stringify({
      title: `k6 Draft ${Date.now()}`,
      content: '<p>Performance test content</p>',
      status: 'draft',
    }),
    { headers },
  );
  checkResponse(createRes, 'mgmt:posts:create', 200);

  let postId = '';
  if (createRes.status === 201 || createRes.status === 200) {
    const body = createRes.json() as { data: { id: string } };
    postId = body.data?.id || '';
  }

  sleep(0.3);

  // 3. Update the post (if created)
  if (postId) {
    const updateRes = http.put(
      `${BASE_URL}/api/v1/site/posts/${postId}`,
      JSON.stringify({
        title: `k6 Updated ${Date.now()}`,
        content: '<p>Updated content</p>',
      }),
      { headers },
    );
    checkResponse(updateRes, 'mgmt:posts:update', 200);
    sleep(0.2);

    // 4. Publish
    const publishRes = http.post(
      `${BASE_URL}/api/v1/site/posts/${postId}/publish`,
      null,
      { headers },
    );
    checkResponse(publishRes, 'mgmt:posts:publish', 200);
    sleep(0.2);

    // 5. Delete (soft)
    const deleteRes = http.del(
      `${BASE_URL}/api/v1/site/posts/${postId}`,
      null,
      { headers },
    );
    checkResponse(deleteRes, 'mgmt:posts:delete', 200);
  }

  sleep(0.3);

  // ── Categories ────────────────────────────────────────
  const catRes = http.get(`${BASE_URL}/api/v1/site/categories`, { headers });
  checkResponse(catRes, 'mgmt:categories:list', 200);
  sleep(0.2);

  // ── Tags ──────────────────────────────────────────────
  const tagRes = http.get(`${BASE_URL}/api/v1/site/tags`, { headers });
  checkResponse(tagRes, 'mgmt:tags:list', 200);
  sleep(0.2);

  const suggestRes = http.get(
    `${BASE_URL}/api/v1/site/tags/suggest?q=test`,
    { headers },
  );
  checkResponse(suggestRes, 'mgmt:tags:suggest', 200);
  sleep(0.2);

  // ── Media ─────────────────────────────────────────────
  const mediaRes = http.get(`${BASE_URL}/api/v1/site/media`, { headers });
  checkResponse(mediaRes, 'mgmt:media:list', 200);

  sleep(0.5);
}
```

---

### Task 7: Feed & Sitemap Scenario

**Files:**
- Create: `web/performance/scenarios/feed-sitemap.ts`

**Step 1: Implement feed and sitemap crawl simulation**

```typescript
// web/performance/scenarios/feed-sitemap.ts
import http from 'k6/http';
import { sleep } from 'k6';
import { BASE_URL, DEFAULT_THRESHOLDS, loadStages, SITE_SLUG } from '../lib/config.ts';
import { checkXML } from '../lib/checks.ts';

export const options = {
  stages: loadStages(),
  thresholds: DEFAULT_THRESHOLDS,
};

const HEADERS = { 'X-Site-Slug': SITE_SLUG };

export default function () {
  // 1. Sitemap index
  const sitemapRes = http.get(`${BASE_URL}/sitemap.xml`, { headers: HEADERS });
  checkXML(sitemapRes, 'feed:sitemap-index');
  sleep(0.5);

  // 2. Posts sitemap
  const postsMapRes = http.get(`${BASE_URL}/sitemap-posts.xml`, {
    headers: HEADERS,
  });
  checkXML(postsMapRes, 'feed:sitemap-posts');
  sleep(0.5);

  // 3. RSS feed
  const rssRes = http.get(`${BASE_URL}/feed/rss.xml`, { headers: HEADERS });
  checkXML(rssRes, 'feed:rss');
  sleep(0.5);

  // 4. Atom feed
  const atomRes = http.get(`${BASE_URL}/feed/atom.xml`, { headers: HEADERS });
  checkXML(atomRes, 'feed:atom');

  sleep(1);
}
```

**Step 2: Commit remaining scenarios**

```bash
git add web/performance/scenarios/
git commit -m "feat(perf): add content-mgmt and feed-sitemap k6 scenarios"
```

---

### Task 8: Smoke Test Entry Point

**Files:**
- Create: `web/performance/smoke.ts`

**Step 1: Create smoke test combining all scenarios at low concurrency**

```typescript
// web/performance/smoke.ts
//
// Smoke test: 5 VUs x 30s — quick validation that all endpoints respond.
// Usage: k6 run web/performance/smoke.ts
//
import http from 'k6/http';
import { sleep } from 'k6';
import { BASE_URL, API_KEY, ADMIN_EMAIL, ADMIN_PASSWORD, SITE_SLUG, DEFAULT_THRESHOLDS, smokeStages } from './lib/config.ts';
import { checkResponse, checkXML } from './lib/checks.ts';

export const options = {
  stages: smokeStages(),
  thresholds: DEFAULT_THRESHOLDS,
};

export default function () {
  const apiHeaders = { 'X-API-Key': API_KEY, 'X-Site-Slug': SITE_SLUG };
  const siteHeaders = { 'X-Site-Slug': SITE_SLUG };

  // ── Health ────────────────────────────────────────────
  checkResponse(http.get(`${BASE_URL}/health`), 'smoke:health');

  // ── Public API (3 key endpoints) ─────────────────────
  checkResponse(
    http.get(`${BASE_URL}/api/public/v1/posts`, { headers: apiHeaders }),
    'smoke:public:posts',
  );
  checkResponse(
    http.get(`${BASE_URL}/api/public/v1/categories`, { headers: apiHeaders }),
    'smoke:public:categories',
  );
  checkResponse(
    http.get(`${BASE_URL}/api/public/v1/tags`, { headers: apiHeaders }),
    'smoke:public:tags',
  );

  // ── Auth ──────────────────────────────────────────────
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  checkResponse(loginRes, 'smoke:auth:login');

  if (loginRes.status === 200) {
    const body = loginRes.json() as { data: { access_token: string } };
    const bearerHeaders = {
      Authorization: `Bearer ${body.data.access_token}`,
      'Content-Type': 'application/json',
      'X-Site-Slug': SITE_SLUG,
    };

    checkResponse(
      http.get(`${BASE_URL}/api/v1/auth/me`, { headers: bearerHeaders }),
      'smoke:auth:me',
    );

    // Site-scoped endpoints
    checkResponse(
      http.get(`${BASE_URL}/api/v1/site/posts`, { headers: bearerHeaders }),
      'smoke:mgmt:posts',
    );
    checkResponse(
      http.get(`${BASE_URL}/api/v1/site/categories`, { headers: bearerHeaders }),
      'smoke:mgmt:categories',
    );

    // Logout
    http.post(`${BASE_URL}/api/v1/auth/logout`, null, { headers: bearerHeaders });
  }

  // ── Feed/Sitemap ──────────────────────────────────────
  checkXML(
    http.get(`${BASE_URL}/feed/rss.xml`, { headers: siteHeaders }),
    'smoke:feed:rss',
  );
  checkXML(
    http.get(`${BASE_URL}/sitemap.xml`, { headers: siteHeaders }),
    'smoke:feed:sitemap',
  );

  sleep(1);
}
```

---

### Task 9: Full Load Test Entry Point

**Files:**
- Create: `web/performance/full-load.ts`

**Step 1: Create full load test using k6 scenarios API**

This is the key design decision — `full-load.ts` uses k6's `scenarios` configuration to run all 4 scenario functions in parallel with realistic traffic distribution.

```typescript
// web/performance/full-load.ts
//
// Full staged load test: 4 scenarios in parallel, ramp to 1000 VUs.
// Usage: k6 run web/performance/full-load.ts
//
import { sleep } from 'k6';
import http from 'k6/http';
import {
  BASE_URL,
  API_KEY,
  ADMIN_EMAIL,
  ADMIN_PASSWORD,
  SITE_SLUG,
  DEFAULT_THRESHOLDS,
} from './lib/config.ts';
import { login, authHeaders } from './lib/auth.ts';
import { checkResponse, checkJSON, checkXML } from './lib/checks.ts';

// ── Setup: login once, share token ──────────────────────────────
export function setup() {
  const { accessToken } = login();
  return { token: accessToken };
}

// ── Scenario configs with traffic distribution ──────────────────
export const options = {
  scenarios: {
    public_api: {
      executor: 'ramping-vus',
      exec: 'publicAPI',
      stages: [
        { duration: '30s', target: 35 },
        { duration: '1m', target: 35 },
        { duration: '1m', target: 350 },
        { duration: '2m', target: 350 },
        { duration: '1m', target: 700 },
        { duration: '2m', target: 700 },
        { duration: '30s', target: 0 },
      ],
    },
    content_mgmt: {
      executor: 'ramping-vus',
      exec: 'contentMgmt',
      stages: [
        { duration: '30s', target: 8 },
        { duration: '1m', target: 8 },
        { duration: '1m', target: 75 },
        { duration: '2m', target: 75 },
        { duration: '1m', target: 150 },
        { duration: '2m', target: 150 },
        { duration: '30s', target: 0 },
      ],
    },
    auth_flow: {
      executor: 'ramping-vus',
      exec: 'authFlow',
      stages: [
        { duration: '30s', target: 5 },
        { duration: '1m', target: 5 },
        { duration: '1m', target: 50 },
        { duration: '2m', target: 50 },
        { duration: '1m', target: 100 },
        { duration: '2m', target: 100 },
        { duration: '30s', target: 0 },
      ],
    },
    feed_sitemap: {
      executor: 'ramping-vus',
      exec: 'feedSitemap',
      stages: [
        { duration: '30s', target: 2 },
        { duration: '1m', target: 2 },
        { duration: '1m', target: 25 },
        { duration: '2m', target: 25 },
        { duration: '1m', target: 50 },
        { duration: '2m', target: 50 },
        { duration: '30s', target: 0 },
      ],
    },
  },
  thresholds: {
    ...DEFAULT_THRESHOLDS,
    'http_req_duration{scenario:public_api}': ['p(99)<200'],
    'http_req_duration{scenario:auth_flow}': ['p(99)<300'],
    'http_req_duration{scenario:content_mgmt}': ['p(99)<500'],
    'http_req_duration{scenario:feed_sitemap}': ['p(99)<200'],
  },
};

// ── Scenario: Public API (70%) ──────────────────────────────────
const apiHeaders = { 'X-API-Key': API_KEY, 'X-Site-Slug': SITE_SLUG };

export function publicAPI() {
  checkResponse(http.get(`${BASE_URL}/api/public/v1/posts?page=1&per_page=10`, { headers: apiHeaders }), 'public:posts:list');
  sleep(0.3);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/posts/test-post`, { headers: apiHeaders }), 'public:posts:detail');
  sleep(0.3);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/categories`, { headers: apiHeaders }), 'public:categories');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/tags`, { headers: apiHeaders }), 'public:tags');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/search?q=test`, { headers: apiHeaders }), 'public:search');
  sleep(0.3);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/posts/test-post/comments`, { headers: apiHeaders }), 'public:comments');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/public/v1/menus?location=header`, { headers: apiHeaders }), 'public:menus');
  sleep(0.5);
}

// ── Scenario: Auth Flow (10%) ───────────────────────────────────
export function authFlow() {
  const loginRes = http.post(
    `${BASE_URL}/api/v1/auth/login`,
    JSON.stringify({ email: ADMIN_EMAIL, password: ADMIN_PASSWORD }),
    { headers: { 'Content-Type': 'application/json' } },
  );
  checkResponse(loginRes, 'auth:login');

  if (loginRes.status !== 200) { sleep(1); return; }

  const body = loginRes.json() as { data: { access_token: string } };
  const bearerHeaders = {
    Authorization: `Bearer ${body.data.access_token}`,
    'Content-Type': 'application/json',
  };

  sleep(0.3);
  checkResponse(http.get(`${BASE_URL}/api/v1/auth/me`, { headers: bearerHeaders }), 'auth:me');
  sleep(0.3);
  http.post(`${BASE_URL}/api/v1/auth/logout`, null, { headers: bearerHeaders });
  sleep(0.5);
}

// ── Scenario: Content Management (15%) ──────────────────────────
export function contentMgmt(data: { token: string }) {
  const headers = authHeaders(data.token);

  // Read operations (majority)
  checkResponse(http.get(`${BASE_URL}/api/v1/site/posts?page=1&per_page=20`, { headers }), 'mgmt:posts:list');
  sleep(0.3);
  checkResponse(http.get(`${BASE_URL}/api/v1/site/categories`, { headers }), 'mgmt:categories');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/v1/site/tags`, { headers }), 'mgmt:tags');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/v1/site/media`, { headers }), 'mgmt:media');
  sleep(0.2);
  checkResponse(http.get(`${BASE_URL}/api/v1/site/tags/suggest?q=test`, { headers }), 'mgmt:tags:suggest');

  // Write operation (less frequent)
  if (Math.random() < 0.2) {
    const createRes = http.post(
      `${BASE_URL}/api/v1/site/posts`,
      JSON.stringify({
        title: `k6 Load ${Date.now()}`,
        content: '<p>Load test</p>',
        status: 'draft',
      }),
      { headers },
    );
    checkResponse(createRes, 'mgmt:posts:create');

    if (createRes.status === 201 || createRes.status === 200) {
      const createBody = createRes.json() as { data: { id: string } };
      const postId = createBody.data?.id;
      if (postId) {
        http.del(`${BASE_URL}/api/v1/site/posts/${postId}`, null, { headers });
      }
    }
  }

  sleep(0.5);
}

// ── Scenario: Feed & Sitemap (5%) ───────────────────────────────
const siteHeaders = { 'X-Site-Slug': SITE_SLUG };

export function feedSitemap() {
  checkXML(http.get(`${BASE_URL}/sitemap.xml`, { headers: siteHeaders }), 'feed:sitemap');
  sleep(0.5);
  checkXML(http.get(`${BASE_URL}/sitemap-posts.xml`, { headers: siteHeaders }), 'feed:sitemap-posts');
  sleep(0.5);
  checkXML(http.get(`${BASE_URL}/feed/rss.xml`, { headers: siteHeaders }), 'feed:rss');
  sleep(0.5);
  checkXML(http.get(`${BASE_URL}/feed/atom.xml`, { headers: siteHeaders }), 'feed:atom');
  sleep(1);
}
```

**Step 2: Commit entry points**

```bash
git add web/performance/smoke.ts web/performance/full-load.ts
git commit -m "feat(perf): add smoke and full-load k6 entry points"
```

---

### Task 10: Makefile Integration

**Files:**
- Modify: `Makefile`

**Step 1: Add performance test targets to Makefile**

Add after the existing `test-coverage` target:

```makefile
test-perf-smoke:
	k6 run web/performance/smoke.ts

test-perf:
	k6 run web/performance/full-load.ts

test-perf-public:
	k6 run web/performance/scenarios/public-api.ts

test-all: test test-perf-smoke
```

Update the `.PHONY` line to include the new targets.

**Step 2: Commit Makefile changes**

```bash
git add Makefile
git commit -m "feat: add k6 performance test targets to Makefile"
```

---

### Task 11: TODO(human) — Full-Load Scenario Traffic Distribution

This task is reserved for the human to implement the `contentMgmt` write-path logic — the decision of how aggressively to create/delete test data during load testing.

---

### Summary

| Task | File(s) | Lines | Purpose |
|------|---------|-------|---------|
| 1 | lib/config.ts | ~50 | Env vars, thresholds, stage templates |
| 2 | lib/auth.ts | ~45 | Login helper, auth headers |
| 3 | lib/checks.ts | ~60 | Shared check/validation functions |
| 4 | scenarios/public-api.ts | ~65 | 7 Public API endpoints |
| 5 | scenarios/auth-flow.ts | ~55 | Login/me/refresh/logout cycle |
| 6 | scenarios/content-mgmt.ts | ~90 | Posts CRUD + categories + tags + media |
| 7 | scenarios/feed-sitemap.ts | ~35 | RSS/Atom/Sitemap crawl |
| 8 | smoke.ts | ~70 | CI smoke test, 5 VUs x 30s |
| 9 | full-load.ts | ~140 | Full parallel load, 4 scenarios, 1000 VUs |
| 10 | Makefile | ~5 | test-perf-smoke, test-perf targets |
