import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CommentsTable } from '../CommentsTable';
import { CommentDetailDialog } from '../CommentDetailDialog';
import { CommentsPage } from '../CommentsPage';
import type { Comment, PaginationMeta } from '@/lib/system-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const label = key.split('.').pop() || key;
      return params
        ? label.replace(/\{\{(\w+)\}\}/g, (_: string, k: string) => String(params[k]))
        : label;
    },
  }),
  initReactI18next: { type: '3rdParty', init: () => {} },
  I18nextProvider: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock('@/i18n/config', () => ({
  default: { use: () => ({ init: () => {} }) },
}));

vi.mock('@/lib/system-api', async () => {
  const actual = await vi.importActual<typeof import('@/lib/system-api')>('@/lib/system-api');
  return {
    ...actual,
    commentsApi: {
      list: vi.fn().mockResolvedValue({
        success: true,
        data: [],
        pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 },
      }),
      get: vi.fn().mockResolvedValue({ success: true, data: null }),
      updateStatus: vi.fn().mockResolvedValue({ success: true, data: { id: '1', status: 'approved' } }),
      togglePin: vi.fn().mockResolvedValue({ success: true, data: { id: '1', is_pinned: true } }),
      reply: vi.fn().mockResolvedValue({ success: true, data: {} }),
      batchStatus: vi.fn().mockResolvedValue({ success: true, data: { updated_count: 2 } }),
      delete: vi.fn().mockResolvedValue({ success: true }),
    },
  };
});

vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

