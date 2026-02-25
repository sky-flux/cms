import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { Sidebar } from '../Sidebar';
import { adminNavSections } from '../nav-items';

// Mock i18next - translate keys as-is for testing
vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('Sidebar', () => {
  const defaultProps = {
    sections: adminNavSections,
    currentPath: '/dashboard',
  };

  it('renders all nav items from sections', () => {
    render(<Sidebar {...defaultProps} />);

    for (const section of adminNavSections) {
      for (const item of section.items) {
        expect(screen.getByText(item.label)).toBeInTheDocument();
      }
    }
  });

  it('renders section titles', () => {
    render(<Sidebar {...defaultProps} />);

    expect(screen.getByText('nav.content')).toBeInTheDocument();
    expect(screen.getByText('nav.system')).toBeInTheDocument();
  });

  it('highlights the active item based on currentPath', () => {
    render(<Sidebar {...defaultProps} currentPath="/dashboard/posts" />);

    const postsLink = screen.getByRole('link', { name: /nav\.posts/ });
    expect(postsLink).toHaveAttribute('data-active', 'true');

    const dashboardLink = screen.getByRole('link', { name: /nav\.dashboard/ });
    expect(dashboardLink).toHaveAttribute('data-active', 'false');
  });

  it('renders nav items as links with correct href', () => {
    render(<Sidebar {...defaultProps} />);

    const dashboardLink = screen.getByRole('link', { name: /nav\.dashboard/ });
    expect(dashboardLink).toHaveAttribute('href', '/dashboard');

    const postsLink = screen.getByRole('link', { name: /nav\.posts/ });
    expect(postsLink).toHaveAttribute('href', '/dashboard/posts');
  });

  it('hides labels when collapsed', () => {
    render(<Sidebar {...defaultProps} collapsed />);

    // Labels should have sr-only class when collapsed
    const labels = screen.getAllByText('nav.dashboard');
    expect(labels[0].closest('[data-collapsed="true"]')).toBeTruthy();
  });

  it('calls onToggle when toggle button is clicked', async () => {
    const user = userEvent.setup();
    const onToggle = vi.fn();

    render(<Sidebar {...defaultProps} onToggle={onToggle} />);

    const toggleButton = screen.getByRole('button', { name: /toggle/i });
    await user.click(toggleButton);

    expect(onToggle).toHaveBeenCalledOnce();
  });

  it('renders navigation element with correct role', () => {
    render(<Sidebar {...defaultProps} />);
    expect(screen.getByRole('navigation')).toBeInTheDocument();
  });
});
