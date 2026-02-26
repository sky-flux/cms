import { defineMiddleware } from 'astro:middleware';

const PUBLIC_PATHS = ['/login', '/setup', '/forgot-password', '/reset-password'];
const SETUP_CHECK_CACHE_KEY = 'cms_setup_check';
const CACHE_DURATION = 60000; // 1 minute cache

export const onRequest = defineMiddleware(async (context, next) => {
  const { pathname } = context.url;

  // Allow public paths
  if (PUBLIC_PATHS.some((p) => pathname.startsWith(p))) {
    return next();
  }

  // Allow API routes and static assets
  if (pathname.startsWith('/api') || pathname.startsWith('/_')) {
    return next();
  }

  // Check installation status (with cache)
  const now = Date.now();
  const cached = context.cookies.get(SETUP_CHECK_CACHE_KEY)?.value;
  let isInstalled = false;

  if (cached) {
    const { timestamp, installed } = JSON.parse(cached);
    // Use cached value if still valid
    if (now - timestamp < CACHE_DURATION) {
      isInstalled = installed;
    }
  }

  // If no cache or expired, check installation status
  if (!cached || now - JSON.parse(cached).timestamp >= CACHE_DURATION) {
    try {
      // Use appropriate API base URL based on environment
      const isDev = import.meta.env.DEV;
      const apiBase = isDev ? 'http://localhost:8080' : 'http://api:8080';
      const apiURL = `${apiBase}/api/v1/setup/check`;

      const response = await fetch(apiURL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
      });

      if (response.ok) {
        const result = await response.json();
        isInstalled = result.data.installed;

        // Cache the result
        context.cookies.set(
          SETUP_CHECK_CACHE_KEY,
          JSON.stringify({ timestamp: now, installed: isInstalled }),
          { path: '/', httpOnly: false, sameSite: 'lax' }
        );
      }
    } catch (error) {
      // If API is unreachable, assume installed (don't block on errors)
      console.error('Failed to check installation status:', error);
      isInstalled = true;
    }
  }

  // Redirect to setup if not installed
  if (!isInstalled && pathname !== '/setup') {
    return context.redirect('/setup');
  }

  // Check for refresh_token cookie to determine authentication status
  const refreshToken = context.cookies.get('refresh_token')?.value;
  if (!refreshToken && pathname.startsWith('/dashboard')) {
    return context.redirect('/login');
  }

  return next();
});
