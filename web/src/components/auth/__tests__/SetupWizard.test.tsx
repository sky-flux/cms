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
import { SetupWizard } from '../SetupWizard';

describe('SetupWizard', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    Object.defineProperty(window, 'location', {
      value: { href: '', assign: vi.fn() },
      writable: true,
    });
  });

  it('renders step 1 (admin account) by default', () => {
    render(<SetupWizard />);
    expect(screen.getByText(/auth\.setupStep1/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
    expect(screen.getByLabelText(/auth\.setupAdminEmail/i)).toBeInTheDocument();
  });

  it('renders title and description', () => {
    render(<SetupWizard />);
    expect(screen.getByText(/auth\.setupTitle/i)).toBeInTheDocument();
  });

  it('validates step 1 before allowing next', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Click next without filling fields
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Should still be on step 1
    expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
  });

  it('navigates to step 2 after valid step 1', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    await waitFor(() => {
      expect(screen.getByText(/auth\.setupStep2/i)).toBeInTheDocument();
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });
  });

  it('can go back from step 2 to step 1', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Fill step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });

    // Go back
    await user.click(screen.getByRole('button', { name: /auth\.setupBack/i }));

    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupAdminUsername/i)).toBeInTheDocument();
    });
  });

  it('navigates to step 3 (review) after valid step 2', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument();
    });
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 -- review
    await waitFor(() => {
      expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument();
      expect(screen.getByText('Admin')).toBeInTheDocument();
      expect(screen.getByText('admin@test.com')).toBeInTheDocument();
      expect(screen.getByText('My Site')).toBeInTheDocument();
    });
  });

  it('calls setup API on install and redirects on success', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockResolvedValue({
      success: true,
      data: { user: {}, site: {}, access_token: 'tok' },
    });

    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 -- install
    await waitFor(() => expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /auth\.setupInstall/i }));

    await waitFor(() => {
      expect(api.post).toHaveBeenCalledWith('/v1/setup/initialize', {
        admin_display_name: 'Admin',
        admin_email: 'admin@test.com',
        admin_password: 'password123',
        site_name: 'My Site',
        site_slug: 'my-site',
        site_url: 'https://example.com',
        locale: 'en',
      });
    });

    await waitFor(() => {
      expect(window.location.href).toBe('/setup/complete');
    });
  });

  it('shows toast on API error during install', async () => {
    const user = userEvent.setup();
    vi.mocked(api.post).mockRejectedValue(new ApiError(500, 'Install failed'));

    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'my-site');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 3 -- install
    await waitFor(() => expect(screen.getByText(/auth\.setupStep3/i)).toBeInTheDocument());
    await user.click(screen.getByRole('button', { name: /auth\.setupInstall/i }));

    await waitFor(() => {
      expect(toast.error).toHaveBeenCalledWith('Install failed');
    });
  });

  it('validates slug format on step 2', async () => {
    const user = userEvent.setup();
    render(<SetupWizard />);

    // Step 1
    await user.type(screen.getByLabelText(/auth\.setupAdminUsername/i), 'Admin');
    await user.type(screen.getByLabelText(/auth\.setupAdminEmail/i), 'admin@test.com');
    await user.type(screen.getByLabelText(/auth\.password/i), 'password123');
    await user.type(screen.getByLabelText(/auth\.confirmPassword/i), 'password123');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Step 2 -- invalid slug
    await waitFor(() => expect(screen.getByLabelText(/auth\.setupSiteName/i)).toBeInTheDocument());
    await user.type(screen.getByLabelText(/auth\.setupSiteName/i), 'My Site');
    await user.type(screen.getByLabelText(/auth\.setupSiteSlug/i), 'INVALID SLUG!');
    await user.type(screen.getByLabelText(/auth\.setupSiteUrl/i), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /auth\.setupNext/i }));

    // Should stay on step 2
    await waitFor(() => {
      expect(screen.getByLabelText(/auth\.setupSiteSlug/i)).toBeInTheDocument();
    });
    expect(screen.queryByText(/auth\.setupStep3/i)).not.toBeInTheDocument();
  });
});
