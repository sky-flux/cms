import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MediaDetailDialog } from '../MediaDetailDialog';
import type { MediaFileDetail } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.mediaDetail': 'File Details',
        'content.mediaAltText': 'Alt Text',
        'content.mediaTitle': 'Title',
        'content.mediaFileName': 'File Name',
        'content.mediaFileSize': 'File Size',
        'content.mediaDimensions': 'Dimensions',
        'content.mediaType': 'Type',
        'content.mediaReferences': 'Referenced by',
        'content.deleteMedia': 'Delete File',
        'common.save': 'Save',
        'common.cancel': 'Cancel',
        'common.loading': 'Loading...',
      };
      if (key === 'content.deleteMediaConfirm' && opts?.name) {
        return `Delete "${opts.name}"?`;
      }
      if (key === 'content.deleteMediaReferenced' && opts?.count !== undefined) {
        return `This file is referenced by ${opts.count} posts. Force delete?`;
      }
      return map[key] ?? key;
    },
  }),
}));

const mockMediaDetail: MediaFileDetail = {
  id: 'media-1',
  file_name: 'hero.jpg',
  original_name: 'hero.jpg',
  mime_type: 'image/jpeg',
  media_type: 'image',
  file_size: 102400,
  public_url: 'https://example.com/hero.jpg',
  thumbnail_urls: { sm: 'https://example.com/hero_sm.jpg', md: 'https://example.com/hero_md.jpg' },
  width: 1920,
  height: 1080,
  alt_text: 'A hero image',
  title: 'Hero Image',
  reference_count: 2,
  referencing_posts: [
    { id: 'post-1', title: 'First Post' },
    { id: 'post-2', title: 'Second Post' },
  ],
  created_at: '2026-01-15T00:00:00Z',
  updated_at: '2026-01-15T00:00:00Z',
};

const mockMediaNoRefs: MediaFileDetail = {
  ...mockMediaDetail,
  id: 'media-2',
  reference_count: 0,
  referencing_posts: [],
};

describe('MediaDetailDialog', () => {
  it('renders file details when open', () => {
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('File Details')).toBeInTheDocument();
    expect(screen.getByText('hero.jpg')).toBeInTheDocument();
  });

  it('does not render when closed', () => {
    render(
      <MediaDetailDialog
        open={false}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.queryByText('File Details')).not.toBeInTheDocument();
  });

  it('displays image preview for image type', () => {
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    const img = screen.getByRole('img');
    expect(img).toHaveAttribute('src', 'https://example.com/hero_md.jpg');
  });

  it('displays file metadata', () => {
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('image/jpeg')).toBeInTheDocument();
    expect(screen.getByText('1920 x 1080')).toBeInTheDocument();
  });

  it('shows editable alt text and title fields', () => {
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByLabelText('Alt Text')).toHaveValue('A hero image');
    expect(screen.getByLabelText('Title')).toHaveValue('Hero Image');
  });

  it('calls onSave with updated metadata', async () => {
    const onSave = vi.fn().mockResolvedValue(undefined);
    const user = userEvent.setup();
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={onSave}
        onDelete={vi.fn()}
      />,
    );

    const altInput = screen.getByLabelText('Alt Text');
    await user.clear(altInput);
    await user.type(altInput, 'Updated alt');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSave).toHaveBeenCalledWith(
        expect.objectContaining({ alt_text: 'Updated alt' }),
      );
    });
  });

  it('shows referencing posts list', () => {
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={vi.fn()}
      />,
    );
    expect(screen.getByText('Referenced by')).toBeInTheDocument();
    expect(screen.getByText('First Post')).toBeInTheDocument();
    expect(screen.getByText('Second Post')).toBeInTheDocument();
  });

  it('calls onDelete when delete button clicked', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaNoRefs}
        onSave={vi.fn()}
        onDelete={onDelete}
      />,
    );

    await user.click(screen.getByRole('button', { name: /delete file/i }));
    expect(onDelete).toHaveBeenCalledWith('media-2', false);
  });

  it('offers force delete when file has references', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <MediaDetailDialog
        open={true}
        onOpenChange={vi.fn()}
        media={mockMediaDetail}
        onSave={vi.fn()}
        onDelete={onDelete}
      />,
    );

    await user.click(screen.getByRole('button', { name: /delete file/i }));
    expect(onDelete).toHaveBeenCalledWith('media-1', true);
  });
});
