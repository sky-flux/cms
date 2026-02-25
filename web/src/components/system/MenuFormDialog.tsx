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
import type { SiteMenu, CreateSiteMenuDTO, UpdateSiteMenuDTO } from '@/lib/system-api';

const menuSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z
    .string()
    .min(1, 'Slug is required')
    .regex(/^[a-z0-9-]+$/, 'Slug must be lowercase letters, numbers, and hyphens'),
  location: z.string().min(1),
  description: z.string().optional(),
});

type MenuFormValues = z.infer<typeof menuSchema>;

interface MenuFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateSiteMenuDTO | UpdateSiteMenuDTO) => void | Promise<void>;
  menu?: SiteMenu;
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

export function MenuFormDialog({
  open,
  onOpenChange,
  onSubmit,
  menu,
  loading = false,
}: MenuFormDialogProps) {
  const { t } = useTranslation();
  const isEdit = !!menu;
  const [slugManuallyEdited, setSlugManuallyEdited] = useState(!!menu);

  const form = useForm<MenuFormValues>({
    resolver: zodResolver(menuSchema),
    defaultValues: {
      name: menu?.name ?? '',
      slug: menu?.slug ?? '',
      location: menu?.location ?? 'header',
      description: menu?.description ?? '',
    },
  });

  useEffect(() => {
    if (open) {
      const edited = !!menu;
      setSlugManuallyEdited(edited);
      form.reset({
        name: menu?.name ?? '',
        slug: menu?.slug ?? '',
        location: menu?.location ?? 'header',
        description: menu?.description ?? '',
      });
    }
  }, [open, menu, form]);

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
      location: values.location,
      description: values.description || undefined,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t('system.menus.editMenu') : t('system.menus.newMenu')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="menu-name">{t('system.menus.menuName')}</Label>
            <Input
              id="menu-name"
              aria-label={t('system.menus.menuName')}
              {...form.register('name')}
              placeholder="Main Navigation"
            />
            {form.formState.errors.name && (
              <p className="text-sm text-destructive">
                {form.formState.errors.name.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="menu-slug">{t('system.menus.slug')}</Label>
            <Input
              id="menu-slug"
              aria-label={t('system.menus.slug')}
              {...form.register('slug', {
                onChange: () => setSlugManuallyEdited(true),
              })}
              placeholder="main-nav"
            />
            {form.formState.errors.slug && (
              <p className="text-sm text-destructive">
                {form.formState.errors.slug.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label>{t('system.menus.location')}</Label>
            <Select
              value={form.watch('location')}
              onValueChange={(val) => form.setValue('location', val)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="header">{t('system.menus.locationHeader')}</SelectItem>
                <SelectItem value="footer">{t('system.menus.locationFooter')}</SelectItem>
                <SelectItem value="sidebar">{t('system.menus.locationSidebar')}</SelectItem>
                <SelectItem value="custom">{t('system.menus.locationCustom')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label htmlFor="menu-description">{t('system.menus.description')}</Label>
            <Input
              id="menu-description"
              aria-label={t('system.menus.description')}
              {...form.register('description')}
              placeholder="Optional description"
            />
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
