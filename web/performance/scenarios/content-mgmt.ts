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

  // 2. Write operations: deterministic every 10th iteration
  // Create → Update → Publish (no delete — tests data accumulation)
  if (__ITER % 10 === 0) {
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

      // Publish triggers Meilisearch index sync
      const publishRes = http.post(
        `${BASE_URL}/api/v1/site/posts/${postId}/publish`,
        null,
        { headers },
      );
      checkResponse(publishRes, 'mgmt:posts:publish', 200);
    }
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
