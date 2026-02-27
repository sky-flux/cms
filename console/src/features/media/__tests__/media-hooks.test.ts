import { describe, it, expect } from 'vitest';

describe('media hooks', () => {
  it('exports useMediaFiles', async () => {
    const { useMediaFiles } = await import('../hooks');
    expect(useMediaFiles).toBeDefined();
  });

  it('exports useUploadMedia', async () => {
    const { useUploadMedia } = await import('../hooks');
    expect(useUploadMedia).toBeDefined();
  });

  it('exports useDeleteMedia', async () => {
    const { useDeleteMedia } = await import('../hooks');
    expect(useDeleteMedia).toBeDefined();
  });
});

describe('media components', () => {
  it('exports MediaLibrary', async () => {
    const { MediaLibrary } = await import('../components');
    expect(MediaLibrary).toBeDefined();
  });

  it('exports MediaUploader', async () => {
    const { MediaUploader } = await import('../components');
    expect(MediaUploader).toBeDefined();
  });
});
