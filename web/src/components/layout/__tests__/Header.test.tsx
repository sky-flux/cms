import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Header } from '../Header';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('Header', () => {
  const mockUser = {
    displayName: 'John Doe',
    email: 'john@example.com',
    avatarUrl: undefined,
  };

  it('renders the site name', () => {
    render(<Header user={mockUser} siteName="My Blog" />);
    expect(screen.getByText('My Blog')).toBeInTheDocument();
  });

  it('renders default site name when not provided', () => {
    render(<Header user={mockUser} />);
    expect(screen.getByText('Sky Flux CMS')).toBeInTheDocument();
  });

  it('renders user display name in avatar fallback', () => {
    render(<Header user={mockUser} />);
    // Avatar fallback should show initials "JD"
    expect(screen.getByText('JD')).toBeInTheDocument();
  });

  it('renders sidebar toggle button', () => {
    const onToggle = vi.fn();
    render(<Header user={mockUser} onToggleSidebar={onToggle} />);
    expect(screen.getByRole('button', { name: /toggle/i })).toBeInTheDocument();
  });

  it('calls onToggleSidebar when hamburger clicked', async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();
    render(<Header user={mockUser} onToggleSidebar={onToggle} />);

    await user.click(screen.getByRole('button', { name: /toggle/i }));
    expect(onToggle).toHaveBeenCalledOnce();
  });

  it('calls onLogout when logout is triggered', async () => {
    const user = userEvent.setup();
    const onLogout = vi.fn();
    render(<Header user={mockUser} onLogout={onLogout} />);

    // Open the user dropdown
    const avatarButton = screen.getByRole('button', { name: /user menu/i });
    await user.click(avatarButton);

    const logoutItem = screen.getByText('header.logout');
    await user.click(logoutItem);

    expect(onLogout).toHaveBeenCalledOnce();
  });

  it('shows profile and settings in user dropdown', async () => {
    const user = userEvent.setup();
    render(<Header user={mockUser} />);

    const avatarButton = screen.getByRole('button', { name: /user menu/i });
    await user.click(avatarButton);

    expect(screen.getByText('header.profile')).toBeInTheDocument();
    expect(screen.getByText('header.settings')).toBeInTheDocument();
  });

  it('renders header element', () => {
    render(<Header user={mockUser} />);
    expect(screen.getByRole('banner')).toBeInTheDocument();
  });
});
