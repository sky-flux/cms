import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { useQueryClient } from '@tanstack/react-query';
import { QueryProvider } from '@/components/providers/QueryProvider';

function QueryConsumer() {
  const client = useQueryClient();
  return <span data-testid="has-client">{client ? 'yes' : 'no'}</span>;
}

describe('QueryProvider', () => {
  it('renders children', () => {
    render(
      <QueryProvider>
        <span data-testid="child">Hello</span>
      </QueryProvider>,
    );
    expect(screen.getByTestId('child')).toHaveTextContent('Hello');
  });

  it('provides a QueryClient to children', () => {
    render(
      <QueryProvider>
        <QueryConsumer />
      </QueryProvider>,
    );
    expect(screen.getByTestId('has-client')).toHaveTextContent('yes');
  });

  it('provides a client with expected default options', () => {
    let staleTime: number | undefined;

    function OptionsChecker() {
      const client = useQueryClient();
      staleTime = client.getDefaultOptions().queries?.staleTime as number;
      return null;
    }

    render(
      <QueryProvider>
        <OptionsChecker />
      </QueryProvider>,
    );
    expect(staleTime).toBe(30_000);
  });
});
