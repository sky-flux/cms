import { useMemo, useState, useCallback } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import {
  MoreHorizontal,
  Plus,
  Search,
  Upload,
  Download,
  Trash2,
} from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { Redirect, PaginationMeta } from '@/lib/system-api';

interface RedirectsTableProps {
  redirects: Redirect[];
  pagination: PaginationMeta;
  loading: boolean;
  onPageChange: (page: number) => void;
  onSearch: (query: string) => void;
  onStatusCodeFilter: (statusCode: string) => void;
  onEdit: (redirect: Redirect) => void;
  onDelete: (redirect: Redirect) => void;
  onToggleActive: (redirect: Redirect) => void;
  onNewRedirect: () => void;
  onImportCsv: () => void;
  onExportCsv: () => void;
  onBatchDelete: (ids: string[]) => void;
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

export function RedirectsTable({
  redirects,
  pagination,
  loading,
  onPageChange,
  onSearch,
  onStatusCodeFilter,
  onEdit,
  onDelete,
  onToggleActive,
  onNewRedirect,
  onImportCsv,
  onExportCsv,
  onBatchDelete,
}: RedirectsTableProps) {
  const { t } = useTranslation();
  const [selectedIds, setSelectedIds] = useState<Set<string>>(new Set());

  const toggleSelect = useCallback(
    (id: string) => {
      setSelectedIds((prev) => {
        const next = new Set(prev);
        if (next.has(id)) {
          next.delete(id);
        } else {
          next.add(id);
        }
        return next;
      });
    },
    [],
  );

  const toggleSelectAll = useCallback(() => {
    if (selectedIds.size === redirects.length) {
      setSelectedIds(new Set());
    } else {
      setSelectedIds(new Set(redirects.map((r) => r.id)));
    }
  }, [selectedIds.size, redirects]);

  const handleBatchDelete = useCallback(() => {
    onBatchDelete(Array.from(selectedIds));
    setSelectedIds(new Set());
  }, [selectedIds, onBatchDelete]);

  const columns: ColumnDef<Redirect, unknown>[] = useMemo(
    () => [
      {
        id: 'select',
        header: () => (
          <Checkbox
            checked={redirects.length > 0 && selectedIds.size === redirects.length}
            onCheckedChange={toggleSelectAll}
            aria-label="Select all"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={selectedIds.has(row.original.id)}
            onCheckedChange={() => toggleSelect(row.original.id)}
            aria-label={`Select ${row.original.source_path}`}
          />
        ),
      },
      {
        accessorKey: 'source_path',
        header: t('system.redirects.sourcePath'),
        cell: ({ row }) => (
          <code className="font-mono text-sm">{row.original.source_path}</code>
        ),
      },
      {
        accessorKey: 'target_url',
        header: t('system.redirects.targetUrl'),
        cell: ({ row }) => (
          <span className="text-sm truncate max-w-[200px] block">{row.original.target_url}</span>
        ),
      },
      {
        accessorKey: 'status_code',
        header: t('system.redirects.statusCode'),
        cell: ({ row }) => {
          const code = row.original.status_code;
          const color =
            code === 301
              ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
              : 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300';
          return (
            <Badge variant="outline" className={color}>
              {code}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'hit_count',
        header: t('system.redirects.hitCount'),
        cell: ({ row }) => (
          <span className="text-muted-foreground">{row.original.hit_count}</span>
        ),
      },
      {
        id: 'last_hit_at',
        header: t('system.redirects.lastHit'),
        cell: ({ row }) => (
          <span className="text-sm text-muted-foreground">
            {formatDate(row.original.last_hit_at)}
          </span>
        ),
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
              <DropdownMenuItem onClick={() => onEdit(row.original)}>
                {t('common.edit')}
              </DropdownMenuItem>
              <DropdownMenuItem
                variant="destructive"
                onClick={() => onDelete(row.original)}
              >
                {t('common.delete')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [t, selectedIds, redirects, toggleSelectAll, toggleSelect, onEdit, onDelete],
  );

  return (
    <div className="space-y-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <Select onValueChange={onStatusCodeFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t('system.redirects.filterByStatusCode')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.status')}</SelectItem>
            <SelectItem value="301">{t('system.redirects.permanent301')}</SelectItem>
            <SelectItem value="302">{t('system.redirects.temporary302')}</SelectItem>
          </SelectContent>
        </Select>

        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('system.redirects.searchPlaceholder')}
            className="pl-8"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>

        <Button variant="outline" onClick={onImportCsv} aria-label={t('system.redirects.importCsv')}>
          <Upload className="mr-2 h-4 w-4" />
          {t('system.redirects.importCsv')}
        </Button>

        <Button variant="outline" onClick={onExportCsv} aria-label={t('system.redirects.exportCsv')}>
          <Download className="mr-2 h-4 w-4" />
          {t('system.redirects.exportCsv')}
        </Button>

        <Button onClick={onNewRedirect} aria-label={t('system.redirects.newRedirect')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.redirects.newRedirect')}
        </Button>
      </div>

      {/* Batch delete bar */}
      {selectedIds.size > 0 && (
        <div className="flex items-center gap-3 rounded-md border bg-muted/50 p-3">
          <span className="text-sm font-medium">
            {t('system.redirects.selected', { count: selectedIds.size })}
          </span>
          <Button
            variant="destructive"
            size="sm"
            onClick={handleBatchDelete}
          >
            <Trash2 className="mr-2 h-4 w-4" />
            {t('content.batchDelete')}
          </Button>
        </div>
      )}

      {/* Table */}
      <DataTable
        columns={columns}
        data={redirects}
        loading={loading}
        emptyMessage={t('system.redirects.noRedirectsFound')}
        pagination={{
          page: pagination.page,
          totalPages: pagination.total_pages,
        }}
        onPageChange={onPageChange}
      />
    </div>
  );
}
