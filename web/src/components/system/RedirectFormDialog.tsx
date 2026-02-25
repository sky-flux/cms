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
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { Redirect, CreateRedirectDTO, UpdateRedirectDTO } from '@/lib/system-api';

const redirectSchema = z.object({
  source_path: z
    .string()
    .min(1, 'Source path is required')
    .refine((val) => val.startsWith('/'), 'Must start with /')
    .refine((val) => !val.includes('?'), 'Must not contain query parameters (?)'),
  target_url: z.string().min(1, 'Target URL is required'),
  status_code: z.number().int(),
  is_active: z.boolean(),
});

type RedirectFormValues = z.infer<typeof redirectSchema>;

interface RedirectFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateRedirectDTO | UpdateRedirectDTO) => void | Promise<void>;
  redirect?: Redirect;
  loading?: boolean;
}

export function RedirectFormDialog({
  open,
  onOpenChange,
  onSubmit,
  redirect,
  loading = false,
}: RedirectFormDialogProps) {
  const { t } = useTranslation();
  const isEdit = !!redirect;

  const form = useForm<RedirectFormValues>({
    resolver: zodResolver(redirectSchema),
    defaultValues: {
      source_path: redirect?.source_path ?? '',
      target_url: redirect?.target_url ?? '',
      status_code: redirect?.status_code ?? 301,
      is_active: redirect?.is_active ?? true,
    },
  });

  useEffect(() => {
    if (open) {
      form.reset({
        source_path: redirect?.source_path ?? '',
        target_url: redirect?.target_url ?? '',
        status_code: redirect?.status_code ?? 301,
        is_active: redirect?.is_active ?? true,
      });
    }
  }, [open, redirect, form]);

  const handleSubmit = form.handleSubmit(async (values) => {
    await onSubmit({
      source_path: values.source_path,
      target_url: values.target_url,
      status_code: values.status_code,
      is_active: values.is_active,
    });
  });

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>
            {isEdit ? t('system.redirects.editRedirect') : t('system.redirects.newRedirect')}
          </DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="redirect-source">{t('system.redirects.sourcePath')}</Label>
            <Input
              id="redirect-source"
              aria-label={t('system.redirects.sourcePath')}
              {...form.register('source_path')}
              placeholder="/old-page"
            />
            {form.formState.errors.source_path && (
              <p className="text-sm text-destructive">
                {form.formState.errors.source_path.message}
              </p>
            )}
            <p className="text-xs text-muted-foreground">
              {t('system.redirects.sourcePathHelp')}
            </p>
          </div>

          <div className="space-y-2">
            <Label htmlFor="redirect-target">{t('system.redirects.targetUrl')}</Label>
            <Input
              id="redirect-target"
              aria-label={t('system.redirects.targetUrl')}
              {...form.register('target_url')}
              placeholder="https://example.com/new-page"
            />
            {form.formState.errors.target_url && (
              <p className="text-sm text-destructive">
                {form.formState.errors.target_url.message}
              </p>
            )}
          </div>

          <div className="space-y-2">
            <Label>{t('system.redirects.statusCode')}</Label>
            <Select
              value={String(form.watch('status_code'))}
              onValueChange={(val) => form.setValue('status_code', Number(val))}
            >
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="301">{t('system.redirects.permanent301')}</SelectItem>
                <SelectItem value="302">{t('system.redirects.temporary302')}</SelectItem>
              </SelectContent>
            </Select>
          </div>

          <div className="flex items-center gap-2">
            <Switch
              checked={form.watch('is_active')}
              onCheckedChange={(checked) => form.setValue('is_active', checked)}
            />
            <Label>{t('system.redirects.active')}</Label>
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
