import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { MediaLibrary } from './MediaLibrary';
import { MediaUploader, type UploadingFile } from './MediaUploader';
import { MediaDetailDialog } from './MediaDetailDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { usePagination } from '@/hooks/use-pagination';
import { useDebounce } from '@/hooks/use-debounce';
import {
  mediaApi,
  type MediaFile,
  type MediaFileDetail,
  type UpdateMediaDTO,
} from '@/lib/content-api';

type ViewMode = 'grid' | 'list';

export function MediaPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage } = usePagination();

  const [viewMode, setViewMode] = useState<ViewMode>('grid');
  const [searchValue, setSearchValue] = useState('');
  const debouncedSearch = useDebounce(searchValue, 300);
  const [selectedIds, setSelectedIds] = useState<string[]>([]);
  const [uploadingFiles, setUploadingFiles] = useState<UploadingFile[]>([]);

  // Detail dialog state
  const [detailDialogOpen, setDetailDialogOpen] = useState(false);
  const [selectedMediaId, setSelectedMediaId] = useState<string | null>(null);

  // Batch delete confirm
  const [batchDeleteOpen, setBatchDeleteOpen] = useState(false);
  const [batchDeleteIds, setBatchDeleteIds] = useState<string[]>([]);

  const queryKey = ['media', { page, perPage, q: debouncedSearch }];

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () =>
      mediaApi.list({
        page,
        per_page: perPage,
        q: debouncedSearch || undefined,
      }),
  });

  const media = data?.data ?? [];
  const pagination = data?.pagination;

  // Detail query (only when dialog is open)
  const { data: detailData } = useQuery({
    queryKey: ['media', selectedMediaId],
    queryFn: () => mediaApi.get(selectedMediaId!),
    enabled: !!selectedMediaId && detailDialogOpen,
  });

  const mediaDetail: MediaFileDetail | null = detailData?.data ?? null;

  const updateMetaMutation = useMutation({
    mutationFn: ({ id, data }: { id: string; data: UpdateMediaDTO }) =>
      mediaApi.updateMeta(id, data),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media'] });
      toast.success(t('messages.updateSuccess'));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: ({ id, force }: { id: string; force: boolean }) =>
      mediaApi.delete(id, force),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media'] });
      setDetailDialogOpen(false);
      setSelectedMediaId(null);
      toast.success(t('messages.deleteSuccess'));
    },
  });

  const batchDeleteMutation = useMutation({
    mutationFn: ({ ids, force }: { ids: string[]; force: boolean }) =>
      mediaApi.batchDelete(ids, force),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media'] });
      setSelectedIds([]);
      setBatchDeleteOpen(false);
      setBatchDeleteIds([]);
      toast.success(t('messages.deleteSuccess'));
    },
  });

  const handleUpload = useCallback(
    async (files: File[]) => {
      const newUploading = files.map((f) => ({ name: f.name, progress: 0 }));
      setUploadingFiles(newUploading);

      for (let i = 0; i < files.length; i++) {
        try {
          setUploadingFiles((prev) =>
            prev.map((f, idx) => (idx === i ? { ...f, progress: 50 } : f)),
          );
          await mediaApi.upload(files[i]);
          setUploadingFiles((prev) =>
            prev.map((f, idx) => (idx === i ? { ...f, progress: 100 } : f)),
          );
        } catch {
          toast.error(`Failed to upload ${files[i].name}`);
        }
      }

      // Clear upload progress after a brief delay
      setTimeout(() => setUploadingFiles([]), 1500);
      queryClient.invalidateQueries({ queryKey: ['media'] });
    },
    [queryClient],
  );

  const handleItemClick = useCallback((item: MediaFile) => {
    setSelectedMediaId(item.id);
    setDetailDialogOpen(true);
  }, []);

  const handleSaveMeta = useCallback(
    async (data: UpdateMediaDTO) => {
      if (selectedMediaId) {
        await updateMetaMutation.mutateAsync({ id: selectedMediaId, data });
      }
    },
    [selectedMediaId, updateMetaMutation],
  );

  const handleDelete = useCallback(
    (id: string, force: boolean) => {
      deleteMutation.mutate({ id, force });
    },
    [deleteMutation],
  );

  const handleBatchDeleteRequest = useCallback((ids: string[]) => {
    setBatchDeleteIds(ids);
    setBatchDeleteOpen(true);
  }, []);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('content.media')}</h1>
      </div>

      <MediaUploader
        onUpload={handleUpload}
        uploadingFiles={uploadingFiles}
      />

      <MediaLibrary
        media={media}
        viewMode={viewMode}
        onViewModeChange={setViewMode}
        onItemClick={handleItemClick}
        selectedIds={selectedIds}
        onSelectionChange={setSelectedIds}
        searchValue={searchValue}
        onSearchChange={setSearchValue}
        onBatchDelete={handleBatchDeleteRequest}
        loading={isLoading}
        pagination={
          pagination
            ? { page: pagination.page, totalPages: pagination.total_pages }
            : undefined
        }
        onPageChange={setPage}
      />

      <MediaDetailDialog
        open={detailDialogOpen}
        onOpenChange={setDetailDialogOpen}
        media={mediaDetail}
        onSave={handleSaveMeta}
        onDelete={handleDelete}
        loading={updateMetaMutation.isPending}
      />

      <ConfirmDialog
        open={batchDeleteOpen}
        onOpenChange={setBatchDeleteOpen}
        title={t('content.batchDelete')}
        description={t('content.batchDeleteConfirm', { count: batchDeleteIds.length })}
        onConfirm={() => batchDeleteMutation.mutate({ ids: batchDeleteIds, force: false })}
        loading={batchDeleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}
