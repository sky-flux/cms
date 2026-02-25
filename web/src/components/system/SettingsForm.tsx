import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Save } from 'lucide-react';

import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Card } from '@/components/ui/card';
import type { SettingItem } from '@/lib/system-api';

interface SettingsFormProps {
  settings: SettingItem[];
  onSave: (key: string, value: string) => void;
  savingKey: string | null;
}

export function SettingsForm({ settings, onSave, savingKey }: SettingsFormProps) {
  const { t } = useTranslation();
  const [localValues, setLocalValues] = useState<Record<string, string>>({});

  if (settings.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 py-8">
        <p className="text-muted-foreground">{t('system.settings.noSettings')}</p>
      </div>
    );
  }

  const getValue = (item: SettingItem) => {
    return localValues[item.key] ?? item.value;
  };

  const handleChange = (key: string, value: string) => {
    setLocalValues((prev) => ({ ...prev, [key]: value }));
  };

  const handleSave = (key: string) => {
    const value = localValues[key];
    if (value !== undefined) {
      onSave(key, value);
    } else {
      const item = settings.find((s) => s.key === key);
      if (item) onSave(key, item.value);
    }
  };

  return (
    <div className="space-y-4">
      {settings.map((item) => (
        <Card key={item.key} className="p-4">
          <div className="flex items-start justify-between gap-4">
            <div className="flex-1 space-y-2">
              <div className="flex items-center gap-2">
                <code className="text-sm font-semibold bg-muted px-1.5 py-0.5 rounded">
                  {item.key}
                </code>
              </div>
              {item.description && (
                <p className="text-sm text-muted-foreground">{item.description}</p>
              )}
              <Input
                value={getValue(item)}
                onChange={(e) => handleChange(item.key, e.target.value)}
              />
            </div>
            <Button
              size="sm"
              onClick={() => handleSave(item.key)}
              disabled={savingKey === item.key}
              aria-label={t('common.save')}
            >
              {savingKey === item.key ? (
                t('common.loading')
              ) : (
                <>
                  <Save className="mr-1 h-3 w-3" />
                  {t('common.save')}
                </>
              )}
            </Button>
          </div>
        </Card>
      ))}
    </div>
  );
}
