import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { ColumnDef } from '@tanstack/react-table';
import { Pencil, Trash2 } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Badge } from '@/components/ui/badge';
import { DataTable } from '@/components/shared/DataTable';
import type { Tag } from '@/lib/content-api';

interface TagsTableProps {
  tags: Tag[];
  onEdit: (tag: Tag) => void;
  onDelete: (tag: Tag) => void;
  searchValue: string;
  onSearchChange: (value: string) => void;
  loading?: boolean;
  pagination?: { page: number; totalPages: number };
  onPageChange?: (page: number) => void;
}

export function TagsTable({
  tags,
  onEdit,
  onDelete,
  searchValue,
  onSearchChange,
  loading = false,
  pagination,
  onPageChange,
}: TagsTableProps) {
  const { t } = useTranslation();

  const columns = useMemo<ColumnDef<Tag, unknown>[]>(
    () => [
      {
        accessorKey: 'name',
        header: t('content.tagName'),
        cell: ({ row }) => (
          <span className="font-medium">{row.original.name}</span>
        ),
      },
      {
        accessorKey: 'slug',
        header: t('content.tagSlug'),
        cell: ({ row }) => (
          <code className="text-sm text-muted-foreground">{row.original.slug}</code>
        ),
      },
      {
        accessorKey: 'post_count',
        header: String(t('content.postCount', { count: 0 })).replace(/0\s*/, ''),
        cell: ({ row }) => (
          <Badge variant="secondary">
            {t('content.postCount', { count: row.original.post_count })}
          </Badge>
        ),
      },
      {
        id: 'actions',
        header: t('common.actions'),
        cell: ({ row }) => (
          <div className="flex items-center gap-1">
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0"
              onClick={() => onEdit(row.original)}
              aria-label="Edit Tag"
            >
              <Pencil className="h-3.5 w-3.5" />
            </Button>
            <Button
              variant="ghost"
              size="sm"
              className="h-7 w-7 p-0 text-destructive hover:text-destructive"
              onClick={() => onDelete(row.original)}
              aria-label="Delete Tag"
            >
              <Trash2 className="h-3.5 w-3.5" />
            </Button>
          </div>
        ),
      },
    ],
    [t, onEdit, onDelete],
  );

  return (
    <div className="space-y-4">
      <div className="flex items-center gap-4">
        <Input
          placeholder={t('content.searchPlaceholder')}
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          className="max-w-sm"
        />
      </div>

      <DataTable
        columns={columns}
        data={tags}
        loading={loading}
        emptyMessage={t('content.noTagsFound')}
        pagination={pagination}
        onPageChange={onPageChange}
      />
    </div>
  );
}
