import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { TagsTable } from './TagsTable';
import { TagForm } from './TagForm';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { usePagination } from '@/hooks/use-pagination';
import { useDebounce } from '@/hooks/use-debounce';
import {
  tagsApi,
  type Tag,
  type CreateTagDTO,
  type UpdateTagDTO,
} from '@/lib/content-api';

export function TagsPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const { page, perPage, setPage } = usePagination();

  const [searchValue, setSearchValue] = useState('');
  const debouncedSearch = useDebounce(searchValue, 300);

  const [formOpen, setFormOpen] = useState(false);
  const [editingTag, setEditingTag] = useState<Tag | undefined>();
  const [deleteTarget, setDeleteTarget] = useState<Tag | null>(null);

  const queryKey = ['tags', { page, perPage, q: debouncedSearch }];

  const { data, isLoading } = useQuery({
    queryKey,
    queryFn: () =>
      tagsApi.list({
        page,
        per_page: perPage,
        q: debouncedSearch || undefined,
      }),
  });

  const tags = data?.data ?? [];
  const pagination = data?.pagination;

  const createMutation = useMutation({
    mutationFn: (dto: CreateTagDTO) => tagsApi.create(dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      setFormOpen(false);
      toast.success(t('messages.createSuccess'));
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, dto }: { id: string; dto: UpdateTagDTO }) =>
      tagsApi.update(id, dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      setFormOpen(false);
      toast.success(t('messages.updateSuccess'));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => tagsApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['tags'] });
      setDeleteTarget(null);
      toast.success(t('messages.deleteSuccess'));
    },
  });

  const handleAdd = useCallback(() => {
    setEditingTag(undefined);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((tag: Tag) => {
    setEditingTag(tag);
    setFormOpen(true);
  }, []);

  const handleDelete = useCallback((tag: Tag) => {
    setDeleteTarget(tag);
  }, []);

  const handleFormSubmit = useCallback(
    async (values: CreateTagDTO | UpdateTagDTO) => {
      if (editingTag) {
        await updateMutation.mutateAsync({ id: editingTag.id, dto: values });
      } else {
        await createMutation.mutateAsync(values as CreateTagDTO);
      }
    },
    [editingTag, createMutation, updateMutation],
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('content.tags')}</h1>
        <Button onClick={handleAdd}>
          <Plus className="h-4 w-4 mr-2" />
          {t('content.addTag')}
        </Button>
      </div>

      <TagsTable
        tags={tags}
        onEdit={handleEdit}
        onDelete={handleDelete}
        searchValue={searchValue}
        onSearchChange={setSearchValue}
        loading={isLoading}
        pagination={
          pagination
            ? { page: pagination.page, totalPages: pagination.total_pages }
            : undefined
        }
        onPageChange={setPage}
      />

      <TagForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleFormSubmit}
        tag={editingTag}
        loading={createMutation.isPending || updateMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title={t('content.deleteTag')}
        description={t('content.deleteTagConfirm', { name: deleteTarget?.name ?? '' })}
        onConfirm={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}
