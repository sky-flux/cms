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
