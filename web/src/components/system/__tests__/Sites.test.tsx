import React from 'react';
import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { QueryClient, QueryClientProvider } from '@tanstack/react-query';
import { SitesTable } from '../SitesTable';
import { SiteFormDialog } from '../SiteFormDialog';
import { SiteUsersDialog } from '../SiteUsersDialog';
import { SitesPage } from '../SitesPage';
import type { Site, SiteUser, PaginationMeta } from '@/lib/system-api';
import { sitesApi } from '@/lib/system-api';

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
    sitesApi: {
      list: vi.fn().mockResolvedValue({
        data: [],
        pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 },
      }),
      get: vi.fn().mockResolvedValue({ data: {} }),
      create: vi.fn().mockResolvedValue({ data: {} }),
      update: vi.fn().mockResolvedValue({ data: {} }),
      deleteSite: vi.fn().mockResolvedValue({ success: true }),
      listUsers: vi.fn().mockResolvedValue({
        data: [],
        pagination: { page: 1, per_page: 20, total: 0, total_pages: 1 },
      }),
      assignRole: vi.fn().mockResolvedValue({ data: {} }),
      removeRole: vi.fn().mockResolvedValue({ success: true }),
    },
    usersApi: {
      list: vi.fn().mockResolvedValue({
        data: [],
        pagination: { page: 1, per_page: 100, total: 0, total_pages: 1 },
      }),
    },
    rolesApi: {
      list: vi.fn().mockResolvedValue({ data: [] }),
    },
  };
});

vi.mock('sonner', () => ({
  toast: { success: vi.fn(), error: vi.fn() },
}));

const mockedSitesApi = vi.mocked(sitesApi);

const mockSites: Site[] = [
  {
    id: 's1',
    name: 'Main Blog',
    slug: 'main_blog',
    domain: 'blog.example.com',
    description: 'Main blog site',
    logo_url: null,
    default_locale: 'en',
    timezone: 'UTC',
    is_active: true,
    settings: {},
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 's2',
    name: 'Dev Portal',
    slug: 'dev_portal',
    domain: null,
    description: null,
    logo_url: null,
    default_locale: 'zh-CN',
    timezone: 'Asia/Shanghai',
    is_active: false,
    settings: {},
    created_at: '2026-01-10T00:00:00Z',
    updated_at: '2026-01-10T00:00:00Z',
  },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 2,
  total_pages: 1,
};

