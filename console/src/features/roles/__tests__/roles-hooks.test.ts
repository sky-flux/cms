import { describe, it, expect } from 'vitest';

describe('roles hooks', () => {
  it('exports useRoles', async () => {
    const { useRoles } = await import('../hooks');
    expect(useRoles).toBeDefined();
  });

  it('exports useCreateRole', async () => {
    const { useCreateRole } = await import('../hooks');
    expect(useCreateRole).toBeDefined();
  });

  it('exports useUpdateRole', async () => {
    const { useUpdateRole } = await import('../hooks');
    expect(useUpdateRole).toBeDefined();
  });

  it('exports useDeleteRole', async () => {
    const { useDeleteRole } = await import('../hooks');
    expect(useDeleteRole).toBeDefined();
  });
});

describe('roles components', () => {
  it('exports RolesTable', async () => {
    const { RolesTable } = await import('../components');
    expect(RolesTable).toBeDefined();
  });

  it('exports RoleFormDialog', async () => {
    const { RoleFormDialog } = await import('../components');
    expect(RoleFormDialog).toBeDefined();
  });

  it('exports RolePermissions', async () => {
    const { RolePermissions } = await import('../components');
    expect(RolePermissions).toBeDefined();
  });
});
