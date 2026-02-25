import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';

// Mock react-i18next
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.revisions': 'Revisions',
        'content.rollback': 'Rollback to this version',
        'content.rollbackConfirm': `Are you sure you want to rollback to version ${opts?.version ?? ''}?`,
        'content.revisionVersion': `Version ${opts?.version ?? ''}`,
        'common.cancel': 'Cancel',
        'common.confirm': 'Confirm',
        'common.loading': 'Loading...',
        'messages.updateSuccess': 'Updated successfully.',
      };
      return map[key] ?? key;
    },
  }),
}));

// Mock data
const mockRevisions = [
  {
    id: 'rev-3',
    version: 3,
    editor: { id: 'user-1', display_name: 'Alice' },
    diff_summary: 'Updated title and content',
    created_at: '2026-02-26T10:00:00Z',
  },
  {
    id: 'rev-2',
    version: 2,
    editor: { id: 'user-2', display_name: 'Bob' },
    diff_summary: 'Added cover image',
    created_at: '2026-02-25T15:00:00Z',
  },
  {
    id: 'rev-1',
    version: 1,
    editor: { id: 'user-1', display_name: 'Alice' },
    diff_summary: 'Initial creation',
    created_at: '2026-02-24T09:00:00Z',
  },
];

let mockRevisionsData: any = { data: mockRevisions };
let mockIsLoading = false;
let mockError: Error | null = null;
const mockRollbackMutate = vi.fn();
let mockRollbackPending = false;

// Mock @tanstack/react-query
vi.mock('@tanstack/react-query', () => ({
  useQuery: () => ({
    data: mockRevisionsData,
    isLoading: mockIsLoading,
    error: mockError,
  }),
  useMutation: ({ onSuccess }: any) => ({
    mutate: (args: any) => {
      mockRollbackMutate(args);
      if ((mockRollbackMutate as any).__autoSuccess && onSuccess) {
        onSuccess();
      }
    },
    isPending: mockRollbackPending,
  }),
  useQueryClient: () => ({
    invalidateQueries: vi.fn(),
  }),
}));

// Mock postsApi
vi.mock('@/lib/content-api', () => ({
  postsApi: {
    getRevisions: vi.fn(),
    rollback: vi.fn().mockResolvedValue({ data: {} }),
  },
}));

// Mock sonner
vi.mock('sonner', () => ({
  toast: {
    success: vi.fn(),
    error: vi.fn(),
  },
}));

import { RevisionHistory } from '../RevisionHistory';

describe('RevisionHistory', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockRevisionsData = { data: mockRevisions };
    mockIsLoading = false;
    mockError = null;
    mockRollbackPending = false;
  });

  it('renders revision list with version numbers', () => {
    render(<RevisionHistory postId="post-1" />);
    expect(screen.getByText('Version 3')).toBeInTheDocument();
    expect(screen.getByText('Version 2')).toBeInTheDocument();
    expect(screen.getByText('Version 1')).toBeInTheDocument();
  });

  it('renders editor names for each revision', () => {
    render(<RevisionHistory postId="post-1" />);
    expect(screen.getAllByText('Alice')).toHaveLength(2);
    expect(screen.getByText('Bob')).toBeInTheDocument();
  });

  it('renders diff_summary for each revision', () => {
    render(<RevisionHistory postId="post-1" />);
    expect(screen.getByText('Updated title and content')).toBeInTheDocument();
    expect(screen.getByText('Added cover image')).toBeInTheDocument();
    expect(screen.getByText('Initial creation')).toBeInTheDocument();
  });

  it('highlights the current (highest) version', () => {
    render(<RevisionHistory postId="post-1" />);
    const currentVersion = screen.getByTestId('revision-rev-3');
    expect(currentVersion.className).toContain('current');
  });

  it('shows rollback buttons on non-current revisions', () => {
    render(<RevisionHistory postId="post-1" />);
    const rollbackButtons = screen.getAllByRole('button', { name: /rollback/i });
    // Only non-current revisions (rev-2, rev-1) should have rollback buttons
    expect(rollbackButtons).toHaveLength(2);
  });

  it('does not show rollback button on current version', () => {
    render(<RevisionHistory postId="post-1" />);
    const currentRevision = screen.getByTestId('revision-rev-3');
    const rollbackBtn = currentRevision.querySelector('button');
    // Current version should not have a rollback button
    expect(rollbackBtn).toBeNull();
  });

  it('shows loading state', () => {
    mockIsLoading = true;
    mockRevisionsData = undefined;
    render(<RevisionHistory postId="post-1" />);
    expect(screen.getByText('Loading...')).toBeInTheDocument();
  });

  it('shows empty state when no revisions', () => {
    mockRevisionsData = { data: [] };
    render(<RevisionHistory postId="post-1" />);
    expect(screen.getByText(/no revisions/i)).toBeInTheDocument();
  });

  it('opens confirm dialog when rollback button clicked', async () => {
    const user = userEvent.setup();
    render(<RevisionHistory postId="post-1" />);
    const rollbackButtons = screen.getAllByRole('button', { name: /rollback/i });
    await user.click(rollbackButtons[0]);
    // Confirm dialog should appear
    await waitFor(() => {
      expect(screen.getByText(/are you sure you want to rollback/i)).toBeInTheDocument();
    });
  });

  it('calls rollback mutation on confirm', async () => {
    const user = userEvent.setup();
    render(<RevisionHistory postId="post-1" />);
    const rollbackButtons = screen.getAllByRole('button', { name: /rollback/i });
    await user.click(rollbackButtons[0]);
    // Click confirm in dialog
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /confirm/i })).toBeInTheDocument();
    });
    await user.click(screen.getByRole('button', { name: /confirm/i }));
    expect(mockRollbackMutate).toHaveBeenCalled();
  });
});