const mockSiteUsers: SiteUser[] = [
  {
    user: {
      id: 'u1',
      email: 'alice@example.com',
      display_name: 'Alice',
      avatar_url: null,
      is_active: true,
    },
    role: 'admin',
    created_at: '2026-01-01T00:00:00Z',
  },
  {
    user: {
      id: 'u2',
      email: 'bob@example.com',
      display_name: 'Bob',
      avatar_url: null,
      is_active: true,
    },
    role: 'editor',
    created_at: '2026-01-02T00:00:00Z',
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

// --- SitesTable tests ---
describe('SitesTable', () => {
  const defaultProps = {
    sites: mockSites,
    pagination: mockPagination,
    loading: false,
    onPageChange: vi.fn(),
    onSearch: vi.fn(),
    onEdit: vi.fn(),
    onManageUsers: vi.fn(),
    onDelete: vi.fn(),
    onNewSite: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders column headers', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('siteName')).toBeInTheDocument();
    expect(screen.getByText('slug')).toBeInTheDocument();
    expect(screen.getByText('domain')).toBeInTheDocument();
    expect(screen.getByText('status')).toBeInTheDocument();
    expect(screen.getByText('timezone')).toBeInTheDocument();
  });

  it('renders site names', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('Main Blog')).toBeInTheDocument();
    expect(screen.getByText('Dev Portal')).toBeInTheDocument();
  });

  it('renders site slugs', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('main_blog')).toBeInTheDocument();
    expect(screen.getByText('dev_portal')).toBeInTheDocument();
  });

  it('renders domain or "--" for null domains', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('blog.example.com')).toBeInTheDocument();
    expect(screen.getByText('--')).toBeInTheDocument();
  });

  it('renders active/inactive status badges', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('active')).toBeInTheDocument();
    expect(screen.getByText('inactive')).toBeInTheDocument();
  });

  it('renders timezone values', () => {
    render(<SitesTable {...defaultProps} />);
    expect(screen.getByText('UTC')).toBeInTheDocument();
    expect(screen.getByText('Asia/Shanghai')).toBeInTheDocument();
  });

  it('calls onEdit from row actions', async () => {
    const user = userEvent.setup();
    render(<SitesTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    const editItem = screen.getByText('edit');
    await user.click(editItem);
    expect(defaultProps.onEdit).toHaveBeenCalledWith(mockSites[0]);
  });

  it('calls onManageUsers from row actions', async () => {
    const user = userEvent.setup();
    render(<SitesTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    const manageUsersItem = screen.getByText('manageUsers');
    await user.click(manageUsersItem);
    expect(defaultProps.onManageUsers).toHaveBeenCalledWith(mockSites[0]);
  });

  it('calls onDelete from row actions', async () => {
    const user = userEvent.setup();
    render(<SitesTable {...defaultProps} />);
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    const deleteItem = screen.getByText('delete');
    await user.click(deleteItem);
    expect(defaultProps.onDelete).toHaveBeenCalledWith(mockSites[0]);
  });

  it('shows empty state when no sites', () => {
    render(<SitesTable {...defaultProps} sites={[]} />);
    expect(screen.getByText('noSitesFound')).toBeInTheDocument();
  });

  it('shows loading skeletons', () => {
    const { container } = render(<SitesTable {...defaultProps} loading={true} sites={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('calls onNewSite when button clicked', async () => {
    const user = userEvent.setup();
    render(<SitesTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newSite/i }));
    expect(defaultProps.onNewSite).toHaveBeenCalled();
  });
});

// --- SiteFormDialog tests ---
describe('SiteFormDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create mode without is_active toggle', () => {
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
      />
    );
    expect(screen.getByText('newSite')).toBeInTheDocument();
    expect(screen.queryByLabelText(/active/i)).not.toBeInTheDocument();
  });

  it('renders edit mode with slug disabled and is_active toggle', () => {
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        site={mockSites[0]}
      />
    );
    expect(screen.getByText('editSite')).toBeInTheDocument();
    const slugInput = screen.getByLabelText('slug');
    expect(slugInput).toBeDisabled();
  });

  it('validates name is required', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        loading={false}
      />
    );
    // Fill slug but not name
    const slugInput = screen.getByLabelText('slug');
    await user.type(slugInput, 'test_slug');
    // Submit
    await user.click(screen.getByRole('button', { name: /save/i }));
    await waitFor(() => {
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  it('validates slug regex pattern', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        loading={false}
      />
    );
    const nameInput = screen.getByLabelText('siteName');
    const slugInput = screen.getByLabelText('slug');
    await user.type(nameInput, 'Test Site');
    await user.type(slugInput, 'NO-CAPITALS!');
    await user.click(screen.getByRole('button', { name: /save/i }));
    await waitFor(() => {
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  it('calls onSubmit with form data on valid submit', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        loading={false}
      />
    );
    await user.type(screen.getByLabelText('siteName'), 'New Site');
    await user.type(screen.getByLabelText('slug'), 'new_site');
    await user.click(screen.getByRole('button', { name: /save/i }));
    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Site',
          slug: 'new_site',
        })
      );
    });
  });

  it('populates fields in edit mode', () => {
    render(
      <SiteFormDialog
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        loading={false}
        site={mockSites[0]}
      />
    );
    expect(screen.getByLabelText('siteName')).toHaveValue('Main Blog');
    expect(screen.getByLabelText('slug')).toHaveValue('main_blog');
    expect(screen.getByLabelText('domain')).toHaveValue('blog.example.com');
  });
});

