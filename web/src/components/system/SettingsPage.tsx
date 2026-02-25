import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { SettingsForm } from './SettingsForm';
import { settingsApi } from '@/lib/system-api';

function SettingsPageInner() {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [savingKey, setSavingKey] = useState<string | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['settings'],
    queryFn: () => settingsApi.get(),
  });

  const updateMutation = useMutation({
    mutationFn: ({ key, value }: { key: string; value: string }) =>
      settingsApi.update(key, value),
    onSuccess: () => {
      toast.success(t('messages.saveSuccess'));
      queryClient.invalidateQueries({ queryKey: ['settings'] });
      setSavingKey(null);
    },
    onError: () => {
      setSavingKey(null);
    },
  });

  const handleSave = useCallback(
    (key: string, value: string) => {
      setSavingKey(key);
      updateMutation.mutate({ key, value });
    },
    [updateMutation],
  );

  const settings = data?.data ?? [];

  return (
    <div className="p-6">
      <h1 className="mb-6 text-2xl font-bold">{t('system.settings.title')}</h1>
      {isLoading ? (
        <p className="text-muted-foreground">{t('common.loading')}</p>
      ) : (
        <SettingsForm
          settings={settings}
          onSave={handleSave}
          savingKey={savingKey}
        />
      )}
    </div>
  );
}

export function SettingsPage() {
  return (
    <QueryProvider>
      <I18nProvider>
        <SettingsPageInner />
      </I18nProvider>
    </QueryProvider>
  );
}
