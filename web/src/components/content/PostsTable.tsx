import { useMemo } from 'react';
import type { ColumnDef } from '@tanstack/react-table';
import { useTranslation } from 'react-i18next';
import { MoreHorizontal, Plus, Search } from 'lucide-react';

import { DataTable } from '@/components/shared/DataTable';
import { StatusBadge } from '@/components/shared/StatusBadge';
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
import type { PostSummary, PaginationMeta } from '@/lib/content-api';

interface PostsTableProps {
  posts: PostSummary[];
  pagination: PaginationMeta;
  loading: boolean;
  onPageChange: (page: number) => void;
  onStatusFilter: (status: string) => void;
  onSearch: (query: string) => void;
  onDelete: (id: string) => void;
  onNewPost: () => void;
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

export function PostsTable({
  posts,
  pagination,
  loading,
  onPageChange,
  onStatusFilter,
  onSearch,
  onDelete,
  onNewPost,
}: PostsTableProps) {
  const { t } = useTranslation();

  const columns: ColumnDef<PostSummary>[] = useMemo(
    () => [
      {
        accessorKey: 'title',
        header: t('content.postTitle'),
        cell: ({ row }) => (
          <a
            href={`/dashboard/posts/${row.original.id}/edit`}
            className="font-medium text-primary hover:underline"
          >
            {row.original.title}
          </a>
        ),
      },
      {
        accessorKey: 'status',
        header: t('content.postStatus'),
        cell: ({ row }) => <StatusBadge status={row.original.status} />,
      },
      {
        id: 'author',
        header: t('common.name'),
        cell: ({ row }) => row.original.author.display_name,
      },
      {
        id: 'published_at',
        header: t('content.statusPublished'),
        cell: ({ row }) => formatDate(row.original.published_at),
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
              <DropdownMenuItem asChild>
                <a href={`/dashboard/posts/${row.original.id}/edit`}>
                  {t('common.edit')}
                </a>
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
    [t, onDelete],
  );

  const emptyContent = (
    <div className="flex flex-col items-center gap-2 py-8">
      <p className="text-muted-foreground">{t('content.noPostsFound')}</p>
      <Button variant="outline" onClick={onNewPost}>
        {t('content.createFirstPost')}
      </Button>
    </div>
  );

  return (
    <div className="space-y-4">
      {/* Filter bar */}
      <div className="flex items-center gap-3">
        <Select onValueChange={onStatusFilter}>
          <SelectTrigger className="w-[160px]">
            <SelectValue placeholder={t('content.filterByStatus')} />
          </SelectTrigger>
          <SelectContent>
            <SelectItem value="all">{t('common.status')}</SelectItem>
            <SelectItem value="draft">{t('content.statusDraft')}</SelectItem>
            <SelectItem value="published">{t('content.statusPublished')}</SelectItem>
            <SelectItem value="scheduled">{t('content.statusScheduled')}</SelectItem>
            <SelectItem value="archived">{t('content.statusArchived')}</SelectItem>
          </SelectContent>
        </Select>

        <div className="relative flex-1">
          <Search className="absolute left-2.5 top-2.5 h-4 w-4 text-muted-foreground" />
          <Input
            placeholder={t('content.searchPlaceholder')}
            className="pl-8"
            onChange={(e) => onSearch(e.target.value)}
          />
        </div>

        <Button onClick={onNewPost} aria-label={t('content.newPost')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('content.newPost')}
        </Button>
      </div>

      {/* Table */}
      {!loading && posts.length === 0 ? (
        emptyContent
      ) : (
        <DataTable
          columns={columns}
          data={posts}
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
