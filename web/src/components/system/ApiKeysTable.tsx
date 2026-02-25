import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Plus } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { ApiKey } from '@/lib/system-api';

interface ApiKeysTableProps {
  apiKeys: ApiKey[];
  loading: boolean;
  onRevoke: (key: ApiKey) => void;
  onNewKey: () => void;
}

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '--';
  const date = new Date(dateStr);
  return date.toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function ApiKeysTable({
  apiKeys,
  loading,
  onRevoke,
  onNewKey,
}: ApiKeysTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<ApiKey>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: t('system.apiKeys.keyName'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'key_prefix',
        header: t('system.apiKeys.keyPrefix'),
        cell: ({ row }) => (
          <code className="text-xs bg-muted px-1.5 py-0.5 rounded font-mono">
            {row.original.key_prefix}
          </code>
        ),
      },
      {
        id: 'status',
        header: t('system.apiKeys.status'),
        cell: ({ row }) => (
          <Badge
            variant="outline"
            className={
              row.original.is_active
                ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                : 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300'
            }
          >
            {row.original.is_active
              ? t('system.apiKeys.active')
              : t('system.apiKeys.revoked')}
          </Badge>
        ),
      },
      {
        id: 'last_used_at',
        header: t('system.apiKeys.lastUsed'),
        cell: ({ row }) =>
          row.original.last_used_at
            ? formatDate(row.original.last_used_at)
            : t('system.apiKeys.never'),
      },
      {
        id: 'expires_at',
        header: t('system.apiKeys.expiresAt'),
        cell: ({ row }) =>
          row.original.expires_at
            ? formatDate(row.original.expires_at)
            : t('system.apiKeys.noExpiry'),
      },
      {
        accessorKey: 'rate_limit',
        header: t('system.apiKeys.rateLimit'),
        cell: ({ row }) => String(row.original.rate_limit),
      },
      {
        id: 'actions',
        header: t('common.actions'),
        cell: ({ row }) => (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" aria-label={t('common.actions')}>
                <MoreHorizontal className="h-4 w-4" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end">
              {row.original.is_active && (
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => onRevoke(row.original)}
                >
                  {t('system.apiKeys.revokeKey')}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [t, onRevoke],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.apiKeys.noKeysFound')}</p>
      <Button variant="outline" onClick={onNewKey}>
        {t('system.apiKeys.newKey')}
      </Button>
    </div>
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <Button onClick={onNewKey} aria-label={t('system.apiKeys.newKey')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.apiKeys.newKey')}
        </Button>
      </div>

      {!loading && apiKeys.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={apiKeys}
          loading={loading}
        />
      )}
    </div>
  );
}
