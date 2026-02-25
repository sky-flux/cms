import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { RedirectsTable } from './RedirectsTable';
import { RedirectFormDialog } from './RedirectFormDialog';
import { CsvImportDialog } from './CsvImportDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { redirectsApi } from '@/lib/system-api';
import { useDebounce } from '@/hooks/use-debounce';
import { usePagination } from '@/hooks/use-pagination';
import type { Redirect, CreateRedirectDTO, UpdateRedirectDTO } from '@/lib/system-api';

function RedirectsPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [statusCodeFilter, setStatusCodeFilter] = useState<string>('');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  const [formOpen, setFormOpen] = useState(false);
  const [editingRedirect, setEditingRedirect] = useState<Redirect | undefined>();
  const [deleteRedirect, setDeleteRedirect] = useState<Redirect | null>(null);
  const [csvImportOpen, setCsvImportOpen] = useState(false);
  const [batchDeleteIds, setBatchDeleteIds] = useState<string[]>([]);

  const queryParams = {
    page,
    per_page: perPage,
    ...(statusCodeFilter && statusCodeFilter !== 'all' ? { status_code: Number(statusCodeFilter) } : {}),
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['redirects', queryParams],
    queryFn: () => redirectsApi.list(queryParams),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateRedirectDTO) => redirectsApi.create(data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['redirects'] });
      setFormOpen(false);
      setEditingRedirect(undefined);
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateRedirectDTO }) =>
      redirectsApi.update(id, data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['redirects'] });
      setFormOpen(false);
      setEditingRedirect(undefined);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => redirectsApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['redirects'] });
      setDeleteRedirect(null);
    },
  });

  const batchDeleteMutation = useMutation({
    mutationFn: (ids: string[]) => redirectsApi.batchDelete(ids),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['redirects'] });
      setBatchDeleteIds([]);
    },
  });

  const importMutation = useMutation({
    mutationFn: (file: File) => redirectsApi.import(file),
    onSuccess: (result) => {
      const data = result.data;
      toast.success(
        t('system.redirects.csvImportResult', {
          imported: data.imported,
          skipped: data.skipped,
          errors: data.errors.length,
        }),
      );
      queryClient.invalidateQueries({ queryKey: ['redirects'] });
      setCsvImportOpen(false);
    },
  });

  const handleNewRedirect = useCallback(() => {
    setEditingRedirect(undefined);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((redirect: Redirect) => {
    setEditingRedirect(redirect);
    setFormOpen(true);
  }, []);

  const handleFormSubmit = useCallback(
    async (data: CreateRedirectDTO | UpdateRedirectDTO) => {
      if (editingRedirect) {
        await updateMutation.mutateAsync({ id: editingRedirect.id, data });
      } else {
        await createMutation.mutateAsync(data as CreateRedirectDTO);
      }
    },
    [editingRedirect, updateMutation, createMutation],
  );

  const handleStatusCodeFilter = useCallback(
    (code: string) => {
      setStatusCodeFilter(code);
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

  const handleExportCsv = useCallback(async () => {
    try {
      const blob = await redirectsApi.export();
      const url = URL.createObjectURL(blob as unknown as Blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = 'redirects.csv';
      document.body.appendChild(a);
      a.click();
      document.body.removeChild(a);
      URL.revokeObjectURL(url);
    } catch {
      toast.error(t('errors.serverError'));
    }
  }, [t]);

  const handleBatchDelete = useCallback(
    (ids: string[]) => {
      setBatchDeleteIds(ids);
    },
    [],
  );

  const redirects = data?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.redirects.title')}</h1>
      <RedirectsTable
        redirects={redirects}
        pagination={pagination}
        loading={isLoading}
        onPageChange={setPage}
        onSearch={handleSearch}
        onStatusCodeFilter={handleStatusCodeFilter}
        onEdit={handleEdit}
        onDelete={(r) => setDeleteRedirect(r)}
        onToggleActive={(r) =>
          updateMutation.mutate({
            id: r.id,
            data: { is_active: !r.is_active },
          })
        }
        onNewRedirect={handleNewRedirect}
        onImportCsv={() => setCsvImportOpen(true)}
        onExportCsv={handleExportCsv}
        onBatchDelete={handleBatchDelete}
      />
      <RedirectFormDialog
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleFormSubmit}
        redirect={editingRedirect}
        loading={createMutation.isPending || updateMutation.isPending}
      />
      <CsvImportDialog
        open={csvImportOpen}
        onOpenChange={setCsvImportOpen}
        onImport={(file) => importMutation.mutate(file)}
        loading={importMutation.isPending}
      />
      <ConfirmDialog
        open={deleteRedirect !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteRedirect(null);
        }}
        title={t('content.confirmDelete')}
        description={t('system.redirects.deleteRedirectConfirm')}
        onConfirm={() => {
          if (deleteRedirect) deleteMutation.mutate(deleteRedirect.id);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
      <ConfirmDialog
        open={batchDeleteIds.length > 0}
        onOpenChange={(open) => {
          if (!open) setBatchDeleteIds([]);
        }}
        title={t('content.confirmDelete')}
        description={t('system.redirects.batchDeleteConfirm', { count: batchDeleteIds.length })}
        onConfirm={() => batchDeleteMutation.mutate(batchDeleteIds)}
        loading={batchDeleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function RedirectsPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <RedirectsPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
