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
