import { render, screen, waitFor, cleanup } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest';

// input-otp uses ResizeObserver and document.elementFromPoint which are not available in jsdom
class ResizeObserverMock {
  observe() {}
  unobserve() {}
  disconnect() {}
}
global.ResizeObserver = ResizeObserverMock as unknown as typeof ResizeObserver;

if (!document.elementFromPoint) {
  document.elementFromPoint = () => null;
}

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
import { TwoFactorForm } from '../TwoFactorForm';

describe('TwoFactorForm', () => {
  const defaultProps = { tempToken: 'test-temp-token' };

  beforeEach(() => {
    vi.useFakeTimers({ shouldAdvanceTime: true });
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  afterEach(() => {
    cleanup();
    vi.runOnlyPendingTimers();
    vi.useRealTimers();
  });

  it('renders title and description', () => {
    render(<TwoFactorForm {...defaultProps} />);
    expect(screen.getByText(/auth\.twoFactorTitle/i)).toBeInTheDocument();
    expect(screen.getByText(/auth\.twoFactorDescription/i)).toBeInTheDocument();
  });

  it('renders OTP input', () => {
    render(<TwoFactorForm {...defaultProps} />);
    // InputOTP renders a single hidden input element
    const input = document.querySelector('input[data-input-otp="true"]') ?? screen.getByRole('textbox');
    expect(input).toBeInTheDocument();
  });

  it('renders back to login link', () => {
    render(<TwoFactorForm {...defaultProps} />);
    expect(screen.getByText(/auth\.twoFactorBackToLogin/i)).toBeInTheDocument();
  });

  it('calls validate2FA API when 6 digits entered and redirects on success', async () => {
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: {
        user: { id: '1', email: 'a@b.com', display_name: 'Admin' },
        access_token: 'jwt-tok',
        token_type: 'Bearer',
        expires_in: 900,
      },
    });

    render(<TwoFactorForm {...defaultProps} />);

    // InputOTP renders a single input; type 6 digits
    const input = document.querySelector('input[data-input-otp="true"]') ?? screen.getByRole('textbox');
    await userEvent.click(input);
    await userEvent.keyboard('123456');

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith(
        '/v1/auth/2fa/validate',
        { code: '123456' },
        { headers: { Authorization: 'Bearer test-temp-token' } },
      );
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/dashboard');
    });
  });

  it('shows toast on invalid code', async () => {
    vi.mocked(api.post).mockRejectedValue(new ApiError(401, 'Invalid TOTP code'));

    render(<TwoFactorForm {...defaultProps} />);
    const input = document.querySelector('input[data-input-otp="true"]') ?? screen.getByRole('textbox');
    await userEvent.click(input);
    await userEvent.keyboard('000000');

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Invalid TOTP code');
    });
  });
});
