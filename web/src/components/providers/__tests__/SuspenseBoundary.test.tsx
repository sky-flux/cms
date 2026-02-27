import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SuspenseBoundary } from '@/components/providers/SuspenseBoundary';

// Component that throws a promise to trigger Suspense
function SuspenseComponent() {
  throw new Promise(() => {});
}

describe('SuspenseBoundary', () => {
  it('should show default fallback while children are suspending', () => {
    render(
      <SuspenseBoundary>
        <SuspenseComponent />
      </SuspenseBoundary>,
    );

    const fallback = screen.getByRole('status');
    expect(fallback).toBeInTheDocument();
    expect(fallback).toHaveTextContent('Loading...');
  });

  it('should render custom fallback when provided', () => {
    const customFallback = <div data-testid="custom-fallback">Custom loading...</div>;

    render(
      <SuspenseBoundary fallback={customFallback}>
        <SuspenseComponent />
      </SuspenseBoundary>,
    );

    expect(screen.getByTestId('custom-fallback')).toBeInTheDocument();
    expect(screen.getByTestId('custom-fallback')).toHaveTextContent('Custom loading...');
  });

  it('should render children when no suspension occurs', () => {
    render(
      <SuspenseBoundary>
        <span data-testid="child">Hello World</span>
      </SuspenseBoundary>,
    );

    expect(screen.getByTestId('child')).toHaveTextContent('Hello World');
    expect(screen.queryByRole('status')).not.toBeInTheDocument();
  });
});
