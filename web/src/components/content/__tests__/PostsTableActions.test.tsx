import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PostsTable } from '../PostsTable';
import type { PostSummary, PaginationMeta } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key.split('.').pop() || key,
  }),
}));

const mockPost: PostSummary = {
  id: 'post-1',
  title: 'Test Post',
  slug: 'test-post',
  status: 'published',
  author: { id: 'a1', display_name: 'Alice' },
  cover_image: null,
  categories: [],
  tags: [],
  view_count: 42,
  published_at: '2026-01-15T10:00:00Z',
  created_at: '2026-01-10T10:00:00Z',
  updated_at: '2026-01-15T10:00:00Z',
};

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 1,
  total_pages: 1,
};

describe('PostsTable actions dropdown', () => {
  const defaultProps = {
    posts: [mockPost],
    pagination: mockPagination,
    loading: false,
    onPageChange: vi.fn(),
    onStatusFilter: vi.fn(),
    onSearch: vi.fn(),
    onDelete: vi.fn(),
    onNewPost: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('shows all four action items when dropdown is opened', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);

    const trigger = screen.getByRole('button', { name: 'actions' });
    await user.click(trigger);

    expect(screen.getByText('edit')).toBeInTheDocument();
    expect(screen.getByText('view')).toBeInTheDocument();
    expect(screen.getByText('schedule')).toBeInTheDocument();
    expect(screen.getByText('delete')).toBeInTheDocument();
  });

  it('renders View link with target="_blank" and correct href', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);

    await user.click(screen.getByRole('button', { name: 'actions' }));

    const viewLink = screen.getByText('view').closest('a');
    expect(viewLink).toHaveAttribute('href', '/posts/test-post');
    expect(viewLink).toHaveAttribute('target', '_blank');
    expect(viewLink).toHaveAttribute('rel', 'noopener noreferrer');
  });

  it('renders Schedule link pointing to edit page with schedule param', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);

    await user.click(screen.getByRole('button', { name: 'actions' }));

    const scheduleLink = screen.getByText('schedule').closest('a');
    expect(scheduleLink).toHaveAttribute(
      'href',
      '/dashboard/posts/post-1/edit?schedule=true',
    );
  });

  it('renders Edit link with correct href', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);

    await user.click(screen.getByRole('button', { name: 'actions' }));

    const editLink = screen.getByText('edit').closest('a');
    expect(editLink).toHaveAttribute('href', '/dashboard/posts/post-1/edit');
  });

  it('calls onDelete when Delete is clicked', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);

    await user.click(screen.getByRole('button', { name: 'actions' }));
    await user.click(screen.getByText('delete'));

    expect(defaultProps.onDelete).toHaveBeenCalledWith('post-1');
  });
});
