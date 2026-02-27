import { describe, it, expect, vi } from 'vitest';

describe('categories hooks', () => {
  it('exports useCategories', async () => {
    const { useCategories } = await import('../hooks');
    expect(useCategories).toBeDefined();
  });

  it('exports useCategory', async () => {
    const { useCategory } = await import('../hooks');
    expect(useCategory).toBeDefined();
  });

  it('exports useCreateCategory', async () => {
    const { useCreateCategory } = await import('../hooks');
    expect(useCreateCategory).toBeDefined();
  });

  it('exports useUpdateCategory', async () => {
    const { useUpdateCategory } = await import('../hooks');
    expect(useUpdateCategory).toBeDefined();
  });

  it('exports useDeleteCategory', async () => {
    const { useDeleteCategory } = await import('../hooks');
    expect(useDeleteCategory).toBeDefined();
  });
});
