import { describe, it, expect } from 'vitest';

describe('tags hooks', () => {
  it('exports useTags', async () => {
    const { useTags } = await import('../hooks');
    expect(useTags).toBeDefined();
  });

  it('exports useCreateTag', async () => {
    const { useCreateTag } = await import('../hooks');
    expect(useCreateTag).toBeDefined();
  });

  it('exports useUpdateTag', async () => {
    const { useUpdateTag } = await import('../hooks');
    expect(useUpdateTag).toBeDefined();
  });

  it('exports useDeleteTag', async () => {
    const { useDeleteTag } = await import('../hooks');
    expect(useDeleteTag).toBeDefined();
  });
});
