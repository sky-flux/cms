import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi, beforeEach } from 'vitest';
import { ThemeToggle } from '../ThemeToggle';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
  }),
}));

describe('ThemeToggle', () => {
  beforeEach(() => {
    localStorage.clear();
    document.documentElement.classList.remove('dark');
  });

  it('renders a toggle button', () => {
    render(<ThemeToggle />);
    expect(screen.getByRole('button', { name: /theme/i })).toBeInTheDocument();
  });

  it('cycles through themes: light -> dark -> system', async () => {
    const user = userEvent.setup();
    render(<ThemeToggle />);

    const button = screen.getByRole('button', { name: /theme/i });

    // Initial state: system (default)
    // Click -> light
    await user.click(button);
    expect(localStorage.getItem('sfc-theme')).toBe('light');

    // Click -> dark
    await user.click(button);
    expect(localStorage.getItem('sfc-theme')).toBe('dark');
    expect(document.documentElement.classList.contains('dark')).toBe(true);

    // Click -> system
    await user.click(button);
    expect(localStorage.getItem('sfc-theme')).toBe('system');
  });

  it('applies dark class for dark theme', async () => {
    const user = userEvent.setup();
    localStorage.setItem('sfc-theme', 'light');

    render(<ThemeToggle />);
    const button = screen.getByRole('button', { name: /theme/i });

    // light -> dark
    await user.click(button);
    expect(document.documentElement.classList.contains('dark')).toBe(true);
  });

  it('removes dark class for light theme', () => {
    document.documentElement.classList.add('dark');
    localStorage.setItem('sfc-theme', 'light');

    render(<ThemeToggle />);
    expect(document.documentElement.classList.contains('dark')).toBe(false);
  });
});
