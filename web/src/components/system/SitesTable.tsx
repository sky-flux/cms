import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Plus, Search } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { Site, PaginationMeta } from '@/lib/system-api';

interface SitesTableProps {
  sites: Site[];
  pagination: PaginationMeta;
  loading: boolean;
  onPageChange: (page: number) => void;
  onSearch: (query: string) => void;
  onEdit: (site: Site) => void;
  onManageUsers: (site: Site) => void;
  onDelete: (site: Site) => void;
  onNewSite: () => void;
}

export function SitesTable({
  sites,
  pagination,
  loading,
  onPageChange,
  onSearch,
  onEdit,
  onManageUsers,
  onDelete,
  onNewSite,
}: SitesTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<Site>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: t('system.sites.siteName'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'slug',
        header: t('system.sites.slug'),
        cell: ({ row }) => (
          <code className="text-xs bg-muted px-1.5 py-0.5 rounded">
            {row.original.slug}
          </code>
        ),
      },
      {
        accessorKey: 'domain',
        header: t('system.sites.domain'),
        cell: ({ row }) => row.original.domain || '--',
      },
      {
        id: 'status',
        header: t('system.sites.status'),
        cell: ({ row }) => (
          <Badge
            variant="outline"
            className={
              row.original.is_active
                ? 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300'
                : 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300'
            }
          >
            {row.original.is_active
              ? t('system.sites.active')
              : t('system.sites.inactive')}
          </Badge>
        ),
      },
      {
        accessorKey: 'timezone',
        header: t('system.sites.timezone'),
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
              <DropdownMenuItem onClick={() => onManageUsers(row.original)}>
                {t('system.sites.manageUsers')}
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
    [t, onEdit, onManageUsers, onDelete],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.sites.noSitesFound')}</p>
      <Button variant="outline" onClick={onNewSite}>
        {t('system.sites.newSite')}
      </Button>
    </div>
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-3">
        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('system.sites.searchPlaceholder')}
            className="pl-8"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>
        <Button onClick={onNewSite} aria-label={t('system.sites.newSite')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.sites.newSite')}
        </Button>
      </div>

      {!loading && sites.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={sites}
          loading={loading}
          pagination={{
            page: pagination.page,
            totalPages: pagination.total_pages,
          }}
          onPageChange={onPageChange}
        />
      )}
    </div>
  );
}
