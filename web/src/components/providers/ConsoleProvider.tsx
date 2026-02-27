import type { ReactNode } from 'react';
import { Toaster } from 'sonner';
import { QueryProvider } from './QueryProvider';
import { I18nProvider } from './I18nProvider';
import { ThemeProvider } from './ThemeProvider';

/**
 * ConsoleProvider - Unified provider composition for dashboard/console pages
 *
 * This component combines all necessary providers for /dashboard routes:
 * - QueryProvider: React Query for data fetching
 * - I18nProvider: Internationalization
 * - ThemeProvider: Dark/light theme management
 * - Toaster: Toast notifications
 *
 * Usage: Only for dashboard/console pages, not for auth/setup pages.
 *
 * TODO(human): Decide on the optimal Provider nesting order and implement it.
 * Consider:
 * 1. Which providers depend on others? (e.g., does Toaster need i18n context?)
 * 2. Should ThemeProvider be inside or outside I18nProvider?
 * 3. What should the defaultTheme be? (current placeholder is 'system')
 *
 * Example structure to implement:
 * <QueryProvider>
 *   <I18nProvider>
 *     <ThemeProvider defaultTheme={defaultTheme}>
 *       {children}
 *       <Toaster />
 *     </ThemeProvider>
 *   </I18nProvider>
 * </QueryProvider>
 */
export function ConsoleProvider({ children, defaultTheme }: ConsoleProviderProps) {
  // TODO(human): Implement the provider composition here
  return null;
}

export interface ConsoleProviderProps {
  children: ReactNode;
  defaultTheme?: 'light' | 'dark' | 'system';
}
