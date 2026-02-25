import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { ApiKeysTable } from './ApiKeysTable';
import { CreateApiKeyDialog } from './CreateApiKeyDialog';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import {
  apiKeysApi,
  type ApiKey,
  type CreateApiKeyDTO,
  type CreateApiKeyResponse,
} from '@/lib/system-api';

function ApiKeysPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [createDialogOpen, setCreateDialogOpen] = useState(false);
  const [createdKey, setCreatedKey] = useState<CreateApiKeyResponse | null>(null);
  const [revokeKey, setRevokeKey] = useState<ApiKey | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['api-keys'],
    queryFn: () => apiKeysApi.list(),
  });

  const createMutation = useMutation({
    mutationFn: (data: CreateApiKeyDTO) => apiKeysApi.create(data),
    onSuccess: (result) => {
      setCreatedKey(result.data);
    },
  });

  const revokeMutation = useMutation({
    mutationFn: (id: string) => apiKeysApi.delete(id),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['api-keys'] });
      setRevokeKey(null);
    },
  });

  const handleNewKey = useCallback(() => {
    setCreatedKey(null);
    setCreateDialogOpen(true);
  }, []);

  const handleCreateSubmit = useCallback(
    (data: CreateApiKeyDTO) => {
      createMutation.mutate(data);
    },
    [createMutation],
  );

  const handleAcknowledge = useCallback(() => {
    setCreatedKey(null);
    setCreateDialogOpen(false);
    queryClient.invalidateQueries({ queryKey: ['api-keys'] });
  }, [queryClient]);

  const handleRevoke = useCallback((key: ApiKey) => {
    setRevokeKey(key);
  }, []);

  const apiKeys = data?.data ?? [];

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.apiKeys.title')}</h1>

      <ApiKeysTable
        apiKeys={apiKeys}
        loading={isLoading}
        onRevoke={handleRevoke}
        onNewKey={handleNewKey}
      />

      <CreateApiKeyDialog
        open={createDialogOpen}
        onOpenChange={(open) => {
          if (!open && !createdKey) setCreateDialogOpen(false);
        }}
        onSubmit={handleCreateSubmit}
        loading={createMutation.isPending}
        createdKey={createdKey}
        onAcknowledge={handleAcknowledge}
      />

      <ConfirmDialog
        open={revokeKey !== null}
        onOpenChange={(open) => {
          if (!open) setRevokeKey(null);
        }}
        title={t('system.apiKeys.revokeKey')}
        description={
          revokeKey
            ? t('system.apiKeys.revokeConfirm', { name: revokeKey.name })
            : ''
        }
        onConfirm={() => {
          if (revokeKey) revokeMutation.mutate(revokeKey.id);
        }}
        loading={revokeMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function ApiKeysPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <ApiKeysPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
