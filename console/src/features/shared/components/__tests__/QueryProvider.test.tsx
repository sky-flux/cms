import { describe, it, expect } from 'vitest';
import { render, screen } from '@testing-library/react';
import { QueryProvider } from '../QueryProvider';
import { useQuery } from '@tanstack/react-query';

const TestQuery = () => {
  const { data } = useQuery({
    queryKey: ['test'],
    queryFn: () => Promise.resolve('test-data'),
  });
  return <div data-testid="data">{data}</div>;
};

describe('QueryProvider', () => {
  it('renders with query client', () => {
    render(
      <QueryProvider>
        <TestQuery />
      </QueryProvider>
    );
    expect(screen.getByTestId('data')).toBeInTheDocument();
  });
});
