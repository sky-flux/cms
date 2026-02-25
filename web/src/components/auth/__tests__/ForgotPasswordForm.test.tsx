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
}));

import { api } from '@/lib/api-client';
import { ForgotPasswordForm } from '../ForgotPasswordForm';

describe('ForgotPasswordForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders email field and submit button', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByLabelText(/auth\.email/i)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i })).toBeInTheDocument();
  });

  it('renders title and description', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByText(/auth\.forgotPasswordTitle/i)).toBeInTheDocument();
    expect(screen.getByText(/auth\.forgotPasswordDescription/i)).toBeInTheDocument();
  });

  it('renders back to login link', () => {
    render(<ForgotPasswordForm />);
    expect(screen.getByText(/auth\.twoFactorBackToLogin/i)).toBeInTheDocument();
  });

  it('does not submit with empty email', async () => {
    const user = userEvent.setup();
    render(<ForgotPasswordForm />);

    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls forgot password API and redirects to sent page', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({ success: true, data: {} });

    render(<ForgotPasswordForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/v1/auth/forgot-password', {
        email: 'a@b.com',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/forgot-password/sent');
    });
  });

  it('redirects to sent page even on API error (prevent email enumeration)', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new Error('network'));

    render(<ForgotPasswordForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.click(screen.getByRole('button', { name: /auth\.forgotPasswordSubmit/i }));

    // Should still redirect -- we never reveal if email exists
    await waitFor(() => {
      expect(window.location.href).toBe('/forgot-password/sent');
    });
  });
});
