import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RolesTable } from '../RolesTable';
import { RoleFormDialog } from '../RoleFormDialog';
import { RolePermissions } from '../RolePermissions';
import type { Role, PaginationMeta } from '@/lib/system-api';

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

const mockRoles: Role[] = [
  {
    id: 'r1',
    name: 'Super Admin',
    slug: 'super',
    description: 'Full access to everything',
    built_in: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 'r2',
    name: 'Editor',
    slug: 'editor',
    description: 'Can edit and publish content',
    built_in: true,
    created_at: '2026-01-01T00:00:00Z',
    updated_at: '2026-01-01T00:00:00Z',
  },
  {
    id: 'r3',
    name: 'Custom Role',
    slug: 'custom-role',
    description: 'A custom role',
    built_in: false,
    created_at: '2026-02-01T00:00:00Z',
    updated_at: '2026-02-01T00:00:00Z',
  },
];

describe('RolesTable', () => {
  const defaultProps = {
    roles: mockRoles,
    loading: false,
    onEdit: vi.fn(),
    onPermissions: vi.fn(),
    onDelete: vi.fn(),
    onNewRole: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders column headers', () => {
    render(<RolesTable {...defaultProps} />);
    expect(screen.getByText('roleName')).toBeInTheDocument();
    expect(screen.getByText('roleSlug')).toBeInTheDocument();
    expect(screen.getByText('description')).toBeInTheDocument();
  });

  it('renders role data rows', () => {
    render(<RolesTable {...defaultProps} />);
    expect(screen.getByText('Super Admin')).toBeInTheDocument();
    expect(screen.getByText('super')).toBeInTheDocument();
    expect(screen.getByText('Editor')).toBeInTheDocument();
    expect(screen.getByText('editor')).toBeInTheDocument();
    expect(screen.getByText('Custom Role')).toBeInTheDocument();
    expect(screen.getByText('custom-role')).toBeInTheDocument();
  });

  it('renders built-in badge for built-in roles', () => {
    render(<RolesTable {...defaultProps} />);
    const builtInBadges = screen.getAllByText('builtIn');
    expect(builtInBadges.length).toBe(2); // Super Admin + Editor
  });

  it('renders custom badge for non-built-in roles', () => {
    render(<RolesTable {...defaultProps} />);
    expect(screen.getByText('custom')).toBeInTheDocument();
  });

  it('renders "New Role" button', () => {
    render(<RolesTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /newRole/i })).toBeInTheDocument();
  });

  it('calls onNewRole when "New Role" button is clicked', async () => {
    const user = userEvent.setup();
    render(<RolesTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newRole/i }));
    expect(defaultProps.onNewRole).toHaveBeenCalled();
  });

  it('shows loading skeletons when loading', () => {
    const { container } = render(<RolesTable {...defaultProps} loading={true} roles={[]} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('shows empty state when no roles', () => {
    render(<RolesTable {...defaultProps} roles={[]} />);
    expect(screen.getByText('noRolesFound')).toBeInTheDocument();
  });

  it('renders role descriptions', () => {
    render(<RolesTable {...defaultProps} />);
    expect(screen.getByText('Full access to everything')).toBeInTheDocument();
    expect(screen.getByText('Can edit and publish content')).toBeInTheDocument();
  });
});

describe('RoleFormDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    onSubmit: vi.fn(),
    loading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create mode with empty fields', () => {
    render(<RoleFormDialog {...defaultProps} />);
    expect(screen.getByText('newRole')).toBeInTheDocument();
    const nameInput = screen.getByLabelText('roleName');
    expect(nameInput).toHaveValue('');
  });

  it('requires name and slug in create mode', async () => {
    const user = userEvent.setup();
    render(<RoleFormDialog {...defaultProps} />);
    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);
    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('renders edit mode with pre-filled fields', () => {
    const editRole = mockRoles[2]; // Custom Role
    render(<RoleFormDialog {...defaultProps} role={editRole} />);
    expect(screen.getByText('editRole')).toBeInTheDocument();
    const nameInput = screen.getByLabelText('roleName');
    expect(nameInput).toHaveValue('Custom Role');
  });

  it('disables slug field in edit mode', () => {
    const editRole = mockRoles[2];
    render(<RoleFormDialog {...defaultProps} role={editRole} />);
    const slugInput = screen.getByLabelText('roleSlug');
    expect(slugInput).toBeDisabled();
  });

  it('calls onSubmit with form data in create mode', async () => {
    const user = userEvent.setup();
    render(<RoleFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('roleName'), 'Reviewer');
    await user.type(screen.getByLabelText('roleSlug'), 'reviewer');
    await user.type(screen.getByLabelText('description'), 'Can review content');

    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Reviewer',
          slug: 'reviewer',
          description: 'Can review content',
        }),
      );
    });
  });

  it('calls onSubmit with form data in edit mode', async () => {
    const user = userEvent.setup();
    const editRole = mockRoles[2];
    render(<RoleFormDialog {...defaultProps} role={editRole} />);

    const nameInput = screen.getByLabelText('roleName');
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Role');

    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Updated Role',
        }),
      );
    });
  });

  it('validates slug format (lowercase, numbers, hyphens)', async () => {
    const user = userEvent.setup();
    render(<RoleFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('roleName'), 'Test Role');
    await user.type(screen.getByLabelText('roleSlug'), 'Invalid Slug!');

    const submitButton = screen.getByRole('button', { name: /save|create|submit/i });
    await user.click(submitButton);

    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('disables submit button when loading', () => {
    render(<RoleFormDialog {...defaultProps} loading={true} />);
    const submitButton = screen.getByRole('button', { name: /save|create|submit|loading/i });
    expect(submitButton).toBeDisabled();
  });
});

