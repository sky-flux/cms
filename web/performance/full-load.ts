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

  // Deterministic write: every 10th iteration creates + publishes a post.
  // No delete — tests data accumulation pressure and Meilisearch index growth.
  if (__ITER % 10 === 0) {
    const createRes = http.post(
      `${BASE_URL}/api/v1/site/posts`,
      JSON.stringify({
        title: `k6 Load ${Date.now()}`,
        content: '<p>Load test content for performance benchmarking</p>',
        status: 'draft',
      }),
      { headers },
    );
    checkResponse(createRes, 'mgmt:posts:create');

    if (createRes.status === 201 || createRes.status === 200) {
      const createBody = createRes.json() as { data: { id: string } };
      const postId = createBody.data?.id;
      if (postId) {
        // Publish triggers Meilisearch index sync
        const publishRes = http.post(
          `${BASE_URL}/api/v1/site/posts/${postId}/publish`,
          null,
          { headers },
        );
        checkResponse(publishRes, 'mgmt:posts:publish');
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
