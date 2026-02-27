import { describe, it, expect } from 'vitest';

describe('sites hooks', () => {
  it('exports useSites', async () => {
    const { useSites } = await import('../hooks');
    expect(useSites).toBeDefined();
  });

  it('exports useCreateSite', async () => {
    const { useCreateSite } = await import('../hooks');
    expect(useCreateSite).toBeDefined();
  });

  it('exports useUpdateSite', async () => {
    const { useUpdateSite } = await import('../hooks');
    expect(useUpdateSite).toBeDefined();
  });

  it('exports useDeleteSite', async () => {
    const { useDeleteSite } = await import('../hooks');
    expect(useDeleteSite).toBeDefined();
  });
});

describe('sites components', () => {
  it('exports SitesTable', async () => {
    const { SitesTable } = await import('../components');
    expect(SitesTable).toBeDefined();
  });

  it('exports SiteFormDialog', async () => {
    const { SiteFormDialog } = await import('../components');
    expect(SiteFormDialog).toBeDefined();
  });
});
