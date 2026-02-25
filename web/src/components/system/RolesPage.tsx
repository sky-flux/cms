import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { RolesTable } from './RolesTable';
import { RoleFormDialog } from './RoleFormDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { rolesApi } from '@/lib/system-api';
import type { Role, CreateRoleDTO, UpdateRoleDTO } from '@/lib/system-api';

function RolesPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingRole, setEditingRole] = useState<Role | undefined>(undefined);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteName, setDeleteName] = useState<string>('');

  const { data, isLoading } = useQuery({
    queryKey: ['roles'],
    queryFn: () => rolesApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateRoleDTO) => rolesApi.create(data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setDialogOpen(false);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateRoleDTO }) => rolesApi.update(id, data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setDialogOpen(false);
      setEditingRole(undefined);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => rolesApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['roles'] });
      setDeleteId(null);
    },
  });

  const handleNewRole = useCallback(() => {
    setEditingRole(undefined);
    setDialogOpen(true);
  }, []);

  const handleEdit = useCallback((role: Role) => {
    setEditingRole(role);
    setDialogOpen(true);
  }, []);

  const handlePermissions = useCallback((role: Role) => {
    window.location.href = `/dashboard/roles/${role.id}/permissions`;
  }, []);

  const handleDelete = useCallback((id: string) => {
    const role = roles.find((r) => r.id === id);
    if (role?.built_in) {
      toast.error(t('system.roles.cannotDeleteBuiltIn'));
      return;
    }
    setDeleteName(role?.name ?? '');
    setDeleteId(id);
  }, []);

  const handleFormSubmit = useCallback(
    (formData: CreateRoleDTO | UpdateRoleDTO) => {
      if (editingRole) {
        updateMutation.mutate({ id: editingRole.id, data: formData as UpdateRoleDTO });
      } else {
        createMutation.mutate(formData as CreateRoleDTO);
      }
    },
    [editingRole, createMutation, updateMutation],
  );

  const roles = data?.data ?? [];

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.roles.title')}</h1>
      <RolesTable
        roles={roles}
        loading={isLoading}
        onEdit={handleEdit}
        onPermissions={handlePermissions}
        onDelete={handleDelete}
        onNewRole={handleNewRole}
      />
      <RoleFormDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open);
          if (!open) setEditingRole(undefined);
        }}
        onSubmit={handleFormSubmit}
        loading={createMutation.isPending || updateMutation.isPending}
        role={editingRole}
      />
      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title={t('content.confirmDelete')}
        description={t('system.roles.deleteRoleConfirm', { name: deleteName })}
        onConfirm={() => {
          if (deleteId) deleteMutation.mutate(deleteId);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function RolesPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <RolesPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
