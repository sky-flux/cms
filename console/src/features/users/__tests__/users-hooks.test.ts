import { describe, it, expect } from 'vitest';

describe('users hooks', () => {
  it('exports useUsers', async () => {
    const { useUsers } = await import('../hooks');
    expect(useUsers).toBeDefined();
  });

  it('exports useCreateUser', async () => {
    const { useCreateUser } = await import('../hooks');
    expect(useCreateUser).toBeDefined();
  });

  it('exports useUpdateUser', async () => {
    const { useUpdateUser } = await import('../hooks');
    expect(useUpdateUser).toBeDefined();
  });

  it('exports useDeleteUser', async () => {
    const { useDeleteUser } = await import('../hooks');
    expect(useDeleteUser).toBeDefined();
  });
});

describe('users components', () => {
  it('exports UsersTable', async () => {
    const { UsersTable } = await import('../components');
    expect(UsersTable).toBeDefined();
  });

  it('exports UserFormDialog', async () => {
    const { UserFormDialog } = await import('../components');
    expect(UserFormDialog).toBeDefined();
  });
});
