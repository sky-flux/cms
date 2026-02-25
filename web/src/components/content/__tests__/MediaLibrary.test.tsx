import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MediaLibrary } from '../MediaLibrary';
import type { MediaFile } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.media': 'Media Library',
        'content.gridView': 'Grid View',
        'content.listView': 'List View',
        'content.searchPlaceholder': 'Search...',
        'content.filterByType': 'Filter by type',
        'content.noMediaFound': 'No media files yet',
        'content.batchDelete': 'Delete Selected',
        'content.mediaFileName': 'File Name',
        'content.mediaType': 'Type',
        'content.mediaFileSize': 'File Size',
        'common.actions': 'Actions',
        'content.deleteMedia': 'Delete',
      };
      if (key === 'content.selected' && opts?.count !== undefined) {
        return `${opts.count} selected`;
      }
      if (key === 'content.batchDeleteConfirm' && opts?.count !== undefined) {
        return `Delete ${opts.count} selected files?`;
      }
      return map[key] ?? key;
    },
  }),
}));

const mockMedia: MediaFile[] = [
  {
    id: 'media-1',
    file_name: 'hero.jpg',
    original_name: 'hero.jpg',
    mime_type: 'image/jpeg',
    media_type: 'image',
    file_size: 102400,
    public_url: 'https://example.com/hero.jpg',
    thumbnail_urls: { sm: 'https://example.com/hero_sm.jpg', md: 'https://example.com/hero_md.jpg' },
    reference_count: 2,
    created_at: '2026-01-15T00:00:00Z',
    updated_at: '2026-01-15T00:00:00Z',
  },
  {
    id: 'media-2',
    file_name: 'document.pdf',
    original_name: 'document.pdf',
    mime_type: 'application/pdf',
    media_type: 'document',
    file_size: 204800,
    public_url: 'https://example.com/document.pdf',
    reference_count: 0,
    created_at: '2026-02-01T00:00:00Z',
    updated_at: '2026-02-01T00:00:00Z',
  },
  {
    id: 'media-3',
    file_name: 'video.mp4',
    original_name: 'video.mp4',
    mime_type: 'video/mp4',
    media_type: 'video',
    file_size: 5242880,
    public_url: 'https://example.com/video.mp4',
    reference_count: 1,
    created_at: '2026-02-10T00:00:00Z',
    updated_at: '2026-02-10T00:00:00Z',
  },
];

describe('MediaLibrary', () => {
  it('renders media items in grid view by default', () => {
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('hero.jpg')).toBeInTheDocument();
    expect(screen.getByText('document.pdf')).toBeInTheDocument();
    expect(screen.getByText('video.mp4')).toBeInTheDocument();
  });

  it('shows empty state when no media', () => {
    render(
      <MediaLibrary
        media={[]}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('No media files yet')).toBeInTheDocument();
  });

  it('toggles between grid and list view', async () => {
    const onViewModeChange = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={onViewModeChange}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );

    await user.click(screen.getByRole('button', { name: /list/i }));
    expect(onViewModeChange).toHaveBeenCalledWith('list');
  });

  it('calls onItemClick when media item is clicked', async () => {
    const onItemClick = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={onItemClick}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );

    await user.click(screen.getByText('hero.jpg'));
    expect(onItemClick).toHaveBeenCalledWith(mockMedia[0]);
  });

  it('shows search input with value', () => {
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue="hero"
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );
    expect(screen.getByPlaceholderText('Search...')).toHaveValue('hero');
  });

  it('calls onSearchChange when typing', async () => {
    const onSearchChange = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={onSearchChange}
        onBatchDelete={vi.fn()}
      />,
    );

    await user.type(screen.getByPlaceholderText('Search...'), 'test');
    expect(onSearchChange).toHaveBeenCalled();
  });

  it('handles item selection via checkbox', async () => {
    const onSelectionChange = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={onSelectionChange}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );

    const checkboxes = screen.getAllByRole('checkbox');
    await user.click(checkboxes[0]);
    expect(onSelectionChange).toHaveBeenCalledWith(['media-1']);
  });

  it('shows batch delete toolbar when items are selected', () => {
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={['media-1', 'media-2']}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('2 selected')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /delete selected/i })).toBeInTheDocument();
  });

  it('calls onBatchDelete when batch delete clicked', async () => {
    const onBatchDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={['media-1', 'media-2']}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={onBatchDelete}
      />,
    );

    await user.click(screen.getByRole('button', { name: /delete selected/i }));
    expect(onBatchDelete).toHaveBeenCalledWith(['media-1', 'media-2']);
  });

  it('renders list view with table when viewMode is list', () => {
    render(
      <MediaLibrary
        media={mockMedia}
        viewMode="list"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
      />,
    );
    // Table should have column headers
    expect(screen.getByText('File Name')).toBeInTheDocument();
    expect(screen.getByText('Type')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    const { container } = render(
      <MediaLibrary
        media={[]}
        viewMode="grid"
        onViewModeChange={vi.fn()}
        onItemClick={vi.fn()}
        selectedIds={[]}
        onSelectionChange={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        onBatchDelete={vi.fn()}
        loading={true}
      />,
    );
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });
});