// --- SiteUsersDialog tests ---
describe('SiteUsersDialog', () => {
  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders user list with role badges', () => {
    render(
      <SiteUsersDialog
        open={true}
        onOpenChange={vi.fn()}
        siteUsers={mockSiteUsers}
        loading={false}
        onAssignRole={vi.fn()}
        onRemoveUser={vi.fn()}
        assignLoading={false}
      />,
      { wrapper: createWrapper() }
    );
    expect(screen.getByText('Alice')).toBeInTheDocument();
    expect(screen.getByText('Bob')).toBeInTheDocument();
    // Both 'admin' and 'editor' should appear as role badges
    // 'editor' also appears in the select, so use getAllByText
    expect(screen.getByText('admin')).toBeInTheDocument();
    expect(screen.getAllByText('editor').length).toBeGreaterThanOrEqual(1);
  });

  it('renders email addresses', () => {
    render(
      <SiteUsersDialog
        open={true}
        onOpenChange={vi.fn()}
        siteUsers={mockSiteUsers}
        loading={false}
        onAssignRole={vi.fn()}
        onRemoveUser={vi.fn()}
        assignLoading={false}
      />,
      { wrapper: createWrapper() }
    );
    expect(screen.getByText('alice@example.com')).toBeInTheDocument();
    expect(screen.getByText('bob@example.com')).toBeInTheDocument();
  });

  it('shows remove button for each user', () => {
    render(
      <SiteUsersDialog
        open={true}
        onOpenChange={vi.fn()}
        siteUsers={mockSiteUsers}
        loading={false}
        onAssignRole={vi.fn()}
        onRemoveUser={vi.fn()}
        assignLoading={false}
      />,
      { wrapper: createWrapper() }
    );
    const removeButtons = screen.getAllByRole('button', { name: /removeUser/i });
    expect(removeButtons).toHaveLength(2);
  });

  it('calls onRemoveUser when remove button clicked', async () => {
    const user = userEvent.setup();
    const onRemoveUser = vi.fn();
    render(
      <SiteUsersDialog
        open={true}
        onOpenChange={vi.fn()}
        siteUsers={mockSiteUsers}
        loading={false}
        onAssignRole={vi.fn()}
        onRemoveUser={onRemoveUser}
        assignLoading={false}
      />,
      { wrapper: createWrapper() }
    );
    const removeButtons = screen.getAllByRole('button', { name: /removeUser/i });
    await user.click(removeButtons[0]);
    expect(onRemoveUser).toHaveBeenCalledWith('u1');
  });

  it('shows empty state when no users', () => {
    render(
      <SiteUsersDialog
        open={true}
        onOpenChange={vi.fn()}
        siteUsers={[]}
        loading={false}
        onAssignRole={vi.fn()}
        onRemoveUser={vi.fn()}
        assignLoading={false}
      />,
      { wrapper: createWrapper() }
    );
    expect(screen.getByText('noUsersFound')).toBeInTheDocument();
  });
});

// --- SitesPage (integration) tests ---
describe('SitesPage', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockedSitesApi.list.mockResolvedValue({
      success: true,
      data: mockSites,
      pagination: mockPagination,
    });
  });

  it('renders the page title', async () => {
    render(<SitesPage />);
    expect(screen.getByText('title')).toBeInTheDocument();
  });

  it('loads and displays sites', async () => {
    render(<SitesPage />);
    await waitFor(() => {
      expect(screen.getByText('Main Blog')).toBeInTheDocument();
    });
  });

  it('delete site requires typing slug to confirm', async () => {
    const user = userEvent.setup();
    render(<SitesPage />);

    await waitFor(() => {
      expect(screen.getByText('Main Blog')).toBeInTheDocument();
    });

    // Open action menu for the first site
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);

    // Click delete
    const deleteItem = screen.getByText('delete');
    await user.click(deleteItem);

    // The confirm dialog should appear and the confirm button should be disabled
    await waitFor(() => {
      expect(screen.getByText('confirmSlug')).toBeInTheDocument();
    });

    const confirmButton = screen.getByRole('button', { name: /confirm/i });
    expect(confirmButton).toBeDisabled();

    // Type the wrong slug
    const slugInput = screen.getByPlaceholderText('main_blog');
    await user.type(slugInput, 'wrong_slug');
    expect(confirmButton).toBeDisabled();

    // Clear and type correct slug
    await user.clear(slugInput);
    await user.type(slugInput, 'main_blog');
    expect(confirmButton).not.toBeDisabled();
  });
});
