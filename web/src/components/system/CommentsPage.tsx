import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { CommentsTable } from './CommentsTable';
import { CommentDetailDialog } from './CommentDetailDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { commentsApi } from '@/lib/system-api';
import type { Comment } from '@/lib/system-api';
import { useDebounce } from '@/hooks/use-debounce';
import { usePagination } from '@/hooks/use-pagination';

function CommentsPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage, resetPage } = usePagination();

  const [statusFilter, setStatusFilter] = useState<string>('');
  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search, 300);

  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [detailComment, setDetailComment] = useState<Comment | null>(null);
  const [deleteId, setDeleteId] = useState<string | null>(null);

  const queryParams = {
    page,
    per_page: perPage,
    ...(statusFilter && statusFilter !== 'all' ? { status: statusFilter } : {}),
    ...(debouncedSearch ? { q: debouncedSearch } : {}),
  };

  const { data, isLoading } = useQuery({
    queryKey: ['comments', queryParams],
    queryFn: () => commentsApi.list(queryParams),
  });

  const updateStatusMutation = useMutation({
    mutationFn: ({ id, status }: { id: string; status: string }) =>
      commentsApi.updateStatus(id, status),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['comments'] });
    },
  });

  const togglePinMutation = useMutation({
    mutationFn: ({ id, isPinned }: { id: string; isPinned: boolean }) =>
      commentsApi.togglePin(id, isPinned),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['comments'] });
    },
  });

  const replyMutation = useMutation({
    mutationFn: ({ id, content }: { id: string; content: string }) =>
      commentsApi.reply(id, content),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['comments'] });
      setDetailComment(null);
    },
  });

  const batchStatusMutation = useMutation({
    mutationFn: ({ ids, status }: { ids: string[]; status: string }) =>
      commentsApi.batchStatus(ids, status),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['comments'] });
      setSelectedIds([]);
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => commentsApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['comments'] });
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

  const handleApprove = useCallback(
    (id: string) => updateStatusMutation.mutate({ id, status: 'approved' }),
    [updateStatusMutation],
  );

  const handleReject = useCallback(
    (id: string) => updateStatusMutation.mutate({ id, status: 'trash' }),
    [updateStatusMutation],
  );

  const handleMarkSpam = useCallback(
    (id: string) => updateStatusMutation.mutate({ id, status: 'spam' }),
    [updateStatusMutation],
  );

  const handleTogglePin = useCallback(
    (id: string, isPinned: boolean) => togglePinMutation.mutate({ id, isPinned }),
    [togglePinMutation],
  );

  const handleReply = useCallback(
    (id: string, content: string) => replyMutation.mutate({ id, content }),
    [replyMutation],
  );

  const handleBatchApprove = useCallback(
    () => batchStatusMutation.mutate({ ids: selectedIds, status: 'approved' }),
    [batchStatusMutation, selectedIds],
  );

  const handleBatchReject = useCallback(
    () => batchStatusMutation.mutate({ ids: selectedIds, status: 'trash' }),
    [batchStatusMutation, selectedIds],
  );

  const handleBatchSpam = useCallback(
    () => batchStatusMutation.mutate({ ids: selectedIds, status: 'spam' }),
    [batchStatusMutation, selectedIds],
  );

  const comments = data?.data ?? [];
  const pagination = data?.pagination ?? {
    page: 1,
    per_page: perPage,
    total: 0,
    total_pages: 1,
  };

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.comments.title')}</h1>
      <CommentsTable
        comments={comments}
        pagination={pagination}
        loading={isLoading}
        selectedIds={selectedIds}
        onPageChange={setPage}
        onStatusFilter={handleStatusFilter}
        onSearch={handleSearch}
        onSelectChange={setSelectedIds}
        onApprove={handleApprove}
        onReject={handleReject}
        onMarkSpam={handleMarkSpam}
        onTogglePin={handleTogglePin}
        onReply={(id) => {
          const comment = comments.find((c) => c.id === id);
          if (comment) setDetailComment(comment);
        }}
        onDelete={(id) => setDeleteId(id)}
        onBatchApprove={handleBatchApprove}
        onBatchReject={handleBatchReject}
        onBatchSpam={handleBatchSpam}
        onViewDetail={setDetailComment}
      />
      <CommentDetailDialog
        comment={detailComment}
        open={detailComment !== null}
        onOpenChange={(open) => {
          if (!open) setDetailComment(null);
        }}
        onReply={handleReply}
        replyLoading={replyMutation.isPending}
      />
      <ConfirmDialog
        open={deleteId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteId(null);
        }}
        title={t('common.delete')}
        description={t('system.comments.deleteCommentConfirm')}
        onConfirm={() => {
          if (deleteId) deleteMutation.mutate(deleteId);
        }}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function CommentsPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <CommentsPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
