import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Plus, Search } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
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
import type { User, Role, PaginationMeta } from '@/lib/system-api';

interface UsersTableProps {
  users: User[];
  roles: Role[];
  pagination: PaginationMeta;
  loading: boolean;
  onPageChange: (page: number) => void;
  onRoleFilter: (role: string) => void;
  onSearch: (query: string) => void;
  onEdit: (user: User) => void;
  onDelete: (id: string) => void;
  onNewUser: () => void;
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

export function UsersTable({
  users,
  roles,
  pagination,
  loading,
  onPageChange,
  onRoleFilter,
  onSearch,
  onEdit,
  onDelete,
  onNewUser,
}: UsersTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<User>[] = useMemo(
    () => [
      {
        accessorKey: 'email',
        header: t('system.users.email'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.email}</span>
        ),
      },
      {
        accessorKey: 'display_name',
        header: t('system.users.displayName'),
      },
      {
        accessorKey: 'role',
        header: t('system.users.role'),
        cell: ({ row }) => (
          <Badge variant="outline">{row.original.role}</Badge>
        ),
      },
      {
        id: 'status',
        header: t('system.users.status'),
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
              ? t('system.users.active')
              : t('system.users.disabled')}
          </Badge>
        ),
      },
      {
        id: 'last_login_at',
        header: t('system.users.lastLogin'),
        cell: ({ row }) => formatDate(row.original.last_login_at),
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
                onClick={() => onDelete(row.original.id)}
              >
                {t('common.delete')}
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        ),
      },
    ],
    [t, onEdit, onDelete],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.users.noUsersFound')}</p>
    </div>
  );

  return (
    <div className="space-y-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <Select onValueChange={onRoleFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder={t('system.users.filterByRole')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('system.users.filterByRole')}</SelectItem>
            {roles.map((role) => (
              <SelectItem key={role.id} value={role.slug}>
                {role.name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>

        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('system.users.searchPlaceholder')}
            className="pl-8"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>

        <Button onClick={onNewUser} aria-label={t('system.users.newUser')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.users.newUser')}
        </Button>
      </div>

      {/* Table */}
      {!loading && users.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={users}
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
