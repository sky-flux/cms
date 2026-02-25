import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';
import type { ColumnDef } from '@tanstack/react-table';
import {
  LayoutGrid,
  List,
  Trash2,
  ImageIcon,
  FileText,
  Video,
} from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Checkbox } from '@/components/ui/checkbox';
import { Skeleton } from '@/components/ui/skeleton';
import { DataTable } from '@/components/shared/DataTable';
import type { MediaFile } from '@/lib/content-api';

type ViewMode = 'grid' | 'list';

interface MediaLibraryProps {
  media: MediaFile[];
  viewMode: ViewMode;
  onViewModeChange: (mode: ViewMode) => void;
  onItemClick: (item: MediaFile) => void;
  selectedIds: string[];
  onSelectionChange: (ids: string[]) => void;
  searchValue: string;
  onSearchChange: (value: string) => void;
  onBatchDelete: (ids: string[]) => void;
  loading?: boolean;
  pagination?: { page: number; totalPages: number };
  onPageChange?: (page: number) => void;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

function MediaIcon({ mediaType }: { mediaType: string }) {
  switch (mediaType) {
    case 'image':
      return <ImageIcon className="h-8 w-8 text-muted-foreground" />;
    case 'video':
      return <Video className="h-8 w-8 text-muted-foreground" />;
    default:
      return <FileText className="h-8 w-8 text-muted-foreground" />;
  }
}

export function MediaLibrary({
  media,
  viewMode,
  onViewModeChange,
  onItemClick,
  selectedIds,
  onSelectionChange,
  searchValue,
  onSearchChange,
  onBatchDelete,
  loading = false,
  pagination,
  onPageChange,
}: MediaLibraryProps) {
  const { t } = useTranslation();

  const toggleSelection = (id: string) => {
    if (selectedIds.includes(id)) {
      onSelectionChange(selectedIds.filter((i) => i !== id));
    } else {
      onSelectionChange([...selectedIds, id]);
    }
  };

  const columns = useMemo<ColumnDef<MediaFile, unknown>[]>(
    () => [
      {
        id: 'select',
        header: '',
        cell: ({ row }) => (
          <Checkbox
            checked={selectedIds.includes(row.original.id)}
            onCheckedChange={() => toggleSelection(row.original.id)}
            aria-label={`Select ${row.original.file_name}`}
          />
        ),
      },
      {
        id: 'preview',
        header: '',
        cell: ({ row }) => {
          const item = row.original;
          if (item.media_type === 'image' && item.thumbnail_urls?.sm) {
            return (
              <img
                src={item.thumbnail_urls.sm}
                alt={item.file_name}
                className="h-10 w-10 rounded object-cover"
              />
            );
          }
          return <MediaIcon mediaType={item.media_type} />;
        },
      },
      {
        accessorKey: 'file_name',
        header: t('content.mediaFileName'),
        cell: ({ row }) => (
          <button
            className="text-left font-medium hover:underline"
            onClick={() => onItemClick(row.original)}
          >
            {row.original.file_name}
          </button>
        ),
      },
      {
        accessorKey: 'media_type',
        header: t('content.mediaType'),
        cell: ({ row }) => (
          <span className="capitalize">{row.original.media_type}</span>
        ),
      },
      {
        accessorKey: 'file_size',
        header: t('content.mediaFileSize'),
        cell: ({ row }) => formatFileSize(row.original.file_size),
      },
    ],
    [t, selectedIds, onItemClick],
  );

  return (
    <div className="space-y-4">
      {/* Toolbar */}
      <div className="flex items-center gap-4">
        <Input
          placeholder={t('content.searchPlaceholder')}
          value={searchValue}
          onChange={(e) => onSearchChange(e.target.value)}
          className="max-w-sm"
        />
        <div className="flex items-center gap-1 ml-auto">
          <Button
            variant={viewMode === 'grid' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => onViewModeChange('grid')}
            aria-label="Grid View"
          >
            <LayoutGrid className="h-4 w-4" />
          </Button>
          <Button
            variant={viewMode === 'list' ? 'secondary' : 'ghost'}
            size="sm"
            onClick={() => onViewModeChange('list')}
            aria-label="List View"
          >
            <List className="h-4 w-4" />
          </Button>
        </div>
      </div>

      {/* Batch actions toolbar */}
      {selectedIds.length > 0 && (
        <div className="flex items-center gap-4 rounded-md border bg-muted/50 px-4 py-2">
          <span className="text-sm font-medium">
            {t('content.selected', { count: selectedIds.length })}
          </span>
          <Button
            variant="destructive"
            size="sm"
            onClick={() => onBatchDelete(selectedIds)}
            aria-label="Delete Selected"
          >
            <Trash2 className="h-4 w-4 mr-1" />
            {t('content.batchDelete')}
          </Button>
        </div>
      )}

      {/* Content */}
      {loading ? (
        <div className="grid grid-cols-4 gap-4">
          {Array.from({ length: 8 }).map((_, i) => (
            <Skeleton key={i} className="h-40 rounded-md" />
          ))}
        </div>
      ) : media.length === 0 ? (
        <div className="flex flex-col items-center justify-center py-12 text-muted-foreground">
          <ImageIcon className="h-12 w-12 mb-4 opacity-50" />
          <p>{t('content.noMediaFound')}</p>
        </div>
      ) : viewMode === 'grid' ? (
        <div className="grid grid-cols-2 gap-4 sm:grid-cols-3 md:grid-cols-4">
          {media.map((item) => (
            <div
              key={item.id}
              className={`group relative rounded-md border overflow-hidden cursor-pointer transition-colors hover:border-primary ${
                selectedIds.includes(item.id) ? 'ring-2 ring-primary' : ''
              }`}
            >
              <div className="absolute top-2 left-2 z-10">
                <Checkbox
                  checked={selectedIds.includes(item.id)}
                  onCheckedChange={() => toggleSelection(item.id)}
                  aria-label={`Select ${item.file_name}`}
                />
              </div>
              <button
                className="w-full text-left"
                onClick={() => onItemClick(item)}
              >
                <div className="aspect-square flex items-center justify-center bg-muted">
                  {item.media_type === 'image' && item.thumbnail_urls?.md ? (
                    <img
                      src={item.thumbnail_urls.md}
                      alt={item.file_name}
                      className="h-full w-full object-cover"
                    />
                  ) : (
                    <MediaIcon mediaType={item.media_type} />
                  )}
                </div>
                <div className="p-2">
                  <p className="truncate text-sm font-medium">{item.file_name}</p>
                  <p className="text-xs text-muted-foreground">
                    {formatFileSize(item.file_size)}
                  </p>
                </div>
              </button>
            </div>
          ))}
        </div>
      ) : (
        <DataTable
          columns={columns}
          data={media}
          emptyMessage={t('content.noMediaFound')}
          pagination={pagination}
          onPageChange={onPageChange}
        />
      )}

      {/* Grid view pagination */}
      {viewMode === 'grid' && pagination && onPageChange && (
        <div className="flex items-center justify-end gap-2 py-4">
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(pagination.page - 1)}
            disabled={pagination.page <= 1}
          >
            Previous
          </Button>
          <span className="text-sm text-muted-foreground">
            {pagination.page} / {pagination.totalPages}
          </span>
          <Button
            variant="outline"
            size="sm"
            onClick={() => onPageChange(pagination.page + 1)}
            disabled={pagination.page >= pagination.totalPages}
          >
            Next
          </Button>
        </div>
      )}
    </div>
  );
}
