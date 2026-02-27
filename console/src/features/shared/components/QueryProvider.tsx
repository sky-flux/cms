import type { ReactNode } from 'react';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';

export interface QueryProviderProps {
  children: ReactNode;
}

const defaultQueryClient = new QueryClient({
  defaultOptions: {
    queries: {
      staleTime: 1000 * 60 * 5, // 5 minutes
      retry: 3,
    },
  },
});

export function QueryProvider({ children }: QueryProviderProps): ReactNode {
  return (
    <QueryClientProvider client={defaultQueryClient}>
      {children}
    </QueryClientProvider>
  );
}
