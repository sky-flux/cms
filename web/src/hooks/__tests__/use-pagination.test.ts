import { describe, it, expect } from 'vitest';
import { renderHook, act } from '@testing-library/react';
import { usePagination } from '../use-pagination';

describe('usePagination', () => {
  it('starts with default values', () => {
    const { result } = renderHook(() => usePagination());
    expect(result.current.page).toBe(1);
    expect(result.current.perPage).toBe(20);
  });

  it('accepts initial values', () => {
    const { result } = renderHook(() => usePagination({ page: 2, perPage: 50 }));
    expect(result.current.page).toBe(2);
    expect(result.current.perPage).toBe(50);
  });

  it('setPage updates page', () => {
    const { result } = renderHook(() => usePagination());
    act(() => { result.current.setPage(3); });
    expect(result.current.page).toBe(3);
  });

  it('setPerPage updates perPage and resets page to 1', () => {
    const { result } = renderHook(() => usePagination({ page: 3 }));
    act(() => { result.current.setPerPage(50); });
    expect(result.current.perPage).toBe(50);
    expect(result.current.page).toBe(1);
  });

  it('resetPage sets page to 1', () => {
    const { result } = renderHook(() => usePagination({ page: 5 }));
    act(() => { result.current.resetPage(); });
    expect(result.current.page).toBe(1);
  });
});
