import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { SettingsForm } from '../SettingsForm';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({ t: (key: string) => key }),
}));

const mockSettings = [
  { key: 'site_name', value: 'Test Site', description: 'Site display name' },
  { key: 'site_url', value: 'https://test.com', description: 'Public URL' },
];

describe('SettingsForm error handling', () => {
  it('should display error message for a specific key', () => {
    render(
      <SettingsForm
        settings={mockSettings}
        onSave={vi.fn()}
        savingKey={null}
        errors={{ site_name: 'Site name is required' }}
      />,
    );
    expect(screen.getByText('Site name is required')).toBeInTheDocument();
  });

  it('should not show error for keys without errors', () => {
    render(
      <SettingsForm
        settings={mockSettings}
        onSave={vi.fn()}
        savingKey={null}
        errors={{ site_name: 'Error here' }}
      />,
    );
    // Only one error message should exist
    expect(screen.getAllByText(/Error here/)).toHaveLength(1);
  });

  it('should mark input as aria-invalid when error exists', () => {
    render(
      <SettingsForm
        settings={mockSettings}
        onSave={vi.fn()}
        savingKey={null}
        errors={{ site_name: 'Site name is required' }}
      />,
    );
    const errorInput = screen.getByDisplayValue('Test Site');
    expect(errorInput).toHaveAttribute('aria-invalid', 'true');

    const okInput = screen.getByDisplayValue('https://test.com');
    expect(okInput).not.toHaveAttribute('aria-invalid', 'true');
  });

  it('should render normally without errors prop', () => {
    render(
      <SettingsForm
        settings={mockSettings}
        onSave={vi.fn()}
        savingKey={null}
      />,
    );
    expect(screen.getByDisplayValue('Test Site')).toBeInTheDocument();
    expect(screen.getByDisplayValue('https://test.com')).toBeInTheDocument();
  });
});
