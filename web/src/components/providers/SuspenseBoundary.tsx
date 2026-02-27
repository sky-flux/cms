import { Suspense, type ReactNode } from 'react';
import { Loader2 } from 'lucide-react';

export function SuspenseBoundary({
  children,
  fallback,
}: {
  children: ReactNode;
  fallback?: ReactNode;
}) {
  const defaultFallback = (
    <div role="status" className="flex items-center gap-2">
      <Loader2 className="h-4 w-4 animate-spin" />
      <span>Loading...</span>
    </div>
  );

  return <Suspense fallback={fallback || defaultFallback}>{children}</Suspense>;
}
