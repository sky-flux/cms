import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { UsersTable } from '../UsersTable';
import { UserFormDialog } from '../UserFormDialog';
import type { User, Role, PaginationMeta } from '@/lib/system-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, params?: Record<string, unknown>) => {
      const label = key.split('.').pop() || key;
      return params
        ? label.replace(/\{\{(\w+)\}\}/g, (_: string, k: string) => String(params[k]))
        : label;
    },
  }),
}));

const mockUsers: User[] = [
  {
    id: 'u1',
    email: 'alice@example.com',
    display_name: 'Alice Admin',
    role: 'super',
    is_active: true,
    avatar_url: null,
    last_login_at: '2026-02-20T10:00:00Z',
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-02-20T10:00:00Z',
  },
  {
    id: 'u2',
    email: 'bob@example.com',
    display_name: 'Bob Editor',
    role: 'editor',
    is_active: true,
    avatar_url: null,
    last_login_at: null,
    created_at: '2026-01-15T00:00:00Z',
    updated_at: '2026-01-15T00:00:00Z',
  },
  {
    id: 'u3',
    email: 'charlie@example.com',
    display_name: 'Charlie Disabled',
    role: 'author',
    is_active: false,
    avatar_url: null,
    last_login_at: '2026-01-10T10:00:00Z',
    created_at: '2026-01-05T00:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
];

const mockRoles: Role[] = [
  { id: 'r1', name: 'Super Admin', slug: 'super', description: '', built_in: true, created_at: '', updated_at: '' },
  { id: 'r2', name: 'Editor', slug: 'editor', description: '', built_in: true, created_at: '', updated_at: '' },
  { id: 'r3', name: 'Author', slug: 'author', description: '', built_in: true, created_at: '', updated_at: '' },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 3,
  total_pages: 1,
};

describe('UsersTable', () => {
  const defaultProps = {
    users: mockUsers,
    roles: mockRoles,
    pagination: mockPagination,
    loading: false,
    onPageChange: vi.fn(),
    onRoleFilter: vi.fn(),
    onSearch: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onNewUser: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders column headers', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByText('email')).toBeInTheDocument();
    expect(screen.getByText('displayName')).toBeInTheDocument();
    expect(screen.getByText('role')).toBeInTheDocument();
    expect(screen.getByText('status')).toBeInTheDocument();
    expect(screen.getByText('lastLogin')).toBeInTheDocument();
    expect(screen.getByText('actions')).toBeInTheDocument();
  });

  it('renders user data rows', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByText('alice@example.com')).toBeInTheDocument();
    expect(screen.getByText('Alice Admin')).toBeInTheDocument();
    expect(screen.getByText('bob@example.com')).toBeInTheDocument();
    expect(screen.getByText('Bob Editor')).toBeInTheDocument();
    expect(screen.getByText('charlie@example.com')).toBeInTheDocument();
    expect(screen.getByText('Charlie Disabled')).toBeInTheDocument();
  });

  it('renders role badges for each user', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByText('super')).toBeInTheDocument();
    expect(screen.getByText('editor')).toBeInTheDocument();
    expect(screen.getByText('author')).toBeInTheDocument();
  });

  it('renders status badges (active/disabled)', () => {
    render(<UsersTable {...defaultProps} />);
    // 2 active users, 1 disabled
    const activeBadges = screen.getAllByText('active');
    expect(activeBadges.length).toBe(2);
    expect(screen.getByText('disabled')).toBeInTheDocument();
  });

  it('renders role filter select', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByText('filterByRole')).toBeInTheDocument();
  });

  it('renders search input', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument();
  });

  it('renders "New User" button', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /newUser/i })).toBeInTheDocument();
  });

  it('calls onNewUser when "New User" button is clicked', async () => {
    const user = userEvent.setup();
    render(<UsersTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newUser/i }));
    expect(defaultProps.onNewUser).toHaveBeenCalled();
  });

  it('calls onSearch when search input changes', async () => {
    const user = userEvent.setup();
    render(<UsersTable {...defaultProps} />);
    const searchInput = screen.getByPlaceholderText('searchPlaceholder');
    await user.type(searchInput, 'alice');
    expect(defaultProps.onSearch).toHaveBeenCalledWith('alice');
  });

  it('shows loading skeletons when loading', () => {
    const { container } = render(<UsersTable {...defaultProps} loading={true} users={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('shows empty state when no users', () => {
    render(<UsersTable {...defaultProps} users={[]} />);
    expect(screen.getByText('noUsersFound')).toBeInTheDocument();
  });

  it('formats last_login_at date', () => {
    render(<UsersTable {...defaultProps} />);
    const cells = screen.getAllByRole('cell');
    const dateCell = cells.find(
      (cell) => cell.textContent && cell.textContent.includes('2026'),
    );
    expect(dateCell).toBeTruthy();
  });

  it('shows "--" for users without last_login_at', () => {
    render(<UsersTable {...defaultProps} />);
    expect(screen.getAllByText('--').length).toBeGreaterThan(0);
  });

  it('renders pagination controls for multi-page results', () => {
    const multiPagePagination: PaginationMeta = {
      page: 1,
      per_page: 20,
      total: 50,
      total_pages: 3,
    };
    render(<UsersTable {...defaultProps} pagination={multiPagePagination} />);
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });
});

describe('UserFormDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    onSubmit: vi.fn(),
    roles: mockRoles,
    loading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create mode with empty fields', () => {
    render(<UserFormDialog {...defaultProps} />);
    expect(screen.getByText('newUser')).toBeInTheDocument();
    const emailInput = screen.getByLabelText('email');
    expect(emailInput).toHaveValue('');
    const nameInput = screen.getByLabelText('displayName');
    expect(nameInput).toHaveValue('');
  });

  it('requires email, display_name, and password in create mode', async () => {
    const user = userEvent.setup();
    render(<UserFormDialog {...defaultProps} />);
    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);
    // Form should not have called onSubmit since fields are empty
    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('renders edit mode with pre-filled fields', () => {
    const editUser = mockUsers[0];
    render(<UserFormDialog {...defaultProps} user={editUser} />);
    expect(screen.getByText('editUser')).toBeInTheDocument();
    const nameInput = screen.getByLabelText('displayName');
    expect(nameInput).toHaveValue('Alice Admin');
  });

  it('hides email and password fields in edit mode', () => {
    const editUser = mockUsers[0];
    render(<UserFormDialog {...defaultProps} user={editUser} />);
    expect(screen.queryByLabelText('email')).not.toBeInTheDocument();
    expect(screen.queryByLabelText('password')).not.toBeInTheDocument();
  });

  it('calls onSubmit with form data in create mode', async () => {
    const user = userEvent.setup();
    render(<UserFormDialog {...defaultProps} roles={mockRoles} defaultRole="editor" />);

    await user.type(screen.getByLabelText('email'), 'new@example.com');
    await user.type(screen.getByLabelText('displayName'), 'New User');
    await user.type(screen.getByLabelText('password'), 'password123');

    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          email: 'new@example.com',
          display_name: 'New User',
          password: 'password123',
          role: 'editor',
        }),
      );
    });
  });

  it('calls onSubmit with form data in edit mode', async () => {
    const user = userEvent.setup();
    const editUser = mockUsers[0];
    render(<UserFormDialog {...defaultProps} user={editUser} />);

    const nameInput = screen.getByLabelText('displayName');
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Admin');

    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          display_name: 'Updated Admin',
        }),
      );
    });
  });

  it('shows password help text in create mode', () => {
    render(<UserFormDialog {...defaultProps} />);
    expect(screen.getByText('passwordHelp')).toBeInTheDocument();
  });

  it('disables submit button when loading', () => {
    render(<UserFormDialog {...defaultProps} loading={true} />);
    const submitButton = screen.getByRole('button', { name: /save|create|submit|loading/i });
    expect(submitButton).toBeDisabled();
  });
});
