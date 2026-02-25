import { useEffect, useState } from 'react';
import { useForm } from 'react-hook-form';
import { z } from 'zod';
import { zodResolver } from '@hookform/resolvers/zod';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { CategoryNode, CreateCategoryDTO, UpdateCategoryDTO } from '@/lib/content-api';

const categorySchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().regex(/^[a-z0-9-]+$/, 'Slug must be lowercase letters, numbers, and hyphens').min(1, 'Slug is required'),
  description: z.string().optional(),
  parent_id: z.string().nullable().optional(),
});

type CategoryFormValues = z.infer<typeof categorySchema>;

interface ParentOption {
  id: string;
  name: string;
  depth: number;
}

interface CategoryFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateCategoryDTO | UpdateCategoryDTO) => void | Promise<void>;
  parentOptions: ParentOption[];
  category?: CategoryNode;
  loading?: boolean;
}

function slugify(str: string): string {
  return str
    .toLowerCase()
    .trim()
    .replace(/[^\w\s-]/g, '')
    .replace(/[\s_]+/g, '-')
    .replace(/-+/g, '-')
    .replace(/^-|-$/g, '');
}

export function CategoryForm({
  open,
  onOpenChange,
  onSubmit,
  parentOptions,
  category,
  loading = false,
}: CategoryFormProps) {
  const { t } = useTranslation();
  const isEdit = !!category;
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

  const form = useForm<CategoryFormValues>({
    resolver: zodResolver(categorySchema),
    defaultValues: {
      name: category?.name ?? '',
      slug: category?.slug ?? '',
      description: category?.description ?? '',
      parent_id: category?.parent_id ?? null,
    },
  });

  // Reset form when dialog opens/closes or category changes
  useEffect(() => {
    if (open) {
      form.reset({
        name: category?.name ?? '',
        slug: category?.slug ?? '',
        description: category?.description ?? '',
        parent_id: category?.parent_id ?? null,
      });
      setSlugManuallyEdited(isEdit);
    }
  }, [open, category, isEdit, form]);

  // Auto-generate slug from name
  const watchName = form.watch('name');
  useEffect(() => {
    if (!slugManuallyEdited && watchName) {
      form.setValue('slug', slugify(watchName));
    }
  }, [watchName, slugManuallyEdited, form]);

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit({
      name: values.name,
      slug: values.slug,
      description: values.description || undefined,
      parent_id: values.parent_id || undefined,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t('content.editCategory') : t('content.addCategory')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="category-name">{t('content.categoryName')}</Label>
            <Input
              id="category-name"
              aria-label={t('content.categoryName')}
              {...form.register('name')}
              placeholder="Category name"
            />
            {form.formState.errors.name && (
              <p className="text-sm text-destructive">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="category-slug">{t('content.categorySlug')}</Label>
            <Input
              id="category-slug"
              aria-label={t('content.categorySlug')}
              {...form.register('slug', {
                onChange: () => setSlugManuallyEdited(true),
              })}
              placeholder="category-slug"
            />
            {form.formState.errors.slug && (
              <p className="text-sm text-destructive">
                {form.formState.errors.slug.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="category-description">{t('content.categoryDescription')}</Label>
            <textarea
              id="category-description"
              aria-label={t('content.categoryDescription')}
              className="flex min-h-[60px] w-full rounded-md border border-input bg-transparent px-3 py-2 text-sm shadow-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-1 focus-visible:ring-ring disabled:cursor-not-allowed disabled:opacity-50"
              {...form.register('description')}
              placeholder="Optional description"
            />
          </div>

          <div className="space-y-2">
            <Label>{t('content.categoryParent')}</Label>
            <Select
              value={form.watch('parent_id') ?? '__none__'}
              onValueChange={(value) => form.setValue('parent_id', value === '__none__' ? null : value)}
            >
              <SelectTrigger>
                <SelectValue placeholder={t('content.categoryNone')} />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="__none__">{t('content.categoryNone')}</SelectItem>
                {parentOptions
                  .filter((opt) => opt.id !== category?.id)
                  .map((opt) => (
                    <SelectItem key={opt.id} value={opt.id}>
                      {'—'.repeat(opt.depth)} {opt.name}
                    </SelectItem>
                  ))}
              </SelectContent>
            </Select>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={loading}>
              {loading ? t('common.loading') : t('common.save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
