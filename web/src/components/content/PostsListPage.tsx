import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { PostsTable } from './PostsTable';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { postsApi } from '@/lib/content-api';
import { useDebounce } from '@/hooks/use-debounce';
import { usePagination } from '@/hooks/use-pagination';

function PostsListInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [statusFilter, setStatusFilter] = useState<string>('');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  const [deleteId, setDeleteId] = useState<string | null>(null);

  const queryParams = {
    page,
    per_page: perPage,
    ...(statusFilter && statusFilter !== 'all' ? { status: statusFilter } : {}),
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['posts', queryParams],
    queryFn: () => postsApi.list(queryParams),
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => postsApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['posts'] });
      setDeleteId(null);
    },
  });

  const handleStatusFilter = useCallback(
    (status: string) => {
      setStatusFilter(status);
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

  const handleNewPost = useCallback(() => {
    window.location.href = '/dashboard/posts/new';
  }, []);

  const posts = data?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('content.posts')}</h1>
      <PostsTable
        posts={posts}
        pagination={pagination}
        loading={isLoading}
        onPageChange={setPage}
        onStatusFilter={handleStatusFilter}
        onSearch={handleSearch}
        onDelete={(id) => setDeleteId(id)}
        onNewPost={handleNewPost}
      />
      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title={t('content.confirmDelete')}
        description={t('content.deleteWarning')}
        onConfirm={() => {
          if (deleteId) deleteMutation.mutate(deleteId);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function PostsListPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <PostsListInner />
      </I18nProvider>
    </QueryProvider>
  );
}
