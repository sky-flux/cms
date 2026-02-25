import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';
import { Copy, Check } from 'lucide-react';

import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogDescription,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Checkbox } from '@/components/ui/checkbox';
import type { CreateApiKeyDTO, CreateApiKeyResponse } from '@/lib/system-api';

const createKeySchema = z.object({
  name: z.string().min(1, 'Name is required'),
  expires_at: z.string(),
  rate_limit: z.number().min(1),
});

type CreateKeyFormData = z.infer<typeof createKeySchema>;

interface CreateApiKeyDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateApiKeyDTO) => void;
  loading: boolean;
  createdKey: CreateApiKeyResponse | null;
  onAcknowledge: () => void;
}

export function CreateApiKeyDialog({
  open,
  onOpenChange,
  onSubmit,
  loading,
  createdKey,
  onAcknowledge,
}: CreateApiKeyDialogProps) {
  const { t } = useTranslation();
  const [copied, setCopied] = useState(false);
  const [acknowledged, setAcknowledged] = useState(false);

  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<CreateKeyFormData>({
    resolver: zodResolver(createKeySchema),
    defaultValues: {
      name: '',
      expires_at: '',
      rate_limit: 100,
    },
  });

  const onFormSubmit = (data: CreateKeyFormData) => {
    onSubmit({
      name: data.name,
      expires_at: data.expires_at || undefined,
      rate_limit: data.rate_limit,
    });
  };

  const handleCopy = async () => {
    if (!createdKey) return;
    try {
      await navigator.clipboard.writeText(createdKey.key);
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    } catch {
      // Fallback for environments without clipboard API
    }
  };

  const handleDone = () => {
    setAcknowledged(false);
    setCopied(false);
    onAcknowledge();
  };

  // Phase 2: Key has been created, show the key
  if (createdKey) {
    return (
      <Dialog open={open} onOpenChange={() => {/* prevent close without acknowledging */}}>
        <DialogContent showCloseButton={false}>
          <DialogHeader>
            <DialogTitle>{t('system.apiKeys.keyCreated')}</DialogTitle>
            <DialogDescription>
              {t('system.apiKeys.keyCreatedDescription')}
            </DialogDescription>
          </DialogHeader>

          <div className="space-y-4">
            <div className="rounded-md bg-muted p-3">
              <code className="text-sm font-mono break-all">{createdKey.key}</code>
            </div>

            <Button
              variant="outline"
              size="sm"
              onClick={handleCopy}
              aria-label={t('system.apiKeys.copyKey')}
              className="w-full"
            >
              {copied ? (
                <>
                  <Check className="mr-2 h-4 w-4" />
                  {t('system.apiKeys.keyCopied')}
                </>
              ) : (
                <>
                  <Copy className="mr-2 h-4 w-4" />
                  {t('system.apiKeys.copyKey')}
                </>
              )}
            </Button>

            <div className="flex items-center gap-2">
              <Checkbox
                id="acknowledge-key"
                checked={acknowledged}
                onCheckedChange={(checked) => setAcknowledged(!!checked)}
              />
              <Label htmlFor="acknowledge-key" className="text-sm cursor-pointer">
                {t('system.apiKeys.keyCreatedDescription')}
              </Label>
            </div>
          </div>

          <DialogFooter>
            <Button
              onClick={handleDone}
              disabled={!acknowledged}
              aria-label="done"
            >
              Done
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>
    );
  }

  // Phase 1: Create form
  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('system.apiKeys.newKey')}</DialogTitle>
        </DialogHeader>
        <form onSubmit={handleSubmit(onFormSubmit)} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="key-name">{t('system.apiKeys.keyName')}</Label>
            <Input
              id="key-name"
              aria-label={t('system.apiKeys.keyName')}
              {...register('name')}
              autoFocus
            />
            {errors.name && (
              <p className="text-sm text-destructive">{errors.name.message}</p>
            )}
          </div>

          <div className="space-y-2">
            <Label htmlFor="key-expires">{t('system.apiKeys.expiresAt')}</Label>
            <Input
              id="key-expires"
              type="date"
              {...register('expires_at')}
            />
          </div>

          <div className="space-y-2">
            <Label htmlFor="key-rate-limit">{t('system.apiKeys.rateLimit')}</Label>
            <Input
              id="key-rate-limit"
              type="number"
              min={1}
              {...register('rate_limit')}
            />
            <p className="text-xs text-muted-foreground">
              {t('system.apiKeys.rateLimitHelp')}
            </p>
          </div>

          <DialogFooter>
            <Button
              type="button"
              variant="outline"
              onClick={() => onOpenChange(false)}
            >
              {t('common.cancel')}
            </Button>
            <Button type="submit" disabled={loading} aria-label={t('common.create')}>
              {loading ? t('common.loading') : t('common.create')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
