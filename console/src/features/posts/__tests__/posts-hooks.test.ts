import { describe, it, expect, vi } from 'vitest';

describe('posts hooks', () => {
  it('exports usePosts', async () => {
    const { usePosts } = await import('../hooks');
    expect(usePosts).toBeDefined();
  });

  it('exports usePost', async () => {
    const { usePost } = await import('../hooks');
    expect(usePost).toBeDefined();
  });

  it('exports useCreatePost', async () => {
    const { useCreatePost } = await import('../hooks');
    expect(useCreatePost).toBeDefined();
  });

  it('exports useUpdatePost', async () => {
    const { useUpdatePost } = await import('../hooks');
    expect(useUpdatePost).toBeDefined();
  });

  it('exports useDeletePost', async () => {
    const { useDeletePost } = await import('../hooks');
    expect(useDeletePost).toBeDefined();
  });

  it('exports usePublishPost', async () => {
    const { usePublishPost } = await import('../hooks');
    expect(usePublishPost).toBeDefined();
  });
});