describe('RolePermissions', () => {
  const mockApis = [
    { id: 'api1', method: 'GET', path: '/api/v1/posts', description: 'List posts' },
    { id: 'api2', method: 'POST', path: '/api/v1/posts', description: 'Create post' },
    { id: 'api3', method: 'GET', path: '/api/v1/users', description: 'List users' },
  ];

  const mockMenus = [
    { id: 'm1', name: 'Dashboard', path: '/dashboard', icon: 'home', parent_id: null, sort_order: 1, children: [] },
    { id: 'm2', name: 'Content', path: '/content', icon: 'file', parent_id: null, sort_order: 2, children: [
      { id: 'm3', name: 'Posts', path: '/posts', icon: 'edit', parent_id: 'm2', sort_order: 1, children: [] },
    ]},
  ];

  const mockTemplates = [
    { id: 't1', name: 'Editor Template', description: 'Editor permissions', created_at: '' },
    { id: 't2', name: 'Admin Template', description: 'Admin permissions', created_at: '' },
  ];

  const defaultProps = {
    roleId: 'r1',
    roleName: 'Super Admin',
    apis: mockApis,
    menus: mockMenus,
    templates: mockTemplates,
    checkedApiIds: ['api1'],
    checkedMenuIds: ['m1'],
    onApiChange: vi.fn(),
    onMenuChange: vi.fn(),
    onSave: vi.fn(),
    onApplyTemplate: vi.fn(),
    saving: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders API Permissions and Menu Permissions tabs', () => {
    render(<RolePermissions {...defaultProps} />);
    expect(screen.getByText('apiPermissions')).toBeInTheDocument();
    expect(screen.getByText('menuPermissions')).toBeInTheDocument();
  });

  it('renders API endpoints as tree nodes', () => {
    render(<RolePermissions {...defaultProps} />);
    expect(screen.getByText('GET /api/v1/posts — List posts')).toBeInTheDocument();
    expect(screen.getByText('POST /api/v1/posts — Create post')).toBeInTheDocument();
  });

  it('renders save button', () => {
    render(<RolePermissions {...defaultProps} />);
    expect(screen.getByRole('button', { name: /save/i })).toBeInTheDocument();
  });

  it('renders apply template button', () => {
    render(<RolePermissions {...defaultProps} />);
    expect(screen.getByRole('button', { name: /applyTemplate/i })).toBeInTheDocument();
  });

  it('calls onSave when save button is clicked', async () => {
    const user = userEvent.setup();
    render(<RolePermissions {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /save/i }));
    expect(defaultProps.onSave).toHaveBeenCalled();
  });

  it('disables save button when saving', () => {
    render(<RolePermissions {...defaultProps} saving={true} />);
    const saveButton = screen.getByRole('button', { name: /save|loading/i });
    expect(saveButton).toBeDisabled();
  });

  it('switches to menu permissions tab', async () => {
    const user = userEvent.setup();
    render(<RolePermissions {...defaultProps} />);
    await user.click(screen.getByText('menuPermissions'));
    expect(screen.getByText('Dashboard')).toBeInTheDocument();
    expect(screen.getByText('Content')).toBeInTheDocument();
  });

  it('renders select all / deselect all buttons', () => {
    render(<RolePermissions {...defaultProps} />);
    const selectBtns = screen.getAllByRole('button', { name: /selectAll/i });
    expect(selectBtns.length).toBeGreaterThanOrEqual(1);
    const deselectBtns = screen.getAllByRole('button', { name: /deselectAll/i });
    expect(deselectBtns.length).toBeGreaterThanOrEqual(1);
  });

  it('calls onApiChange with all ids when Select All is clicked', async () => {
    const user = userEvent.setup();
    render(<RolePermissions {...defaultProps} />);
    // First "Select All" is in the API tab (visible by default)
    const selectBtns = screen.getAllByRole('button', { name: /selectAll/i });
    await user.click(selectBtns[0]);
    expect(defaultProps.onApiChange).toHaveBeenCalledWith(['api1', 'api2', 'api3']);
  });

  it('calls onApiChange with empty array when Deselect All is clicked', async () => {
    const user = userEvent.setup();
    render(<RolePermissions {...defaultProps} />);
    const deselectBtns = screen.getAllByRole('button', { name: /deselectAll/i });
    await user.click(deselectBtns[0]);
    expect(defaultProps.onApiChange).toHaveBeenCalledWith([]);
  });
});
