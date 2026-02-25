import { useState, useMemo, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';
import { Plus } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { CategoryTree } from './CategoryTree';
import { CategoryForm } from './CategoryForm';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import {
  categoriesApi,
  type CategoryNode,
  type CreateCategoryDTO,
  type UpdateCategoryDTO,
  type ReorderItem,
} from '@/lib/content-api';
import { ApiError } from '@/lib/api-client';

function flattenCategories(
  nodes: CategoryNode[],
  depth = 0,
): { id: string; name: string; depth: number }[] {
  const result: { id: string; name: string; depth: number }[] = [];
  for (const node of nodes) {
    result.push({ id: node.id, name: node.name, depth });
    if (node.children.length > 0) {
      result.push(...flattenCategories(node.children, depth + 1));
    }
  }
  return result;
}

export function CategoriesPage() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [formOpen, setFormOpen] = useState(false);
  const [editingCategory, setEditingCategory] = useState<CategoryNode | undefined>();
  const [parentIdForNew, setParentIdForNew] = useState<string | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<CategoryNode | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['categories'],
    queryFn: () => categoriesApi.tree(),
  });

  const categories = data?.data ?? [];
  const parentOptions = useMemo(() => flattenCategories(categories), [categories]);

  const createMutation = useMutation({
    mutationFn: (dto: CreateCategoryDTO) => categoriesApi.create(dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setFormOpen(false);
      toast.success(t('messages.createSuccess'));
    },
  });

  const updateMutation = useMutation({
    mutationFn: ({ id, dto }: { id: string; dto: UpdateCategoryDTO }) =>
      categoriesApi.update(id, dto),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setFormOpen(false);
      toast.success(t('messages.updateSuccess'));
    },
  });

  const deleteMutation = useMutation({
    mutationFn: (id: string) => categoriesApi.delete(id),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
      setDeleteTarget(null);
      toast.success(t('messages.deleteSuccess'));
    },
    onError: (err) => {
      if (err instanceof ApiError && err.status === 409) {
        toast.error(t('content.categoryHasChildren'));
      }
      setDeleteTarget(null);
    },
  });

  const reorderMutation = useMutation({
    mutationFn: (orders: ReorderItem[]) => categoriesApi.reorder(orders),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['categories'] });
    },
  });

  const handleAddRoot = useCallback(() => {
    setEditingCategory(undefined);
    setParentIdForNew(null);
    setFormOpen(true);
  }, []);

  const handleEdit = useCallback((category: CategoryNode) => {
    setEditingCategory(category);
    setParentIdForNew(null);
    setFormOpen(true);
  }, []);

  const handleAddChild = useCallback((parentId: string) => {
    setEditingCategory(undefined);
    setParentIdForNew(parentId);
    setFormOpen(true);
  }, []);

  const handleDelete = useCallback((category: CategoryNode) => {
    setDeleteTarget(category);
  }, []);

  const handleFormSubmit = useCallback(
    async (values: CreateCategoryDTO | UpdateCategoryDTO) => {
      if (editingCategory) {
        await updateMutation.mutateAsync({ id: editingCategory.id, dto: values });
      } else {
        const dto = {
          ...values,
          parent_id: parentIdForNew ?? values.parent_id,
        } as CreateCategoryDTO;
        await createMutation.mutateAsync(dto);
      }
    },
    [editingCategory, parentIdForNew, createMutation, updateMutation],
  );

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h1 className="text-2xl font-bold">{t('content.categories')}</h1>
        <Button onClick={handleAddRoot}>
          <Plus className="h-4 w-4 mr-2" />
          {t('content.addCategory')}
        </Button>
      </div>

      {isLoading ? (
        <div className="space-y-2">
          {Array.from({ length: 3 }).map((_, i) => (
            <div key={i} className="h-12 rounded-md border bg-muted animate-pulse" />
          ))}
        </div>
      ) : (
        <CategoryTree
          categories={categories}
          onEdit={handleEdit}
          onAddChild={handleAddChild}
          onDelete={handleDelete}
          onReorder={(orders) => reorderMutation.mutate(orders)}
        />
      )}

      <CategoryForm
        open={formOpen}
        onOpenChange={setFormOpen}
        onSubmit={handleFormSubmit}
        parentOptions={parentOptions}
        category={editingCategory}
        loading={createMutation.isPending || updateMutation.isPending}
      />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => !open && setDeleteTarget(null)}
        title={t('content.deleteCategory')}
        description={t('content.deleteCategoryConfirm', { name: deleteTarget?.name ?? '' })}
        onConfirm={() => deleteTarget && deleteMutation.mutate(deleteTarget.id)}
        loading={deleteMutation.isPending}
        variant="danger"
      />
    </div>
  );
}
