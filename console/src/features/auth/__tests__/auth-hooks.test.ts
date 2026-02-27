import { describe, it, expect } from 'vitest';

describe('auth hooks', () => {
  it('exports useLogin', async () => {
    const { useLogin } = await import('../hooks');
    expect(useLogin).toBeDefined();
  });

  it('exports useLogout', async () => {
    const { useLogout } = await import('../hooks');
    expect(useLogout).toBeDefined();
  });

  it('exports useMe', async () => {
    const { useMe } = await import('../hooks');
    expect(useMe).toBeDefined();
  });
});
