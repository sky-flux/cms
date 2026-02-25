import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SettingsForm } from '../SettingsForm';
import { SettingsPage } from '../SettingsPage';
import type { SettingItem } from '@/lib/system-api';
import { settingsApi } from '@/lib/system-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const label = key.split('.').pop() || key;
      return params
        ? label.replace(/\{\{(\w+)\}\}/g, (_: string, k: string) => String(params[k]))
        : label;
    },
  }),
  initReactI18next: { type: '3rdParty', init: () => {} },
  I18nextProvider: ({ children }: { children: React.ReactNode }) => children,
}));

vi.mock('@/lib/system-api', async () => {
  const actual = await vi.importActual<object>('@/lib/system-api');
  return {
    ...actual,
    settingsApi: {
      get: vi.fn().mockResolvedValue({
        data: [],
      }),
      update: vi.fn().mockResolvedValue({
        data: { key: 'site_name', value: 'New Value', description: '' },
      }),
    },
  };
});

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const mockedSettingsApi = vi.mocked(settingsApi);

const mockSettings: SettingItem[] = [
  {
    key: 'site_name',
    value: 'My CMS',
    description: 'The name of the site',
  },
  {
    key: 'site_description',
    value: 'A content management system',
    description: 'A brief description of the site',
  },
  {
    key: 'posts_per_page',
    value: '20',
    description: 'Number of posts displayed per page',
  },
];

function createWrapper() {
  const queryClient = new QueryClient({
    defaultOptions: { queries: { retry: false } },
  });
  return function Wrapper({ children }: { children: React.ReactNode }) {
    return <QueryClientProvider client={queryClient}>{children}</QueryClientProvider>;
  };
}

// --- SettingsForm tests ---
describe('SettingsForm', () => {
  const defaultProps = {
    settings: mockSettings,
    onSave: vi.fn(),
    savingKey: null as string | null,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders config items as key-value pairs', () => {
    render(<SettingsForm {...defaultProps} />);
    expect(screen.getByText('site_name')).toBeInTheDocument();
    expect(screen.getByText('site_description')).toBeInTheDocument();
    expect(screen.getByText('posts_per_page')).toBeInTheDocument();
    expect(screen.getByDisplayValue('My CMS')).toBeInTheDocument();
    expect(screen.getByDisplayValue('A content management system')).toBeInTheDocument();
    expect(screen.getByDisplayValue('20')).toBeInTheDocument();
  });

  it('shows description for each setting', () => {
    render(<SettingsForm {...defaultProps} />);
    expect(screen.getByText('The name of the site')).toBeInTheDocument();
    expect(screen.getByText('A brief description of the site')).toBeInTheDocument();
    expect(screen.getByText('Number of posts displayed per page')).toBeInTheDocument();
  });

  it('calls onSave with key+value on individual save', async () => {
    const user = userEvent.setup();
    const onSave = vi.fn();
    render(<SettingsForm {...defaultProps} onSave={onSave} />);

    // Change the value of site_name
    const input = screen.getByDisplayValue('My CMS');
    await user.clear(input);
    await user.type(input, 'Updated Name');

    // Click the save button for site_name
    const saveButtons = screen.getAllByRole('button', { name: /save/i });
    await user.click(saveButtons[0]);

    expect(onSave).toHaveBeenCalledWith('site_name', 'Updated Name');
  });

  it('shows individual save button per setting', () => {
    render(<SettingsForm {...defaultProps} />);
    const saveButtons = screen.getAllByRole('button', { name: /save/i });
    expect(saveButtons).toHaveLength(3);
  });

  it('disables save button while saving a specific key', () => {
    render(<SettingsForm {...defaultProps} savingKey="site_name" />);
    const saveButtons = screen.getAllByRole('button', { name: /save|loading/i });
    // The first button (for site_name) should be disabled
    expect(saveButtons[0]).toBeDisabled();
  });

  it('shows empty state when no settings', () => {
    render(<SettingsForm settings={[]} onSave={vi.fn()} savingKey={null} />);
    expect(screen.getByText('noSettings')).toBeInTheDocument();
  });
});

// --- SettingsPage tests ---
describe('SettingsPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedSettingsApi.get.mockResolvedValue({
      success: true,
      data: mockSettings,
    });
  });

  it('renders the page title', async () => {
    render(<SettingsPage />);
    expect(screen.getByText('title')).toBeInTheDocument();
  });

  it('loads settings and displays them', async () => {
    render(<SettingsPage />);
    await waitFor(() => {
      expect(screen.getByDisplayValue('My CMS')).toBeInTheDocument();
    });
  });

  it('handles update when save clicked', async () => {
    const user = userEvent.setup();
    mockedSettingsApi.update.mockResolvedValue({
      success: true,
      data: { key: 'site_name', value: 'Updated', description: 'The name of the site' },
    });

    render(<SettingsPage />);

    await waitFor(() => {
      expect(screen.getByDisplayValue('My CMS')).toBeInTheDocument();
    });

    // Modify and save
    const input = screen.getByDisplayValue('My CMS');
    await user.clear(input);
    await user.type(input, 'Updated');

    const saveButtons = screen.getAllByRole('button', { name: /save/i });
    await user.click(saveButtons[0]);

    await waitFor(() => {
      expect(mockedSettingsApi.update).toHaveBeenCalledWith('site_name', 'Updated');
    });
  });
});
