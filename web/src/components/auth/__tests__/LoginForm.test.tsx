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

import { api, ApiError } from '@/lib/api-client';
import { toast } from 'sonner';
import { LoginForm } from '../LoginForm';

describe('LoginForm', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    // Reset window.location
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders email and password fields', () => {
    render(<LoginForm />);
    expect(screen.getByLabelText(/auth\.email/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.password/i)).toBeInTheDocument();
  });

  it('renders sign in button', () => {
    render(<LoginForm />);
    expect(screen.getByRole('button', { name: /auth\.loginButton/i })).toBeInTheDocument();
  });

  it('renders forgot password link', () => {
    render(<LoginForm />);
    expect(screen.getByText(/auth\.forgotPasswordLink/i)).toBeInTheDocument();
  });

  it('shows validation error for empty email', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('shows validation error for short password', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    await user.type(screen.getByLabelText(/auth\.email/i), 'test@example.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'short');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).not.toHaveBeenCalled();
    });
  });

  it('calls login API and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        user: { id: '1', email: 'a@b.com', display_name: 'Admin' },
        access_token: 'jwt-tok',
        token_type: 'Bearer',
        expires_in: 900,
      },
    });

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/v1/auth/login', {
        email: 'a@b.com',
        password: 'password123',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/dashboard');
    });
  });

  it('redirects to 2FA page when requires_2fa', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        temp_token: 'tmp-tok',
        token_type: 'Bearer',
        expires_in: 300,
        requires: 'totp',
      },
    });

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(window.location.href).toContain('/login/2fa');
      expect(window.location.href).toContain('temp=tmp-tok');
    });
  });

  it('shows toast on API error', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(401, 'Invalid credentials'));

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Invalid credentials');
    });
  });

  it('disables button while submitting', async () => {
    const user = userEvent.setup();
    let resolveLogin: (v: unknown) => void;
    vi.mocked(api.post).mockImplementation(
      () => new Promise((resolve) => { resolveLogin = resolve; }),
    );

    render(<LoginForm />);
    await user.type(screen.getByLabelText(/auth\.email/i), 'a@b.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.loginButton/i }));

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /auth\.loginButton/i })).toBeDisabled();
    });

    // Resolve and wait for state update to complete
    resolveLogin!({ success: true, data: { user: { id: '1', email: 'a@b.com', display_name: 'A' }, access_token: 'tok' } });
    await waitFor(() => {
      expect(window.location.href).toBe('/dashboard');
    });
  });

  it('toggles password visibility', async () => {
    const user = userEvent.setup();
    render(<LoginForm />);

    const passwordInput = screen.getByLabelText(/auth\.password/i);
    expect(passwordInput).toHaveAttribute('type', 'password');

    const toggleBtn = screen.getByRole('button', { name: /auth\.showPassword/i });
    await user.click(toggleBtn);
    expect(passwordInput).toHaveAttribute('type', 'text');
  });
});
