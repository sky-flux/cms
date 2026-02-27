import { describe, it, expect, vi } from 'vitest';

describe('apiClient', () => {
  it('should export ApiError class', async () => {
    const { ApiError } = await import('../api-client');
    expect(() => {
      new ApiError('test', 500);
    }).not.toThrow();
  });

  it('should have get, post, put, patch, delete methods', async () => {
    const { apiClient } = await import('../api-client');
    expect(typeof apiClient.get).toBe('function');
    expect(typeof apiClient.post).toBe('function');
    expect(typeof apiClient.put).toBe('function');
    expect(typeof apiClient.patch).toBe('function');
    expect(typeof apiClient.delete).toBe('function');
  });
});
