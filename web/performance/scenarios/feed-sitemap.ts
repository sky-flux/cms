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
