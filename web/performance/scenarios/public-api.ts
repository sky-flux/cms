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
