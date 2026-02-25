import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';

import { DataTable } from '@/components/shared/DataTable';
import { Badge } from '@/components/ui/badge';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { AuditLog, PaginationMeta } from '@/lib/system-api';

interface AuditTableProps {
  logs: AuditLog[];
  pagination: PaginationMeta;
  loading: boolean;
  onPageChange: (page: number) => void;
  onActionFilter: (action: string) => void;
  onResourceTypeFilter: (resourceType: string) => void;
  onStartDateChange: (date: string) => void;
  onEndDateChange: (date: string) => void;
}

const actionColors: Record<string, string> = {
  create: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  update: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  delete: 'bg-red-100 text-red-700 dark:bg-red-900 dark:text-red-300',
  login: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
};

const resourceColors: Record<string, string> = {
  post: 'bg-sky-100 text-sky-700 dark:bg-sky-900 dark:text-sky-300',
  user: 'bg-amber-100 text-amber-700 dark:bg-amber-900 dark:text-amber-300',
  setting: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  comment: 'bg-teal-100 text-teal-700 dark:bg-teal-900 dark:text-teal-300',
  media: 'bg-pink-100 text-pink-700 dark:bg-pink-900 dark:text-pink-300',
  menu: 'bg-indigo-100 text-indigo-700 dark:bg-indigo-900 dark:text-indigo-300',
  redirect: 'bg-orange-100 text-orange-700 dark:bg-orange-900 dark:text-orange-300',
};

const actionLabels: Record<string, string> = {
  create: 'Create',
  update: 'Update',
  delete: 'Delete',
  login: 'Login',
};

const resourceLabels: Record<string, string> = {
  post: 'Post',
  user: 'User',
  setting: 'Setting',
  comment: 'Comment',
  media: 'Media',
  menu: 'Menu',
  redirect: 'Redirect',
};

function formatDate(dateStr: string | null): string {
  if (!dateStr) return '--';
  return new Date(dateStr).toLocaleDateString('en-US', {
    year: 'numeric',
    month: 'short',
    day: 'numeric',
  });
}

export function AuditTable({
  logs,
  pagination,
  loading,
  onPageChange,
  onActionFilter,
  onResourceTypeFilter,
  onStartDateChange,
  onEndDateChange,
}: AuditTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<AuditLog>[] = useMemo(
    () => [
      {
        id: 'actor',
        header: t('system.audit.actor'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.actor.display_name}</span>
        ),
      },
      {
        accessorKey: 'action',
        header: t('system.audit.action'),
        cell: ({ row }) => {
          const action = row.original.action;
          return (
            <Badge variant="outline" className={actionColors[action] ?? ''}>
              {actionLabels[action] ?? action}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'resource_type',
        header: t('system.audit.resourceType'),
        cell: ({ row }) => {
          const rt = row.original.resource_type;
          return (
            <Badge variant="outline" className={resourceColors[rt] ?? ''}>
              {resourceLabels[rt] ?? rt}
            </Badge>
          );
        },
      },
      {
        accessorKey: 'resource_id',
        header: t('system.audit.resourceId'),
        cell: ({ row }) => (
          <span className="font-mono text-xs">{row.original.resource_id}</span>
        ),
      },
      {
        accessorKey: 'ip_address',
        header: t('system.audit.ipAddress'),
        cell: ({ row }) => (
          <span className="text-sm">{row.original.ip_address}</span>
        ),
      },
      {
        id: 'created_at',
        header: t('system.audit.timestamp'),
        cell: ({ row }) => formatDate(row.original.created_at),
      },
    ],
    [t],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.audit.noLogsFound')}</p>
    </div>
  );

  return (
    <div className="space-y-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3 flex-wrap">
        <Select onValueChange={onActionFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder={t('system.audit.filterByAction')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('system.audit.action')}</SelectItem>
            <SelectItem value="create">{t('system.audit.actions.create')}</SelectItem>
            <SelectItem value="update">{t('system.audit.actions.update')}</SelectItem>
            <SelectItem value="delete">{t('system.audit.actions.delete')}</SelectItem>
            <SelectItem value="login">{t('system.audit.actions.login')}</SelectItem>
          </SelectContent>
        </Select>

        <Select onValueChange={onResourceTypeFilter}>
          <SelectTrigger className="w-[180px]">
            <SelectValue placeholder={t('system.audit.filterByResource')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('system.audit.resourceType')}</SelectItem>
            <SelectItem value="post">{t('system.audit.resources.post')}</SelectItem>
            <SelectItem value="user">{t('system.audit.resources.user')}</SelectItem>
            <SelectItem value="setting">{t('system.audit.resources.setting')}</SelectItem>
            <SelectItem value="comment">{t('system.audit.resources.comment')}</SelectItem>
            <SelectItem value="media">{t('system.audit.resources.media')}</SelectItem>
            <SelectItem value="menu">{t('system.audit.resources.menu')}</SelectItem>
            <SelectItem value="redirect">{t('system.audit.resources.redirect')}</SelectItem>
          </SelectContent>
        </Select>

        <div className="flex items-center gap-2">
          <label htmlFor="audit-start-date" className="sr-only">
            {t('system.audit.startDate')}
          </label>
          <input
            id="audit-start-date"
            type="date"
            aria-label={t('system.audit.startDate')}
            className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs"
            onChange={(e) => onStartDateChange(e.target.value)}
          />
        </div>

        <div className="flex items-center gap-2">
          <label htmlFor="audit-end-date" className="sr-only">
            {t('system.audit.endDate')}
          </label>
          <input
            id="audit-end-date"
            type="date"
            aria-label={t('system.audit.endDate')}
            className="h-9 rounded-md border border-input bg-background px-3 py-1 text-sm shadow-xs"
            onChange={(e) => onEndDateChange(e.target.value)}
          />
        </div>
      </div>

      {/* Table */}
      {!loading && logs.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={logs}
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
