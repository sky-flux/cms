import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PostsTable } from '../PostsTable';
import type { PostSummary, PaginationMeta } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const label = key.split('.').pop() || key;
      return params
        ? label.replace(/\{\{(\w+)\}\}/g, (_: string, k: string) => String(params[k]))
        : label;
    },
  }),
}));

const mockPosts: PostSummary[] = [
  {
    id: '1',
    title: 'First Post',
    slug: 'first-post',
    status: 'published',
    author: { id: 'a1', display_name: 'Alice' },
    cover_image: null,
    categories: [{ id: 'c1', name: 'Tech', slug: 'tech' }],
    tags: [{ id: 't1', name: 'React' }],
    view_count: 100,
    published_at: '2026-01-15T10:00:00Z',
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
  {
    id: '2',
    title: 'Draft Article',
    slug: 'draft-article',
    status: 'draft',
    author: { id: 'a2', display_name: 'Bob' },
    cover_image: null,
    categories: [],
    tags: [],
    view_count: 0,
    published_at: null,
    created_at: '2026-01-12T10:00:00Z',
    updated_at: '2026-01-12T10:00:00Z',
  },
  {
    id: '3',
    title: 'Scheduled Post',
    slug: 'scheduled-post',
    status: 'scheduled',
    author: { id: 'a1', display_name: 'Alice' },
    cover_image: null,
    categories: [],
    tags: [],
    view_count: 0,
    published_at: null,
    created_at: '2026-01-20T10:00:00Z',
    updated_at: '2026-01-20T10:00:00Z',
  },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 3,
  total_pages: 1,
};

describe('PostsTable', () => {
  const defaultProps = {
    posts: mockPosts,
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

  it('renders column headers', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getByText('postTitle')).toBeInTheDocument();
    expect(screen.getByText('postStatus')).toBeInTheDocument();
  });

  it('renders post titles', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getByText('First Post')).toBeInTheDocument();
    expect(screen.getByText('Draft Article')).toBeInTheDocument();
    expect(screen.getByText('Scheduled Post')).toBeInTheDocument();
  });

  it('renders author names', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getAllByText('Alice')).toHaveLength(2);
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('renders status badges for each post', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getByText('Published')).toBeInTheDocument();
    expect(screen.getByText('Draft')).toBeInTheDocument();
    expect(screen.getByText('Scheduled')).toBeInTheDocument();
  });

  it('shows loading skeletons when loading', () => {
    const { container } = render(<PostsTable {...defaultProps} loading={true} posts={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('shows empty state with CTA when no posts', () => {
    render(<PostsTable {...defaultProps} posts={[]} />);
    expect(screen.getByText('noPostsFound')).toBeInTheDocument();
    expect(screen.getByText('createFirstPost')).toBeInTheDocument();
  });

  it('calls onNewPost when empty state CTA is clicked', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} posts={[]} />);
    await user.click(screen.getByText('createFirstPost'));
    expect(defaultProps.onNewPost).toHaveBeenCalled();
  });

  it('renders "New Post" button in filter bar', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /newPost/i })).toBeInTheDocument();
  });

  it('calls onNewPost when "New Post" button is clicked', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newPost/i }));
    expect(defaultProps.onNewPost).toHaveBeenCalled();
  });

  it('renders search input', () => {
    render(<PostsTable {...defaultProps} />);
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument();
  });

  it('calls onSearch when search input changes', async () => {
    const user = userEvent.setup();
    render(<PostsTable {...defaultProps} />);
    const searchInput = screen.getByPlaceholderText('searchPlaceholder');
    await user.type(searchInput, 'test');
    expect(defaultProps.onSearch).toHaveBeenCalledWith('test');
  });

  it('renders pagination controls', () => {
    const multiPagePagination: PaginationMeta = {
      page: 1,
      per_page: 20,
      total: 50,
      total_pages: 3,
    };
    render(<PostsTable {...defaultProps} pagination={multiPagePagination} />);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('calls onPageChange when navigating pages', async () => {
    const user = userEvent.setup();
    const multiPagePagination: PaginationMeta = {
      page: 1,
      per_page: 20,
      total: 50,
      total_pages: 3,
    };
    render(<PostsTable {...defaultProps} pagination={multiPagePagination} />);
    await user.click(screen.getByRole('button', { name: /next/i }));
    expect(defaultProps.onPageChange).toHaveBeenCalledWith(2);
  });

  it('formats published_at date', () => {
    render(<PostsTable {...defaultProps} />);
    // The formatted date for 2026-01-15 should appear somehow
    // We check that a non-null published_at is displayed (not showing "--")
    const cells = screen.getAllByRole('cell');
    const dateCell = cells.find(
      (cell) => cell.textContent && cell.textContent.includes('2026'),
    );
    expect(dateCell).toBeTruthy();
  });

  it('shows "--" for posts without published_at', () => {
    render(<PostsTable {...defaultProps} />);
    // Draft Article has null published_at, should show "--"
    expect(screen.getAllByText('--').length).toBeGreaterThan(0);
  });
});
