import type { ReactNode } from 'react';
import { QueryProvider } from './QueryProvider';
import { ThemeProvider } from './ThemeProvider';
import { ErrorBoundary } from './ErrorBoundary';
import { SuspenseBoundary } from './SuspenseBoundary';

export interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
  showErrorDetails?: boolean;
}

export function ConsoleProvider({
  children,
  defaultTheme = 'system',
  showErrorDetails = import.meta.env.DEV
}: ConsoleProviderProps): ReactNode {
  return (
    <ErrorBoundary showErrorDetails={showErrorDetails}>
      <QueryProvider>
        <ThemeProvider defaultTheme={defaultTheme}>
          <SuspenseBoundary>
            {children}
          </SuspenseBoundary>
        </ThemeProvider>
      </QueryProvider>
    </ErrorBoundary>
  );
}
