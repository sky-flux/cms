import { describe, it, expect, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ThemeProvider } from '@/components/providers/ThemeProvider';

describe('ThemeProvider', () => {
  beforeEach(() => {
    document.documentElement.classList.remove('dark');
    // jsdom lacks matchMedia; provide a light-mode default
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      configurable: true,
      value: () => ({
        matches: false,
        media: '',
        addEventListener: () => {},
        removeEventListener: () => {},
      }),
    });
  });

  it('renders children', () => {
    render(
      <ThemeProvider>
        <span data-testid="child">Hello</span>
      </ThemeProvider>,
    );
    expect(screen.getByTestId('child')).toHaveTextContent('Hello');
  });

  it('applies dark class when defaultTheme is "dark"', () => {
    render(
      <ThemeProvider defaultTheme="dark">
        <span>Dark mode</span>
      </ThemeProvider>,
    );
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('removes dark class when defaultTheme is "light"', () => {
    document.documentElement.classList.add('dark');

    render(
      <ThemeProvider defaultTheme="light">
        <span>Light mode</span>
      </ThemeProvider>,
    );
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('defaults to system preference when no theme specified', () => {
    // jsdom has no matchMedia by default, so we mock it
    const darkMql = { matches: true, media: '', addEventListener: () => {}, removeEventListener: () => {} };
    Object.defineProperty(window, 'matchMedia', {
      writable: true,
      value: () => darkMql,
    });

    render(
      <ThemeProvider>
        <span>System</span>
      </ThemeProvider>,
    );
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });
});
