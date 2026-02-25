import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PostStatusActions } from '../PostStatusActions';

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

// Mock postsApi
const mockPublish = vi.fn();
const mockUnpublish = vi.fn();
const mockRevertToDraft = vi.fn();
const mockRestore = vi.fn();

vi.mock('@/lib/content-api', () => ({
  postsApi: {
    publish: (...args: unknown[]) => mockPublish(...args),
    unpublish: (...args: unknown[]) => mockUnpublish(...args),
    revertToDraft: (...args: unknown[]) => mockRevertToDraft(...args),
    restore: (...args: unknown[]) => mockRestore(...args),
  },
}));

describe('PostStatusActions', () => {
  const defaultProps = {
    postId: 'post-1',
    onStatusChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
    mockPublish.mockResolvedValue({ success: true, data: {} });
    mockUnpublish.mockResolvedValue({ success: true, data: {} });
    mockRevertToDraft.mockResolvedValue({ success: true, data: {} });
    mockRestore.mockResolvedValue({ success: true, data: {} });
  });

  it('shows "Publish" button for draft status', () => {
    render(<PostStatusActions {...defaultProps} status="draft" />);
    expect(screen.getByRole('button', { name: /publish/i })).toBeInTheDocument();
  });

  it('shows "Unpublish" and "Revert to Draft" for published status', () => {
    render(<PostStatusActions {...defaultProps} status="published" />);
    expect(screen.getByRole('button', { name: /unpublish/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /revertToDraft/i })).toBeInTheDocument();
  });

  it('shows "Publish Now" and "Revert to Draft" for scheduled status', () => {
    render(<PostStatusActions {...defaultProps} status="scheduled" />);
    // "Publish Now" uses the publish label
    const buttons = screen.getAllByRole('button');
    const buttonTexts = buttons.map((b) => b.textContent);
    expect(buttonTexts.some((t) => t && /publish/i.test(t))).toBe(true);
    expect(screen.getByRole('button', { name: /revertToDraft/i })).toBeInTheDocument();
  });

  it('shows "Restore" and "Revert to Draft" for archived status', () => {
    render(<PostStatusActions {...defaultProps} status="archived" />);
    expect(screen.getByRole('button', { name: /restore/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /revertToDraft/i })).toBeInTheDocument();
  });

  it('calls publish API and onStatusChange for draft -> publish', async () => {
    const user = userEvent.setup();
    render(<PostStatusActions {...defaultProps} status="draft" />);
    await user.click(screen.getByRole('button', { name: /publish/i }));
    await waitFor(() => {
      expect(mockPublish).toHaveBeenCalledWith('post-1');
    });
    await waitFor(() => {
      expect(defaultProps.onStatusChange).toHaveBeenCalled();
    });
  });

  it('calls unpublish API for published -> unpublish', async () => {
    const user = userEvent.setup();
    render(<PostStatusActions {...defaultProps} status="published" />);
    await user.click(screen.getByRole('button', { name: /unpublish/i }));
    await waitFor(() => {
      expect(mockUnpublish).toHaveBeenCalledWith('post-1');
    });
    await waitFor(() => {
      expect(defaultProps.onStatusChange).toHaveBeenCalled();
    });
  });

  it('calls revertToDraft API for published -> revert', async () => {
    const user = userEvent.setup();
    render(<PostStatusActions {...defaultProps} status="published" />);
    await user.click(screen.getByRole('button', { name: /revertToDraft/i }));
    await waitFor(() => {
      expect(mockRevertToDraft).toHaveBeenCalledWith('post-1');
    });
    await waitFor(() => {
      expect(defaultProps.onStatusChange).toHaveBeenCalled();
    });
  });

  it('calls restore API for archived -> restore', async () => {
    const user = userEvent.setup();
    render(<PostStatusActions {...defaultProps} status="archived" />);
    await user.click(screen.getByRole('button', { name: /restore/i }));
    await waitFor(() => {
      expect(mockRestore).toHaveBeenCalledWith('post-1');
    });
    await waitFor(() => {
      expect(defaultProps.onStatusChange).toHaveBeenCalled();
    });
  });

  it('disables buttons while an action is pending', async () => {
    // Make publish never resolve
    mockPublish.mockReturnValue(new Promise(() => {}));
    const user = userEvent.setup();
    render(<PostStatusActions {...defaultProps} status="draft" />);
    const publishBtn = screen.getByRole('button', { name: /publish/i });
    await user.click(publishBtn);
    await waitFor(() => {
      expect(publishBtn).toBeDisabled();
    });
  });

  it('renders nothing meaningful for unknown status', () => {
    const { container } = render(<PostStatusActions {...defaultProps} status="unknown" />);
    const buttons = container.querySelectorAll('button');
    expect(buttons.length).toBe(0);
  });
});
