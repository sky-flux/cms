import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach } from 'vitest';

vi.mock('@/lib/api-client', () => ({
  api: { post: vi.fn(), get: vi.fn() },
  ApiError: class extends Error {
    status: number;
    constructor(s: number, m: string) { super(m); this.status = s; this.name = 'ApiError'; }
  },
}));

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: { changeLanguage: vi.fn(), language: 'en' },
  }),
}));

vi.mock('sonner', () => ({
  toast: { error: vi.fn(), success: vi.fn() },
  Toaster: () => null,
}));

vi.mock('@/components/providers/I18nProvider', () => ({
  I18nProvider: ({ children }: { children: React.ReactNode }) => children,
}));

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { ResetPasswordForm } from '../ResetPasswordForm';

describe('ResetPasswordForm', () => {
  const defaultProps = { token: 'valid-reset-token' };

  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders new password and confirm password fields', () => {
    render(<ResetPasswordForm {...defaultProps} />);
    expect(screen.getByLabelText(/auth\.newPassword/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.confirmPassword/i)).toBeInTheDocument();
  });

  it('renders title', () => {
    render(<ResetPasswordForm {...defaultProps} />);
    expect(screen.getByText(/auth\.resetPasswordTitle/i)).toBeInTheDocument();
  });

  it('shows error when passwords do not match', async () => {
    const user = userEvent.setup();
    render(<ResetPasswordForm {...defaultProps} />);

    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'different123');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('shows error for short password', async () => {
    const user = userEvent.setup();
    render(<ResetPasswordForm {...defaultProps} />);

    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'short');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'short');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls reset password API and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

    render(<ResetPasswordForm {...defaultProps} />);
    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'newpass1234');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'newpass1234');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/v1/auth/reset-password', {
        token: 'valid-reset-token',
        new_password: 'newpass1234',
      });
    });

    await waitFor(() => {
      expect(toast.success).toHaveBeenCalledWith('auth.resetPasswordSuccess');
    });
  });

  it('shows toast on API error', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(400, 'Token expired'));

    render(<ResetPasswordForm {...defaultProps} />);
    await user.type(screen.getByLabelText(/auth\.newPassword/i), 'newpass1234');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'newpass1234');
    await user.click(screen.getByRole('button', { name: /auth\.resetPasswordSubmit/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Token expired');
    });
  });
});
