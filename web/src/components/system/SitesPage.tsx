import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { SitesTable } from './SitesTable';
import { SiteFormDialog } from './SiteFormDialog';
import { SiteUsersDialog } from './SiteUsersDialog';
import {
  sitesApi,
  type Site,
  type CreateSiteDTO,
  type UpdateSiteDTO,
  type SiteUser,
} from '@/lib/system-api';
import { useDebounce } from '@/hooks/use-debounce';
import { usePagination } from '@/hooks/use-pagination';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';

import {
  AlertDialog,
  AlertDialogContent,
  AlertDialogDescription,
  AlertDialogFooter,
  AlertDialogHeader,
  AlertDialogTitle,
  AlertDialogCancel,
  AlertDialogAction,
} from '@/components/ui/alert-dialog';

interface SitesPageProps {
  initialSiteSlug?: string;
}

function SitesPageInner({ initialSiteSlug }: SitesPageProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  // Dialog states
  const [formOpen, setFormOpen] = useState(false);
  const [editSite, setEditSite] = useState<Site | undefined>(undefined);
  const [usersDialogOpen, setUsersDialogOpen] = useState(false);
  const [selectedSite, setSelectedSite] = useState<Site | null>(null);
  const [deleteSite, setDeleteSite] = useState<Site | null>(null);
  const [deleteSlugInput, setDeleteSlugInput] = useState('');

  const queryParams = {
    page,
    per_page: perPage,
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['sites', queryParams],
    queryFn: () => sitesApi.list(queryParams),
  });

  // Site users query
  const { data: siteUsersData, isLoading: siteUsersLoading } = useQuery({
    queryKey: ['site-users', selectedSite?.slug],
    queryFn: () =>
      selectedSite
        ? sitesApi.listUsers(selectedSite.slug, { page: 1, per_page: 100 })
        : Promise.resolve({ data: [] as SiteUser[], pagination: { page: 1, per_page: 100, total: 0, total_pages: 1 } }),
    enabled: !!selectedSite && usersDialogOpen,
  });

  // Mutations
  const createMutation = useMutation({
    mutationFn: (data: CreateSiteDTO) => sitesApi.create(data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['sites'] });
      setFormOpen(false);
    },
  });

  const updateMutation = useMutation({
    mutationFn: (params: { slug: string; data: UpdateSiteDTO }) =>
      sitesApi.update(params.slug, params.data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['sites'] });
      setFormOpen(false);
      setEditSite(undefined);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (params: { slug: string; confirmSlug: string }) =>
      sitesApi.deleteSite(params.slug, params.confirmSlug),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['sites'] });
      setDeleteSite(null);
      setDeleteSlugInput('');
    },
  });

  const assignRoleMutation = useMutation({
    mutationFn: (params: { slug: string; userId: string; role: string }) =>
      sitesApi.assignRole(params.slug, params.userId, params.role),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-users'] });
    },
  });

  const removeRoleMutation = useMutation({
    mutationFn: (params: { slug: string; userId: string }) =>
      sitesApi.removeRole(params.slug, params.userId),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-users'] });
    },
  });

  const handleSearch = useCallback(
    (query: string) => {
      setSearch(query);
      resetPage();
    },
    [resetPage],
  );

  const handleEdit = useCallback((site: Site) => {
    setEditSite(site);
    setFormOpen(true);
  }, []);

  const handleManageUsers = useCallback((site: Site) => {
    setSelectedSite(site);
    setUsersDialogOpen(true);
  }, []);

  const handleDelete = useCallback((site: Site) => {
    setDeleteSite(site);
    setDeleteSlugInput('');
  }, []);

  const handleNewSite = useCallback(() => {
    setEditSite(undefined);
    setFormOpen(true);
  }, []);

  const handleFormSubmit = useCallback(
    (data: CreateSiteDTO | UpdateSiteDTO) => {
      if (editSite) {
        updateMutation.mutate({ slug: editSite.slug, data: data as UpdateSiteDTO });
      } else {
        createMutation.mutate(data as CreateSiteDTO);
      }
    },
    [editSite, updateMutation, createMutation],
  );

  const handleAssignRole = useCallback(
    (userId: string, role: string) => {
      if (!selectedSite) return;
      assignRoleMutation.mutate({
        slug: selectedSite.slug,
        userId,
        role,
      });
    },
    [selectedSite, assignRoleMutation],
  );

  const handleRemoveUser = useCallback(
    (userId: string) => {
      if (!selectedSite) return;
      removeRoleMutation.mutate({ slug: selectedSite.slug, userId });
    },
    [selectedSite, removeRoleMutation],
  );

  const sites = data?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  // If initialSiteSlug provided, auto-open user management
  useState(() => {
    if (initialSiteSlug && sites.length > 0) {
      const found = sites.find((s) => s.slug === initialSiteSlug);
      if (found) {
        handleManageUsers(found);
      }
    }
  });

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.sites.title')}</h1>

      <SitesTable
        sites={sites}
        pagination={pagination}
        loading={isLoading}
        onPageChange={setPage}
        onSearch={handleSearch}
        onEdit={handleEdit}
        onManageUsers={handleManageUsers}
        onDelete={handleDelete}
        onNewSite={handleNewSite}
      />

      {/* Create / Edit dialog */}
      <SiteFormDialog
        open={formOpen}
        onOpenChange={(open) => {
          setFormOpen(open);
          if (!open) setEditSite(undefined);
        }}
        onSubmit={handleFormSubmit}
        loading={createMutation.isPending || updateMutation.isPending}
        site={editSite}
      />

      {/* Site users dialog */}
      <SiteUsersDialog
        open={usersDialogOpen}
        onOpenChange={(open) => {
          setUsersDialogOpen(open);
          if (!open) setSelectedSite(null);
        }}
        siteUsers={siteUsersData?.data ?? []}
        loading={siteUsersLoading}
        onAssignRole={handleAssignRole}
        onRemoveUser={handleRemoveUser}
        assignLoading={assignRoleMutation.isPending}
      />

      {/* Delete confirmation with slug typing */}
      <AlertDialog
        open={deleteSite !== null}
        onOpenChange={(open) => {
          if (!open) {
            setDeleteSite(null);
            setDeleteSlugInput('');
          }
        }}
      >
        <AlertDialogContent>
          <AlertDialogHeader>
            <AlertDialogTitle>{t('content.confirmDelete')}</AlertDialogTitle>
            <AlertDialogDescription>
              {deleteSite &&
                t('system.sites.deleteSiteConfirm', { name: deleteSite.name })}
            </AlertDialogDescription>
          </AlertDialogHeader>
          <div className="space-y-2">
            <p className="text-sm font-medium">{t('system.sites.confirmSlug')}</p>
            <Input
              placeholder={deleteSite?.slug ?? ''}
              value={deleteSlugInput}
              onChange={(e) => setDeleteSlugInput(e.target.value)}
            />
          </div>
          <AlertDialogFooter>
            <AlertDialogCancel>{t('common.cancel')}</AlertDialogCancel>
            <AlertDialogAction
              disabled={
                !deleteSite ||
                deleteSlugInput !== deleteSite.slug ||
                deleteMutation.isPending
              }
              onClick={() => {
                if (deleteSite) {
                  deleteMutation.mutate({
                    slug: deleteSite.slug,
                    confirmSlug: deleteSlugInput,
                  });
                }
              }}
              className="bg-destructive text-white hover:bg-destructive/90"
            >
              {deleteMutation.isPending
                ? t('common.loading')
                : t('common.confirm')}
            </AlertDialogAction>
          </AlertDialogFooter>
        </AlertDialogContent>
      </AlertDialog>
    </div>
  );
}

export function SitesPage(props: SitesPageProps = {}) {
  return (
    <QueryProvider>
      <I18nProvider>
        <SitesPageInner {...props} />
      </I18nProvider>
    </QueryProvider>
  );
}
