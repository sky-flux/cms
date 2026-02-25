import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { describe, expect, it, vi } from 'vitest';
import { LocaleSwitcher } from '../LocaleSwitcher';

const changeLanguageMock = vi.fn();

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => key,
    i18n: {
      language: 'en',
      changeLanguage: changeLanguageMock,
    },
  }),
}));

describe('LocaleSwitcher', () => {
  it('renders a language button', () => {
    render(<LocaleSwitcher />);
    expect(screen.getByRole('button', { name: /language/i })).toBeInTheDocument();
  });

  it('shows dropdown with language options when clicked', async () => {
    const user = userEvent.setup();
    render(<LocaleSwitcher />);

    await user.click(screen.getByRole('button', { name: /language/i }));

    expect(screen.getByText('English')).toBeInTheDocument();
    expect(screen.getByText(/中文/)).toBeInTheDocument();
  });

  it('calls changeLanguage when a language is selected', async () => {
    const user = userEvent.setup();
    render(<LocaleSwitcher />);

    await user.click(screen.getByRole('button', { name: /language/i }));
    await user.click(screen.getByText(/中文/));

    expect(changeLanguageMock).toHaveBeenCalledWith('zh-CN');
  });
});
