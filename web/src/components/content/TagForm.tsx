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
import type { Tag, CreateTagDTO, UpdateTagDTO } from '@/lib/content-api';

const tagSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().regex(/^[a-z0-9-]+$/, 'Slug must be lowercase letters, numbers, and hyphens').min(1, 'Slug is required'),
});

type TagFormValues = z.infer<typeof tagSchema>;

interface TagFormProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateTagDTO | UpdateTagDTO) => void | Promise<void>;
  tag?: Tag;
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

export function TagForm({
  open,
  onOpenChange,
  onSubmit,
  tag,
  loading = false,
}: TagFormProps) {
  const { t } = useTranslation();
  const isEdit = !!tag;
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(false);

  const form = useForm<TagFormValues>({
    resolver: zodResolver(tagSchema),
    defaultValues: {
      name: tag?.name ?? '',
      slug: tag?.slug ?? '',
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({
        name: tag?.name ?? '',
        slug: tag?.slug ?? '',
      });
      setSlugManuallyEdited(isEdit);
    }
  }, [open, tag, isEdit, form]);

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
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t('content.editTag') : t('content.addTag')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="tag-name">{t('content.tagName')}</Label>
            <Input
              id="tag-name"
              aria-label={t('content.tagName')}
              {...form.register('name')}
              placeholder="Tag name"
            />
            {form.formState.errors.name && (
              <p className="text-sm text-destructive">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="tag-slug">{t('content.tagSlug')}</Label>
            <Input
              id="tag-slug"
              aria-label={t('content.tagSlug')}
              {...form.register('slug', {
                onChange: () => setSlugManuallyEdited(true),
              })}
              placeholder="tag-slug"
            />
            {form.formState.errors.slug && (
              <p className="text-sm text-destructive">
                {form.formState.errors.slug.message}
              </p>
            )}
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
