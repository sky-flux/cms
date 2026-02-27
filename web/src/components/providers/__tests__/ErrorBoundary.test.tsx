import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import ErrorBoundary from '../ErrorBoundary';

// A component that throws an error for testing
const ThrowError = ({ message = 'Test error' }: { message?: string }) => {
  throw new Error(message);
};

describe('ErrorBoundary', () => {
  it('should catch errors and display fallback UI', () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError}>
        <ThrowError message="Something went wrong" />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
  });

  it('should call onError callback when error occurs', async () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError}>
        <ThrowError message="Test error" />
      </ErrorBoundary>
    );

    await waitFor(() => {
      expect(onError).toHaveBeenCalledWith(
        expect.any(Error),
        expect.objectContaining({
          componentStack: expect.any(String),
        })
      );
    });
  });

  it('should render custom fallback when provided', () => {
    const customFallback = <div>Custom error message</div>;
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError} fallback={customFallback}>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByText('Custom error message')).toBeInTheDocument();
  });

  it('should not show error details when showErrorDetails is false', () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError} showErrorDetails={false}>
        <ThrowError message="Secret error" />
      </ErrorBoundary>
    );

    expect(screen.getByText('Something went wrong')).toBeInTheDocument();
    expect(screen.queryByText('Secret error')).not.toBeInTheDocument();
  });

  it('should show error details when showErrorDetails is true', () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError} showErrorDetails={true}>
        <ThrowError message="Visible error" />
      </ErrorBoundary>
    );

    expect(screen.getByText('Visible error')).toBeInTheDocument();
  });

  it('should have Reset and Reload Page buttons', () => {
    const onError = vi.fn();

    render(
      <ErrorBoundary onError={onError}>
        <ThrowError />
      </ErrorBoundary>
    );

    expect(screen.getByRole('button', { name: /reset/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /reload page/i })).toBeInTheDocument();
  });
});
