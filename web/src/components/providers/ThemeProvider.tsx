import { useEffect, type ReactNode } from 'react';

type Theme = 'light' | 'dark' | 'system';

export function ThemeProvider({
  children,
  defaultTheme = 'system',
}: {
  children: ReactNode;
  defaultTheme?: Theme;
}) {
  useEffect(() => {
    const root = document.documentElement;
    const resolved =
      defaultTheme === 'system'
        ? window.matchMedia('(prefers-color-scheme: dark)').matches
          ? 'dark'
          : 'light'
        : defaultTheme;
    root.classList.toggle('dark', resolved === 'dark');
  }, [defaultTheme]);

  return <>{children}</>;
}
