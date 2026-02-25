import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { UsersTable } from './UsersTable';
import { UserFormDialog } from './UserFormDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { usersApi, rolesApi } from '@/lib/system-api';
import type { User, CreateUserDTO, UpdateUserDTO } from '@/lib/system-api';
import { useDebounce } from '@/hooks/use-debounce';
import { usePagination } from '@/hooks/use-pagination';

function UsersPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [roleFilter, setRoleFilter] = useState<string>('');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  const [dialogOpen, setDialogOpen] = useState(false);
  const [editingUser, setEditingUser] = useState<User | undefined>(undefined);
  const [deleteId, setDeleteId] = useState<string | null>(null);
  const [deleteName, setDeleteName] = useState<string>('');

  const queryParams = {
    page,
    per_page: perPage,
    ...(roleFilter && roleFilter !== 'all' ? { role: roleFilter } : {}),
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['users', queryParams],
    queryFn: () => usersApi.list(queryParams),
  });

  const { data: rolesData } = useQuery({
    queryKey: ['roles'],
    queryFn: () => rolesApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateUserDTO) => usersApi.create(data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDialogOpen(false);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateUserDTO }) => usersApi.update(id, data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDialogOpen(false);
      setEditingUser(undefined);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => usersApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['users'] });
      setDeleteId(null);
    },
  });

  const handleRoleFilter = useCallback(
    (role: string) => {
      setRoleFilter(role);
      resetPage();
    },
    [resetPage],
  );

  const handleSearch = useCallback(
    (query: string) => {
      setSearch(query);
      resetPage();
    },
    [resetPage],
  );

  const handleNewUser = useCallback(() => {
    setEditingUser(undefined);
    setDialogOpen(true);
  }, []);

  const handleEdit = useCallback((user: User) => {
    setEditingUser(user);
    setDialogOpen(true);
  }, []);

  const handleDelete = useCallback((id: string) => {
    const user = users.find((u) => u.id === id);
    setDeleteName(user?.display_name ?? '');
    setDeleteId(id);
  }, []);

  const handleFormSubmit = useCallback(
    (formData: CreateUserDTO | UpdateUserDTO) => {
      if (editingUser) {
        updateMutation.mutate({ id: editingUser.id, data: formData as UpdateUserDTO });
      } else {
        createMutation.mutate(formData as CreateUserDTO);
      }
    },
    [editingUser, createMutation, updateMutation],
  );

  const users = data?.data ?? [];
  const roles = rolesData?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.users.title')}</h1>
      <UsersTable
        users={users}
        roles={roles}
        pagination={pagination}
        loading={isLoading}
        onPageChange={setPage}
        onRoleFilter={handleRoleFilter}
        onSearch={handleSearch}
        onEdit={handleEdit}
        onDelete={handleDelete}
        onNewUser={handleNewUser}
      />
      <UserFormDialog
        open={dialogOpen}
        onOpenChange={(open) => {
          setDialogOpen(open);
          if (!open) setEditingUser(undefined);
        }}
        onSubmit={handleFormSubmit}
        roles={roles}
        loading={createMutation.isPending || updateMutation.isPending}
        user={editingUser}
      />
      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title={t('content.confirmDelete')}
        description={t('system.users.deleteUserConfirm', { name: deleteName })}
        onConfirm={() => {
          if (deleteId) deleteMutation.mutate(deleteId);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function UsersPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <UsersPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
