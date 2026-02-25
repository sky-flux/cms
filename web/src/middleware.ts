import { defineMiddleware } from 'astro:middleware';

const PUBLIC_PATHS = ['/login', '/setup', '/forgot-password', '/reset-password'];

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

  // Check for auth cookie on dashboard routes
  const token = context.cookies.get('access_token')?.value;
  if (!token && pathname.startsWith('/dashboard')) {
    return context.redirect('/login');
  }

  return next();
});
