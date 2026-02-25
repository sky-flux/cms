import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Plus, Shield } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuTrigger,
} from '@/components/ui/dropdown-menu';
import type { Role } from '@/lib/system-api';

interface RolesTableProps {
  roles: Role[];
  loading: boolean;
  onEdit: (role: Role) => void;
  onPermissions: (role: Role) => void;
  onDelete: (id: string) => void;
  onNewRole: () => void;
}

export function RolesTable({
  roles,
  loading,
  onEdit,
  onPermissions,
  onDelete,
  onNewRole,
}: RolesTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<Role>[] = useMemo(
    () => [
      {
        accessorKey: 'name',
        header: t('system.roles.roleName'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'slug',
        header: t('system.roles.roleSlug'),
        cell: ({ row }) => (
          <code className="rounded bg-muted px-1.5 py-0.5 text-sm">{row.original.slug}</code>
        ),
      },
      {
        accessorKey: 'description',
        header: t('system.roles.description'),
        cell: ({ row }) => (
          <span className="text-muted-foreground">{row.original.description}</span>
        ),
      },
      {
        id: 'type',
        header: '',
        cell: ({ row }) => (
          <Badge
            variant="outline"
            className={
              row.original.built_in
                ? 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300'
                : 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300'
            }
          >
            {row.original.built_in
              ? t('system.roles.builtIn')
              : t('system.roles.custom')}
          </Badge>
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
              <DropdownMenuItem onClick={() => onPermissions(row.original)}>
                <Shield className="mr-2 h-4 w-4" />
                {t('system.roles.permissions')}
              </DropdownMenuItem>
              {!row.original.built_in && (
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => onDelete(row.original.id)}
                >
                  {t('common.delete')}
                </DropdownMenuItem>
              )}
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [t, onEdit, onPermissions, onDelete],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.roles.noRolesFound')}</p>
    </div>
  );

  return (
    <div className="space-y-4">
      {/* Header with New Role button */}
      <div className="flex items-center justify-end">
        <Button onClick={onNewRole} aria-label={t('system.roles.newRole')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.roles.newRole')}
        </Button>
      </div>

      {/* Table */}
      {!loading && roles.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={roles}
          loading={loading}
        />
      )}
    </div>
  );
}
