import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Pin, Search } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { StatusBadge } from '@/components/shared/StatusBadge';
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
import type { Comment, PaginationMeta } from '@/lib/system-api';

interface CommentsTableProps {
  comments: Comment[];
  pagination: PaginationMeta;
  loading: boolean;
  selectedIds: string[];
  onPageChange: (page: number) => void;
  onStatusFilter: (status: string) => void;
  onSearch: (query: string) => void;
  onSelectChange: (ids: string[]) => void;
  onApprove: (id: string) => void;
  onReject: (id: string) => void;
  onMarkSpam: (id: string) => void;
  onTogglePin: (id: string, pinned: boolean) => void;
  onReply: (id: string) => void;
  onDelete: (id: string) => void;
  onBatchApprove: () => void;
  onBatchReject: () => void;
  onBatchSpam: () => void;
  onViewDetail: (comment: Comment) => void;
}

function truncateContent(content: string, maxLength = 100): string {
  if (content.length <= maxLength) return content;
  return content.substring(0, maxLength) + '...';
}

export function CommentsTable({
  comments,
  pagination,
  loading,
  selectedIds,
  onPageChange,
  onStatusFilter,
  onSearch,
  onSelectChange,
  onApprove,
  onReject,
  onMarkSpam,
  onTogglePin,
  onReply,
  onDelete,
  onBatchApprove,
  onBatchReject,
  onBatchSpam,
  onViewDetail,
}: CommentsTableProps) {
  const { t } = useTranslation();

  const allIds = useMemo(() => comments.map((c) => c.id), [comments]);
  const allSelected = allIds.length > 0 && allIds.every((id) => selectedIds.includes(id));

  function handleToggleAll() {
    if (allSelected) {
      onSelectChange([]);
    } else {
      onSelectChange(allIds);
    }
  }

  function handleToggleOne(id: string) {
    if (selectedIds.includes(id)) {
      onSelectChange(selectedIds.filter((sid) => sid !== id));
    } else {
      onSelectChange([...selectedIds, id]);
    }
  }

  const columns: ColumnDef<Comment>[] = useMemo(
    () => [
      {
        id: 'select',
        header: () => (
          <Checkbox
            checked={allSelected}
            onCheckedChange={handleToggleAll}
            aria-label="Select all"
          />
        ),
        cell: ({ row }) => (
          <Checkbox
            checked={selectedIds.includes(row.original.id)}
            onCheckedChange={() => handleToggleOne(row.original.id)}
            aria-label={`Select ${row.original.author_name}`}
          />
        ),
      },
      {
        accessorKey: 'author_name',
        header: t('system.comments.author'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.author_name}</span>
        ),
      },
      {
        id: 'content',
        header: t('system.comments.content'),
        cell: ({ row }) => (
          <button
            type="button"
            className="max-w-xs truncate text-left text-sm hover:underline"
            onClick={() => onViewDetail(row.original)}
          >
            {truncateContent(row.original.content)}
          </button>
        ),
      },
      {
        id: 'post',
        header: t('system.comments.post'),
        cell: ({ row }) => row.original.post.title,
      },
      {
        accessorKey: 'status',
        header: t('system.comments.status'),
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        id: 'is_pinned',
        header: t('system.comments.pinned'),
        cell: ({ row }) =>
          row.original.is_pinned ? (
            <Pin className="h-4 w-4 text-primary" data-testid="pin-icon" />
          ) : null,
      },
      {
        id: 'actions',
        header: t('common.actions'),
        cell: ({ row }) => {
          const comment = row.original;
          return (
            <DropdownMenu>
              <DropdownMenuTrigger asChild>
                <Button variant="ghost" size="sm" aria-label={t('common.actions')}>
                  <MoreHorizontal className="h-4 w-4" />
                </Button>
              </DropdownMenuTrigger>
              <DropdownMenuContent align="end">
                {comment.status !== 'approved' && (
                  <DropdownMenuItem onClick={() => onApprove(comment.id)}>
                    {t('system.comments.approve')}
                  </DropdownMenuItem>
                )}
                {comment.status !== 'trash' && comment.status !== 'spam' && (
                  <DropdownMenuItem onClick={() => onReject(comment.id)}>
                    {t('system.comments.reject')}
                  </DropdownMenuItem>
                )}
                {comment.status !== 'spam' && (
                  <DropdownMenuItem onClick={() => onMarkSpam(comment.id)}>
                    {t('system.comments.markSpam')}
                  </DropdownMenuItem>
                )}
                {comment.is_pinned ? (
                  <DropdownMenuItem onClick={() => onTogglePin(comment.id, false)}>
                    {t('system.comments.unpin')}
                  </DropdownMenuItem>
                ) : (
                  <DropdownMenuItem onClick={() => onTogglePin(comment.id, true)}>
                    {t('system.comments.pin')}
                  </DropdownMenuItem>
                )}
                <DropdownMenuItem onClick={() => onReply(comment.id)}>
                  {t('system.comments.reply')}
                </DropdownMenuItem>
                <DropdownMenuItem
                  variant="destructive"
                  onClick={() => onDelete(comment.id)}
                >
                  {t('common.delete')}
                </DropdownMenuItem>
              </DropdownMenuContent>
            </DropdownMenu>
          );
        },
      },
    ],
    [t, selectedIds, allSelected, onApprove, onReject, onMarkSpam, onTogglePin, onReply, onDelete, onViewDetail],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('system.comments.noCommentsFound')}</p>
    </div>
  );

  return (
    <div className="space-y-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <Select onValueChange={onStatusFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder={t('system.comments.filterByStatus')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.status')}</SelectItem>
            <SelectItem value="pending">{t('system.comments.pending')}</SelectItem>
            <SelectItem value="approved">{t('system.comments.approved')}</SelectItem>
            <SelectItem value="spam">{t('system.comments.spam')}</SelectItem>
            <SelectItem value="trash">{t('system.comments.trash')}</SelectItem>
          </SelectContent>
        </Select>

        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('system.comments.searchPlaceholder')}
            className="pl-8"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>
      </div>

      {/* Table */}
      {!loading && comments.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={comments}
          loading={loading}
          pagination={{
            page: pagination.page,
            totalPages: pagination.total_pages,
          }}
          onPageChange={onPageChange}
        />
      )}

      {/* Batch action bar */}
      {selectedIds.length > 0 && (
        <div className="fixed bottom-4 left-1/2 -translate-x-1/2 z-50 flex items-center gap-3 rounded-lg border bg-background px-4 py-3 shadow-lg">
          <span className="text-sm font-medium">
            {t('system.comments.selected', { count: selectedIds.length })}
          </span>
          <Button size="sm" variant="outline" onClick={onBatchApprove}>
            {t('system.comments.batchApprove')}
          </Button>
          <Button size="sm" variant="outline" onClick={onBatchReject}>
            {t('system.comments.batchReject')}
          </Button>
          <Button size="sm" variant="outline" onClick={onBatchSpam}>
            {t('system.comments.batchSpam')}
          </Button>
        </div>
      )}
    </div>
  );
}
