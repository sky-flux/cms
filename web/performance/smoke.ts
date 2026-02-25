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
