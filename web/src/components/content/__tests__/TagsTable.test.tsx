import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TagsTable } from '../TagsTable';
import type { Tag } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.tagName': 'Name',
        'content.tagSlug': 'Slug',
        'content.postCount': `${opts?.count ?? 0} posts`,
        'content.searchPlaceholder': 'Search...',
        'content.noTagsFound': 'No tags yet',
        'content.editTag': 'Edit Tag',
        'content.deleteTag': 'Delete Tag',
        'common.actions': 'Actions',
      };
      if (key === 'content.postCount' && opts?.count !== undefined) {
        return `${opts.count} posts`;
      }
      return map[key] ?? key;
    },
  }),
}));

const mockTags: Tag[] = [
  { id: 'tag-1', name: 'JavaScript', slug: 'javascript', post_count: 15, created_at: '2026-01-15T00:00:00Z' },
  { id: 'tag-2', name: 'TypeScript', slug: 'typescript', post_count: 8, created_at: '2026-01-20T00:00:00Z' },
  { id: 'tag-3', name: 'React', slug: 'react', post_count: 12, created_at: '2026-02-01T00:00:00Z' },
];

describe('TagsTable', () => {
  it('renders tag names in the table', () => {
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    expect(screen.getByText('JavaScript')).toBeInTheDocument();
    expect(screen.getByText('TypeScript')).toBeInTheDocument();
    expect(screen.getByText('React')).toBeInTheDocument();
  });

  it('renders slug column', () => {
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    expect(screen.getByText('javascript')).toBeInTheDocument();
    expect(screen.getByText('typescript')).toBeInTheDocument();
  });

  it('renders post count', () => {
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    expect(screen.getByText('15 posts')).toBeInTheDocument();
    expect(screen.getByText('8 posts')).toBeInTheDocument();
  });

  it('renders empty state when no tags', () => {
    render(
      <TagsTable
        tags={[]}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    expect(screen.getByText('No tags yet')).toBeInTheDocument();
  });

  it('renders search input with value', () => {
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue="java"
        onSearchChange={vi.fn()}
      />,
    );
    const searchInput = screen.getByPlaceholderText('Search...');
    expect(searchInput).toHaveValue('java');
  });

  it('calls onSearchChange when typing in search', async () => {
    const onSearchChange = vi.fn();
    const user = userEvent.setup();
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={onSearchChange}
      />,
    );
    await user.type(screen.getByPlaceholderText('Search...'), 'test');
    expect(onSearchChange).toHaveBeenCalled();
  });

  it('calls onEdit when edit button clicked', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();
    render(
      <TagsTable
        tags={mockTags}
        onEdit={onEdit}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    await user.click(editButtons[0]);
    expect(onEdit).toHaveBeenCalledWith(mockTags[0]);
  });

  it('calls onDelete when delete button clicked', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={onDelete}
        searchValue=""
        onSearchChange={vi.fn()}
      />,
    );
    const deleteButtons = screen.getAllByRole('button', { name: /delete/i });
    await user.click(deleteButtons[0]);
    expect(onDelete).toHaveBeenCalledWith(mockTags[0]);
  });

  it('shows loading state', () => {
    const { container } = render(
      <TagsTable
        tags={[]}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        loading={true}
      />,
    );
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });

  it('renders pagination when provided', () => {
    render(
      <TagsTable
        tags={mockTags}
        onEdit={vi.fn()}
        onDelete={vi.fn()}
        searchValue=""
        onSearchChange={vi.fn()}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={vi.fn()}
      />,
    );
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });
});
