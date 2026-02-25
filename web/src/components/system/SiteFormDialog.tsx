import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';

import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { Site, CreateSiteDTO, UpdateSiteDTO } from '@/lib/system-api';

const siteSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().regex(/^[a-z0-9_]{3,50}$/, 'Only lowercase letters, numbers, and underscores (3-50 chars)'),
  domain: z.string(),
  description: z.string(),
  default_locale: z.string(),
  timezone: z.string(),
  is_active: z.boolean(),
});

type SiteFormData = z.infer<typeof siteSchema>;

interface SiteFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateSiteDTO | UpdateSiteDTO) => void;
  loading: boolean;
  site?: Site;
}

const LOCALES = [
  { value: 'en', label: 'English' },
  { value: 'zh-CN', label: '中文 (简体)' },
  { value: 'ja', label: '日本語' },
  { value: 'ko', label: '한국어' },
  { value: 'es', label: 'Español' },
  { value: 'fr', label: 'Français' },
  { value: 'de', label: 'Deutsch' },
];

const TIMEZONES = [
  'UTC',
  'America/New_York',
  'America/Chicago',
  'America/Denver',
  'America/Los_Angeles',
  'Europe/London',
  'Europe/Paris',
  'Europe/Berlin',
  'Asia/Shanghai',
  'Asia/Tokyo',
  'Asia/Seoul',
  'Asia/Singapore',
  'Australia/Sydney',
];

export function SiteFormDialog({
  open,
  onOpenChange,
  onSubmit,
  loading,
  site,
}: SiteFormDialogProps) {
  const { t } = useTranslation();
  const isEdit = !!site;

  const {
    register,
    handleSubmit,
    reset,
    setValue,
    watch,
    formState: { errors },
  } = useForm<SiteFormData>({
    resolver: zodResolver(siteSchema),
    defaultValues: {
      name: '',
      slug: '',
      domain: '',
      description: '',
      default_locale: 'en',
      timezone: 'UTC',
      is_active: true,
    },
  });

  useEffect(() => {
    if (site) {
      reset({
        name: site.name,
        slug: site.slug,
        domain: site.domain || '',
        description: site.description || '',
        default_locale: site.default_locale,
        timezone: site.timezone,
        is_active: site.is_active,
      });
    } else {
      reset({
        name: '',
        slug: '',
        domain: '',
        description: '',
        default_locale: 'en',
        timezone: 'UTC',
        is_active: true,
      });
    }
  }, [site, reset]);

  const watchLocale = watch('default_locale');
  const watchTimezone = watch('timezone');
  const watchIsActive = watch('is_active');

  const onFormSubmit = (data: SiteFormData) => {
    if (isEdit) {
      const updateData: UpdateSiteDTO = {
        name: data.name,
        domain: data.domain || undefined,
        description: data.description || undefined,
        default_locale: data.default_locale,
        timezone: data.timezone,
        is_active: data.is_active,
      };
      onSubmit(updateData);
    } else {
      const createData: CreateSiteDTO = {
        name: data.name,
        slug: data.slug,
        domain: data.domain || undefined,
        description: data.description || undefined,
        default_locale: data.default_locale,
        timezone: data.timezone,
      };
      onSubmit(createData);
    }
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t('system.sites.editSite') : t('system.sites.newSite')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="site-name">{t('system.sites.siteName')}</Label>
            <Input
              id="site-name"
              aria-label={t('system.sites.siteName')}
              {...register('name')}
              autoFocus
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="site-slug">{t('system.sites.slug')}</Label>
            <Input
              id="site-slug"
              aria-label={t('system.sites.slug')}
              {...register('slug')}
              disabled={isEdit}
            />
            {errors.slug && (
              <p className="text-sm text-destructive">{errors.slug.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="site-domain">{t('system.sites.domain')}</Label>
            <Input
              id="site-domain"
              aria-label={t('system.sites.domain')}
              {...register('domain')}
              placeholder="example.com"
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="site-description">{t('system.sites.description')}</Label>
            <Input
              id="site-description"
              aria-label={t('system.sites.description')}
              {...register('description')}
            />
          </div>

          <div className="space-y-2">
            <Label>{t('system.sites.locale')}</Label>
            <Select
              value={watchLocale}
              onValueChange={(v) => setValue('default_locale', v)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {LOCALES.map((l) => (
                  <SelectItem key={l.value} value={l.value}>
                    {l.label}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          <div className="space-y-2">
            <Label>{t('system.sites.timezone')}</Label>
            <Select
              value={watchTimezone}
              onValueChange={(v) => setValue('timezone', v)}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {TIMEZONES.map((tz) => (
                  <SelectItem key={tz} value={tz}>
                    {tz}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>

          {isEdit && (
            <div className="flex items-center gap-3">
              <Switch
                id="site-active"
                aria-label={t('system.sites.active')}
                checked={watchIsActive}
                onCheckedChange={(checked) => setValue('is_active', !!checked)}
              />
              <Label htmlFor="site-active">{t('system.sites.active')}</Label>
            </div>
          )}

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
