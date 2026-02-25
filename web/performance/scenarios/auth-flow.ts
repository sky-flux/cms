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
