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
