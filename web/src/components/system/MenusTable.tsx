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
import type { SiteMenu } from '@/lib/system-api';

interface MenusTableProps {
  menus: SiteMenu[];
  loading: boolean;
  onEdit: (menu: SiteMenu) => void;
  onManageItems: (menu: SiteMenu) => void;
  onDelete: (menu: SiteMenu) => void;
  onNewMenu: () => void;
}

const locationColors: Record<string, string> = {
  header: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  footer: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  sidebar: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
  custom: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
};

export function MenusTable({
  menus,
  loading,
  onEdit,
  onManageItems,
  onDelete,
  onNewMenu,
}: MenusTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<SiteMenu, unknown>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: t('system.menus.menuName'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'slug',
        header: t('system.menus.slug'),
        cell: ({ row }) => (
          <code className="text-sm text-muted-foreground">{row.original.slug}</code>
        ),
      },
      {
        accessorKey: 'location',
        header: t('system.menus.location'),
        cell: ({ row }) => {
          const loc = row.original.location;
          const locationKey = `location${loc.charAt(0).toUpperCase()}${loc.slice(1)}` as string;
          return (
            <Badge variant="outline" className={locationColors[loc] ?? ''}>
              {t(`system.menus.${locationKey}`)}
            </Badge>
          );
        },
      },
      {
        id: 'item_count',
        header: t('system.menus.items'),
        cell: ({ row }) => (
          <span className="text-muted-foreground">
            {t('system.menus.itemCount', { count: row.original.item_count })}
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
              <DropdownMenuItem onClick={() => onManageItems(row.original)}>
                {t('system.menus.manageItems')}
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
    [t, onEdit, onManageItems, onDelete],
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-end">
        <Button onClick={onNewMenu} aria-label={t('system.menus.newMenu')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.menus.newMenu')}
        </Button>
      </div>

      <DataTable
        columns={columns}
        data={menus}
        loading={loading}
        emptyMessage={t('system.menus.noMenusFound')}
      />
    </div>
  );
}
