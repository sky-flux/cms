import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { ConsoleProvider } from '../ConsoleProvider';

// Mock import.meta.env.DEV
vi.mock('import.meta.env', () => ({
  DEV: true,
}));

describe('ConsoleProvider', () => {
  it('renders children', () => {
    render(
      <ConsoleProvider>
        <div data-testid="child">Test Content</div>
      </ConsoleProvider>
    );
    expect(screen.getByTestId('child')).toHaveTextContent('Test Content');
  });
});
