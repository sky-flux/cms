import { useState, useCallback } from 'react';

interface UsePaginationOptions {
  page?: number;
  perPage?: number;
}

export function usePagination(options?: UsePaginationOptions) {
  const [page, setPage] = useState(options?.page ?? 1);
  const [perPage, setPerPageValue] = useState(options?.perPage ?? 20);

  const setPerPage = useCallback((newPerPage: number) => {
    setPerPageValue(newPerPage);
    setPage(1);
  }, []);

  const resetPage = useCallback(() => setPage(1), []);

  return { page, perPage, setPage, setPerPage, resetPage };
}
