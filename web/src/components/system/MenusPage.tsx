import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { MenusTable } from './MenusTable';
import { MenuFormDialog } from './MenuFormDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { siteMenusApi } from '@/lib/system-api';
import type { SiteMenu, CreateSiteMenuDTO, UpdateSiteMenuDTO } from '@/lib/system-api';

function MenusPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [formOpen, setFormOpen] = useState(false);
  const [editingMenu, setEditingMenu] = useState<SiteMenu | undefined>();
  const [deleteMenu, setDeleteMenu] = useState<SiteMenu | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['site-menus'],
    queryFn: () => siteMenusApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateSiteMenuDTO) => siteMenusApi.create(data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menus'] });
      setFormOpen(false);
      setEditingMenu(undefined);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateSiteMenuDTO }) =>
      siteMenusApi.update(id, data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menus'] });
      setFormOpen(false);
      setEditingMenu(undefined);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => siteMenusApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menus'] });
      setDeleteMenu(null);
    },
  });

  const handleNewMenu = useCallback(() => {
    setEditingMenu(undefined);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((menu: SiteMenu) => {
    setEditingMenu(menu);
    setFormOpen(true);
  }, []);

  const handleManageItems = useCallback((menu: SiteMenu) => {
    window.location.href = `/dashboard/menus/${menu.id}/items`;
  }, []);

  const handleFormSubmit = useCallback(
    async (data: CreateSiteMenuDTO | UpdateSiteMenuDTO) => {
      if (editingMenu) {
        await updateMutation.mutateAsync({ id: editingMenu.id, data });
      } else {
        await createMutation.mutateAsync(data as CreateSiteMenuDTO);
      }
    },
    [editingMenu, updateMutation, createMutation],
  );

  const menus = data?.data ?? [];

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.menus.title')}</h1>
      <MenusTable
        menus={menus}
        loading={isLoading}
        onEdit={handleEdit}
        onManageItems={handleManageItems}
        onDelete={(menu) => setDeleteMenu(menu)}
        onNewMenu={handleNewMenu}
      />
      <MenuFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleFormSubmit}
        menu={editingMenu}
        loading={createMutation.isPending || updateMutation.isPending}
      />
      <ConfirmDialog
        open={deleteMenu !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteMenu(null);
        }}
        title={t('content.confirmDelete')}
        description={t('system.menus.deleteMenuConfirm', { name: deleteMenu?.name ?? '' })}
        onConfirm={() => {
          if (deleteMenu) deleteMutation.mutate(deleteMenu.id);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function MenusPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <MenusPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