const mockComments: Comment[] = [
  {
    id: 'c1',
    post: { id: 'p1', title: 'First Post', slug: 'first-post' },
    parent_id: null,
    user_id: null,
    author_name: 'John Doe',
    author_email: 'john@example.com',
    author_url: null,
    author_ip: '192.168.1.1',
    gravatar_url: 'https://gravatar.com/avatar/test',
    content: 'This is a great article with lots of useful information that exceeds one hundred characters in total length for truncation testing purposes.',
    status: 'pending',
    is_pinned: false,
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
  {
    id: 'c2',
    post: { id: 'p2', title: 'Second Post', slug: 'second-post' },
    parent_id: null,
    user_id: 'u1',
    author_name: 'Jane Smith',
    author_email: 'jane@example.com',
    author_url: 'https://jane.com',
    author_ip: '10.0.0.1',
    gravatar_url: 'https://gravatar.com/avatar/test2',
    content: 'Nice work!',
    status: 'approved',
    is_pinned: true,
    created_at: '2026-01-16T10:00:00Z',
    updated_at: '2026-01-16T10:00:00Z',
  },
  {
    id: 'c3',
    post: { id: 'p1', title: 'First Post', slug: 'first-post' },
    parent_id: null,
    user_id: null,
    author_name: 'Spammer Bot',
    author_email: 'spam@example.com',
    author_url: null,
    author_ip: '1.2.3.4',
    gravatar_url: 'https://gravatar.com/avatar/spam',
    content: 'Buy cheap stuff now!',
    status: 'spam',
    is_pinned: false,
    created_at: '2026-01-17T10:00:00Z',
    updated_at: '2026-01-17T10:00:00Z',
  },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 3,
  total_pages: 1,
};

// ============================================================
// CommentsTable Tests
// ============================================================

describe('CommentsTable', () => {
  const defaultProps = {
    comments: mockComments,
    pagination: mockPagination,
    loading: false,
    selectedIds: [] as string[],
    onPageChange: vi.fn(),
    onStatusFilter: vi.fn(),
    onSearch: vi.fn(),
    onSelectChange: vi.fn(),
    onApprove: vi.fn(),
    onReject: vi.fn(),
    onMarkSpam: vi.fn(),
    onTogglePin: vi.fn(),
    onReply: vi.fn(),
    onDelete: vi.fn(),
    onBatchApprove: vi.fn(),
    onBatchReject: vi.fn(),
    onBatchSpam: vi.fn(),
    onViewDetail: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders author names', () => {
    render(<CommentsTable {...defaultProps} />);
    expect(screen.getByText('John Doe')).toBeInTheDocument();
    expect(screen.getByText('Jane Smith')).toBeInTheDocument();
    expect(screen.getByText('Spammer Bot')).toBeInTheDocument();
  });

  it('renders content excerpts truncated to 100 chars', () => {
    render(<CommentsTable {...defaultProps} />);
    // The first comment content exceeds 100 chars, should be truncated with ellipsis
    const longContent = mockComments[0].content;
    // Full content should NOT be in the table
    expect(screen.queryByText(longContent)).not.toBeInTheDocument();
    // Truncated version (100 chars + ...) should be present
    const truncated = longContent.substring(0, 100) + '...';
    expect(screen.getByText(truncated)).toBeInTheDocument();
  });

  it('renders post titles', () => {
    render(<CommentsTable {...defaultProps} />);
    expect(screen.getAllByText('First Post')).toHaveLength(2);
    expect(screen.getByText('Second Post')).toBeInTheDocument();
  });

  it('renders status badges', () => {
    render(<CommentsTable {...defaultProps} />);
    expect(screen.getByText('Pending')).toBeInTheDocument();
    expect(screen.getByText('Approved')).toBeInTheDocument();
    expect(screen.getByText('Spam')).toBeInTheDocument();
  });

  it('renders pin icon for pinned comments', () => {
    render(<CommentsTable {...defaultProps} />);
    // The pinned comment (c2) should have a pin indicator
    const rows = screen.getAllByRole('row');
    // row[0] is header, row[2] is the approved/pinned comment
    const pinnedRow = rows[2];
    const pinIcon = within(pinnedRow).getByTestId('pin-icon');
    expect(pinIcon).toBeInTheDocument();
  });

  it('renders filter bar with status select and search input', () => {
    render(<CommentsTable {...defaultProps} />);
    // Status filter trigger
    expect(screen.getByText('filterByStatus')).toBeInTheDocument();
    // Search input
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument();
  });

  it('calls onSearch when search input changes', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    const searchInput = screen.getByPlaceholderText('searchPlaceholder');
    await user.type(searchInput, 'test');
    expect(defaultProps.onSearch).toHaveBeenCalledWith('test');
  });

  it('renders row actions dropdown with approve for pending comments', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    // Click actions button on the first row (pending comment)
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    // Approve should be available for pending
    expect(screen.getByText('approve')).toBeInTheDocument();
    // Reject should be available
    expect(screen.getByText('reject')).toBeInTheDocument();
    // Mark as Spam
    expect(screen.getByText('markSpam')).toBeInTheDocument();
  });

  it('renders row actions with unpin for pinned comments', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    // Click actions button on the second row (approved + pinned comment)
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[1]);
    // Unpin should be available for pinned
    expect(screen.getByText('unpin')).toBeInTheDocument();
  });

  it('calls onApprove when approve action is clicked', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]); // first row (pending)
    await user.click(screen.getByText('approve'));
    expect(defaultProps.onApprove).toHaveBeenCalledWith('c1');
  });

  it('calls onDelete when delete action is clicked', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('delete'));
    expect(defaultProps.onDelete).toHaveBeenCalledWith('c1');
  });

  it('renders checkbox column and toggles selection', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    // Checkboxes should be present
    const checkboxes = screen.getAllByRole('checkbox');
    // header checkbox + 3 rows = 4
    expect(checkboxes.length).toBe(4);
    // Click a row checkbox
    await user.click(checkboxes[1]); // first data row
    expect(defaultProps.onSelectChange).toHaveBeenCalledWith(['c1']);
  });

  it('shows batch action bar when items are selected', () => {
    render(<CommentsTable {...defaultProps} selectedIds={['c1', 'c3']} />);
    // Mock t('system.comments.selected', { count: 2 }) resolves to 'selected' (no {{count}} in key part)
    expect(screen.getByText('selected')).toBeInTheDocument();
    expect(screen.getByText('batchApprove')).toBeInTheDocument();
    expect(screen.getByText('batchReject')).toBeInTheDocument();
    expect(screen.getByText('batchSpam')).toBeInTheDocument();
  });

  it('calls onBatchApprove when batch approve button is clicked', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} selectedIds={['c1', 'c3']} />);
    await user.click(screen.getByText('batchApprove'));
    expect(defaultProps.onBatchApprove).toHaveBeenCalled();
  });

  it('shows loading skeletons when loading', () => {
    const { container } = render(<CommentsTable {...defaultProps} loading={true} comments={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('shows empty state when no comments', () => {
    render(<CommentsTable {...defaultProps} comments={[]} />);
    expect(screen.getByText('noCommentsFound')).toBeInTheDocument();
  });

  it('selects all via header checkbox', async () => {
    const user = userEvent.setup();
    render(<CommentsTable {...defaultProps} />);
    const checkboxes = screen.getAllByRole('checkbox');
    // Click header checkbox (first checkbox)
    await user.click(checkboxes[0]);
    expect(defaultProps.onSelectChange).toHaveBeenCalledWith(['c1', 'c2', 'c3']);
  });
});

// ============================================================
// CommentDetailDialog Tests
// ============================================================

describe('CommentDetailDialog', () => {
  const commentWithReplies: Comment = {
    ...mockComments[0],
    replies: [
      {
        id: 'r1',
        post: { id: 'p1', title: 'First Post', slug: 'first-post' },
        parent_id: 'c1',
        user_id: 'admin1',
        author_name: 'Admin User',
        author_email: 'admin@example.com',
        author_url: null,
        author_ip: '127.0.0.1',
        gravatar_url: 'https://gravatar.com/avatar/admin',
        content: 'Thank you for your comment!',
        status: 'approved',
        is_pinned: false,
        replies: [
          {
            id: 'r2',
            post: { id: 'p1', title: 'First Post', slug: 'first-post' },
            parent_id: 'r1',
            user_id: null,
            author_name: 'John Doe',
            author_email: 'john@example.com',
            author_url: null,
            author_ip: '192.168.1.1',
            gravatar_url: 'https://gravatar.com/avatar/test',
            content: 'You are welcome!',
            status: 'approved',
            is_pinned: false,
            created_at: '2026-01-15T12:00:00Z',
            updated_at: '2026-01-15T12:00:00Z',
          },
        ],
        created_at: '2026-01-15T11:00:00Z',
        updated_at: '2026-01-15T11:00:00Z',
      },
    ],
  };

  const defaultProps = {
    comment: commentWithReplies,
    open: true,
    onOpenChange: vi.fn(),
    onReply: vi.fn(),
    replyLoading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders full comment content', () => {
    render(<CommentDetailDialog {...defaultProps} />);
    expect(screen.getByText(commentWithReplies.content)).toBeInTheDocument();
  });

  it('renders author info (name, email, IP)', () => {
    render(<CommentDetailDialog {...defaultProps} />);
    // John Doe appears in both author info and reply tree
    expect(screen.getAllByText('John Doe').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('john@example.com')).toBeInTheDocument();
    expect(screen.getByText('192.168.1.1')).toBeInTheDocument();
  });

  it('renders reply tree recursively', () => {
    render(<CommentDetailDialog {...defaultProps} />);
    // First level reply
    expect(screen.getByText('Thank you for your comment!')).toBeInTheDocument();
    expect(screen.getByText('Admin User')).toBeInTheDocument();
    // Second level reply
    expect(screen.getByText('You are welcome!')).toBeInTheDocument();
  });

  it('renders admin reply textarea and submit button', () => {
    render(<CommentDetailDialog {...defaultProps} />);
    expect(screen.getByPlaceholderText('replyPlaceholder')).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /reply/i })).toBeInTheDocument();
  });

  it('calls onReply when reply is submitted', async () => {
    const user = userEvent.setup();
    render(<CommentDetailDialog {...defaultProps} />);
    const textarea = screen.getByPlaceholderText('replyPlaceholder');
    await user.type(textarea, 'My admin reply');
    await user.click(screen.getByRole('button', { name: /reply/i }));
    expect(defaultProps.onReply).toHaveBeenCalledWith('c1', 'My admin reply');
  });

  it('disables reply button when textarea is empty', () => {
    render(<CommentDetailDialog {...defaultProps} />);
    const replyButton = screen.getByRole('button', { name: /reply/i });
    expect(replyButton).toBeDisabled();
  });
});

// ============================================================
// CommentsPage Tests
// ============================================================

describe('CommentsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the comments page title', async () => {
    render(<CommentsPage />);
    expect(screen.getByText('title')).toBeInTheDocument();
  });

  it('renders CommentsTable component', async () => {
    render(<CommentsPage />);
    // The filter bar should be present
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument();
  });
});
