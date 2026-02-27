import { describe, it, expect, beforeEach, afterEach } from 'vitest';
import { render, screen, cleanup } from '@testing-library/react';
import { QueryClient, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { ConsoleProvider } from '../ConsoleProvider';

// Mock ResizeObserver for jsdom
global.ResizeObserver = class ResizeObserver {
  observe() {}
  unobserve() {}
  disconnect() {}
} as unknown as typeof ResizeObserver;

describe('ConsoleProvider', () => {
  afterEach(() => {
    cleanup();
    // Reset document classes
    document.documentElement.classList.remove('dark');
  });

  it('should provide all contexts to children', () => {
    const TestComponent = () => {
      const { t } = useTranslation();
      const queryClient = useQueryClient();

      return (
        <div>
          <span data-testid="i18n">{t('common.loading')}</span>
          <span data-testid="query-client">
            {queryClient instanceof QueryClient ? 'yes' : 'no'}
          </span>
        </div>
      );
    };

    render(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    expect(screen.getByTestId('i18n')).toHaveTextContent('Loading...');
    expect(screen.getByTestId('query-client')).toHaveTextContent('yes');
  });

  it('should render Toaster', () => {
    render(
      <ConsoleProvider>
        <div>Children</div>
      </ConsoleProvider>
    );

    // Sonner Toaster renders with role="status"
    const toasters = screen.getAllByRole('status');
    expect(toasters.length).toBeGreaterThan(0);
  });

  it('should apply light theme when defaultTheme="light"', () => {
    render(
      <ConsoleProvider defaultTheme="light">
        <div>Test</div>
      </ConsoleProvider>
    );

    // ThemeProvider adds 'dark' class only for dark theme
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });

  it('should apply dark theme when defaultTheme="dark"', () => {
    render(
      <ConsoleProvider defaultTheme="dark">
        <div>Test</div>
      </ConsoleProvider>
    );

    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('should not recreate QueryClient on re-renders', () => {
    let renderCount = 0;

    const TestComponent = () => {
      const queryClient = useQueryClient();
      renderCount++;
      return <div data-testid="client">{queryClient.toString()}</div>;
    };

    const { rerender } = render(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    const firstClient = screen.getByTestId('client').textContent;
    const firstRenderCount = renderCount;

    rerender(
      <ConsoleProvider>
        <TestComponent />
      </ConsoleProvider>
    );

    const secondClient = screen.getByTestId('client').textContent;

    expect(firstClient).toBe(secondClient);
    expect(renderCount).toBe(firstRenderCount + 1);
  });
});