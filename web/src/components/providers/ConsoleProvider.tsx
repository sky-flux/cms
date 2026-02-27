import { useState, type ReactNode } from 'react';
import { Toaster } from 'sonner';
import { QueryProvider } from './QueryProvider';
import { I18nProvider } from './I18nProvider';
import { ThemeProvider } from './ThemeProvider';
import { ErrorBoundary } from './ErrorBoundary';
import { SuspenseBoundary } from './SuspenseBoundary';

export interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
  showErrorDetails?: boolean;
}

/**
 * ConsoleProvider - Unified provider composition for dashboard/console pages
 *
 * Combines all necessary providers in the optimal order:
 * 1. ErrorBoundary - Catches all render errors
 * 2. QueryProvider - React Query for data fetching
 * 3. I18nProvider - Internationalization
 * 4. ThemeProvider - Dark/light theme management
 * 5. SuspenseBoundary - Unified loading states
 * 6. Toaster - Toast notifications
 *
 * Usage: Only for /dashboard pages, not for auth/setup pages.
 */
export function ConsoleProvider({
  children,
  defaultTheme = 'system',
  showErrorDetails = import.meta.env.DEV
}: ConsoleProviderProps): ReactNode {
  const [error, setError] = useState<Error | undefined>();

  const handleError = (err: Error) => {
    setError(err);
  };

  return (
    <ErrorBoundary showErrorDetails={showErrorDetails} onError={handleError}>
      <QueryProvider>
        <I18nProvider>
          <ThemeProvider defaultTheme={defaultTheme}>
            <SuspenseBoundary>
              {children}
              <Toaster />
            </SuspenseBoundary>
          </ThemeProvider>
        </I18nProvider>
      </QueryProvider>
    </ErrorBoundary>
  );
}
