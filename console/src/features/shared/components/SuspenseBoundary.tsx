import { Suspense, type ReactNode } from 'react';

interface SuspenseBoundaryProps {
  children: ReactNode;
  fallback?: ReactNode;
}

/**
 * A thin wrapper around React Suspense with a default fallback.
 * In React 19, you can also use Suspense directly.
 *
 * Usage:
 * <SuspenseBoundary fallback={<Loading />}>
 *   <SomeComponent />
 * </SuspenseBoundary>
 */
export function SuspenseBoundary({
  children,
  fallback = <div style={{ padding: '2rem', textAlign: 'center' }}>Loading...</div>
}: SuspenseBoundaryProps): ReactNode {
  return <Suspense fallback={fallback}>{children}</Suspense>;
}

// Also export React's Suspense directly for advanced use cases
export { Suspense };
