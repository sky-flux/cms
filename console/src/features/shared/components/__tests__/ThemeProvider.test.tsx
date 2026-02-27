import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ThemeProvider, useTheme } from '../ThemeProvider';

const TestComponent = () => {
  const { theme } = useTheme();
  return <div data-testid="theme">{theme}</div>;
};

describe('ThemeProvider', () => {
  it('provides theme context', () => {
    render(
      <ThemeProvider>
        <TestComponent />
      </ThemeProvider>
    );
    expect(screen.getByTestId('theme')).toBeInTheDocument();
  });
});
