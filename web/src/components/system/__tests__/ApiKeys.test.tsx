import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { ApiKeysTable } from '../ApiKeysTable';
import { CreateApiKeyDialog } from '../CreateApiKeyDialog';
import { ApiKeysPage } from '../ApiKeysPage';
import type { ApiKey, CreateApiKeyResponse } from '@/lib/system-api';
import { apiKeysApi } from '@/lib/system-api';

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
    apiKeysApi: {
      list: vi.fn().mockResolvedValue({
        data: [],
      }),
      create: vi.fn().mockResolvedValue({
        data: {
          id: 'new-key-1',
          name: 'Test Key',
          key: 'sk_live_abc123def456ghi789jkl012mno345pqr678stu901vwx234',
          key_prefix: 'sk_live_abc',
          expires_at: null,
          rate_limit: 100,
          created_at: '2026-02-01T00:00:00Z',
        },
      }),
      delete: vi.fn().mockResolvedValue({ success: true }),
    },
  };
});

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const mockedApiKeysApi = vi.mocked(apiKeysApi);

const mockApiKeys: ApiKey[] = [
  {
    id: 'k1',
    name: 'Production Key',
    key_prefix: 'sk_live_abc',
    is_active: true,
    last_used_at: '2026-02-15T10:00:00Z',
    expires_at: '2027-01-01T00:00:00Z',
    rate_limit: 100,
    created_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 'k2',
    name: 'Test Key',
    key_prefix: 'sk_test_xyz',
    is_active: false,
    last_used_at: null,
    expires_at: null,
    rate_limit: 50,
    created_at: '2026-01-10T00:00:00Z',
  },
  {
    id: 'k3',
    name: 'Dev Key',
    key_prefix: 'sk_dev_def',
    is_active: true,
    last_used_at: '2026-02-20T08:00:00Z',
    expires_at: null,
    rate_limit: 200,
    created_at: '2026-01-15T00:00:00Z',
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

// --- ApiKeysTable tests ---
describe('ApiKeysTable', () => {
  const defaultProps = {
    apiKeys: mockApiKeys,
    loading: false,
    onRevoke: vi.fn(),
    onNewKey: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders column headers', () => {
    render(<ApiKeysTable {...defaultProps} />);
    expect(screen.getByText('keyName')).toBeInTheDocument();
    expect(screen.getByText('keyPrefix')).toBeInTheDocument();
    expect(screen.getByText('status')).toBeInTheDocument();
    expect(screen.getByText('lastUsed')).toBeInTheDocument();
    expect(screen.getByText('expiresAt')).toBeInTheDocument();
    expect(screen.getByText('rateLimit')).toBeInTheDocument();
  });

  it('renders api key names', () => {
    render(<ApiKeysTable {...defaultProps} />);
    expect(screen.getByText('Production Key')).toBeInTheDocument();
    expect(screen.getByText('Test Key')).toBeInTheDocument();
    expect(screen.getByText('Dev Key')).toBeInTheDocument();
  });

  it('renders key prefixes in monospace', () => {
    render(<ApiKeysTable {...defaultProps} />);
    expect(screen.getByText('sk_live_abc')).toBeInTheDocument();
    expect(screen.getByText('sk_test_xyz')).toBeInTheDocument();
  });

  it('renders active/revoked status badges', () => {
    render(<ApiKeysTable {...defaultProps} />);
    // Two active keys and one revoked
    const activeBadges = screen.getAllByText('active');
    expect(activeBadges).toHaveLength(2);
    expect(screen.getByText('revoked')).toBeInTheDocument();
  });

  it('shows "Never" for null last_used_at', () => {
    render(<ApiKeysTable {...defaultProps} />);
    expect(screen.getByText('never')).toBeInTheDocument();
  });

  it('shows "No expiry" for null expires_at', () => {
    render(<ApiKeysTable {...defaultProps} />);
    const noExpiryTexts = screen.getAllByText('noExpiry');
    expect(noExpiryTexts).toHaveLength(2);
  });

  it('renders rate limit values', () => {
    render(<ApiKeysTable {...defaultProps} />);
    expect(screen.getByText('100')).toBeInTheDocument();
    expect(screen.getByText('50')).toBeInTheDocument();
    expect(screen.getByText('200')).toBeInTheDocument();
  });

  it('shows "Revoke" action for active keys', async () => {
    const user = userEvent.setup();
    render(<ApiKeysTable {...defaultProps} />);
    // Click action menu on first (active) key
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    expect(screen.getByText('revokeKey')).toBeInTheDocument();
  });

  it('calls onRevoke when revoke action clicked', async () => {
    const user = userEvent.setup();
    render(<ApiKeysTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    const revokeItem = screen.getByText('revokeKey');
    await user.click(revokeItem);
    expect(defaultProps.onRevoke).toHaveBeenCalledWith(mockApiKeys[0]);
  });

  it('shows empty state when no keys', () => {
    render(<ApiKeysTable {...defaultProps} apiKeys={[]} />);
    expect(screen.getByText('noKeysFound')).toBeInTheDocument();
  });

  it('shows loading skeletons', () => {
    const { container } = render(<ApiKeysTable {...defaultProps} loading={true} apiKeys={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

// --- CreateApiKeyDialog tests ---
describe('CreateApiKeyDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders the create form in phase 1', () => {
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        createdKey={null}
        onAcknowledge={vi.fn()}
      />
    );
    expect(screen.getByText('newKey')).toBeInTheDocument();
    expect(screen.getByLabelText('keyName')).toBeInTheDocument();
  });

  it('validates name is required', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        loading={false}
        createdKey={null}
        onAcknowledge={vi.fn()}
      />
    );
    // Try to submit without name
    await user.click(screen.getByRole('button', { name: /create/i }));
    await waitFor(() => {
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  it('calls onSubmit with form data on valid submit', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        loading={false}
        createdKey={null}
        onAcknowledge={vi.fn()}
      />
    );
    await user.type(screen.getByLabelText('keyName'), 'My API Key');
    await user.click(screen.getByRole('button', { name: /create/i }));
    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({ name: 'My API Key' })
      );
    });
  });

  it('shows key after creation with copy button (phase 2)', () => {
    const createdKey: CreateApiKeyResponse = {
      id: 'new-key-1',
      name: 'Test Key',
      key: 'sk_live_abc123def456ghi789jkl012mno345pqr678stu901vwx234',
      key_prefix: 'sk_live_abc',
      expires_at: null,
      rate_limit: 100,
      created_at: '2026-02-01T00:00:00Z',
    };
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        createdKey={createdKey}
        onAcknowledge={vi.fn()}
      />
    );
    expect(screen.getByText('keyCreated')).toBeInTheDocument();
    expect(screen.getByText(createdKey.key)).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /copyKey/i })).toBeInTheDocument();
  });

  it('requires acknowledging before close in phase 2', () => {
    const createdKey: CreateApiKeyResponse = {
      id: 'new-key-1',
      name: 'Test Key',
      key: 'sk_live_abc123',
      key_prefix: 'sk_live_abc',
      expires_at: null,
      rate_limit: 100,
      created_at: '2026-02-01T00:00:00Z',
    };
    const onAcknowledge = vi.fn();
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        createdKey={createdKey}
        onAcknowledge={onAcknowledge}
      />
    );
    // The done button should be disabled until checkbox is checked
    const doneButton = screen.getByRole('button', { name: /done/i });
    expect(doneButton).toBeDisabled();
  });

  it('enables done button after checking acknowledgment', async () => {
    const user = userEvent.setup();
    const createdKey: CreateApiKeyResponse = {
      id: 'new-key-1',
      name: 'Test Key',
      key: 'sk_live_abc123',
      key_prefix: 'sk_live_abc',
      expires_at: null,
      rate_limit: 100,
      created_at: '2026-02-01T00:00:00Z',
    };
    const onAcknowledge = vi.fn();
    render(
      <CreateApiKeyDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        createdKey={createdKey}
        onAcknowledge={onAcknowledge}
      />
    );
    // Check the acknowledgment checkbox
    const checkbox = screen.getByRole('checkbox');
    await user.click(checkbox);
    // Now the done button should be enabled
    const doneButton = screen.getByRole('button', { name: /done/i });
    expect(doneButton).not.toBeDisabled();
    // Click done
    await user.click(doneButton);
    expect(onAcknowledge).toHaveBeenCalled();
  });
});

// --- ApiKeysPage tests ---
describe('ApiKeysPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedApiKeysApi.list.mockResolvedValue({
      success: true,
      data: mockApiKeys,
    });
  });

  it('renders the page title', async () => {
    render(<ApiKeysPage />);
    expect(screen.getByText('title')).toBeInTheDocument();
  });

  it('loads and displays API keys', async () => {
    render(<ApiKeysPage />);
    await waitFor(() => {
      expect(screen.getByText('Production Key')).toBeInTheDocument();
    });
  });
});
