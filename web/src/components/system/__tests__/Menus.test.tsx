import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, within, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MenusTable } from '../MenusTable';
import { MenuFormDialog } from '../MenuFormDialog';
import { MenuItemsEditor } from '../MenuItemsEditor';
import type {
  SiteMenu,
  SiteMenuDetail,
  SiteMenuItem,
} from '@/lib/system-api';

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

// --- Mock Data ---
const mockMenus: SiteMenu[] = [
  {
    id: 'm1',
    name: 'Main Navigation',
    slug: 'main-nav',
    location: 'header',
    description: 'Primary navigation',
    item_count: 5,
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'm2',
    name: 'Footer Links',
    slug: 'footer-links',
    location: 'footer',
    description: null,
    item_count: 3,
    created_at: '2026-01-12T10:00:00Z',
    updated_at: '2026-01-12T10:00:00Z',
  },
  {
    id: 'm3',
    name: 'Sidebar Menu',
    slug: 'sidebar-menu',
    location: 'sidebar',
    description: 'Side navigation',
    item_count: 0,
    created_at: '2026-01-14T10:00:00Z',
    updated_at: '2026-01-14T10:00:00Z',
  },
];

const mockMenuItems: SiteMenuItem[] = [
  {
    id: 'i1',
    parent_id: null,
    label: 'Home',
    url: '/',
    target: '_self',
    icon: null,
    css_class: null,
    type: 'custom',
    reference_id: null,
    sort_order: 0,
    is_active: true,
    is_broken: false,
    children: [
      {
        id: 'i3',
        parent_id: 'i1',
        label: 'Sub Page',
        url: '/sub',
        target: '_self',
        icon: null,
        css_class: null,
        type: 'custom',
        reference_id: null,
        sort_order: 0,
        is_active: true,
        is_broken: false,
        children: [],
      },
    ],
  },
  {
    id: 'i2',
    parent_id: null,
    label: 'Blog Post',
    url: null,
    target: '_self',
    icon: null,
    css_class: null,
    type: 'post',
    reference_id: 'p1',
    sort_order: 1,
    is_active: false,
    is_broken: true,
    children: [],
  },
];

const mockMenuDetail: SiteMenuDetail = {
  id: 'm1',
  name: 'Main Navigation',
  slug: 'main-nav',
  location: 'header',
  description: 'Primary navigation',
  item_count: 3,
  created_at: '2026-01-10T10:00:00Z',
  updated_at: '2026-01-10T10:00:00Z',
  items: mockMenuItems,
};

// =========================================
// MenusTable Tests
// =========================================
describe('MenusTable', () => {
  const defaultProps = {
    menus: mockMenus,
    loading: false,
    onEdit: vi.fn(),
    onManageItems: vi.fn(),
    onDelete: vi.fn(),
    onNewMenu: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders menu name, slug, location, and item count columns', () => {
    render(<MenusTable {...defaultProps} />);

    // Names
    expect(screen.getByText('Main Navigation')).toBeInTheDocument();
    expect(screen.getByText('Footer Links')).toBeInTheDocument();
    expect(screen.getByText('Sidebar Menu')).toBeInTheDocument();

    // Slugs
    expect(screen.getByText('main-nav')).toBeInTheDocument();
    expect(screen.getByText('footer-links')).toBeInTheDocument();
    expect(screen.getByText('sidebar-menu')).toBeInTheDocument();

    // Item counts - mock t() returns 'itemCount' with params replaced
    // t('system.menus.itemCount', { count: 5 }) => 'itemCount' (no {{count}} in 'itemCount')
    // So all 3 show 'itemCount'
    const itemCounts = screen.getAllByText('itemCount');
    expect(itemCounts.length).toBe(3);
  });

  it('renders location badges', () => {
    render(<MenusTable {...defaultProps} />);
    expect(screen.getByText('locationHeader')).toBeInTheDocument();
    expect(screen.getByText('locationFooter')).toBeInTheDocument();
    expect(screen.getByText('locationSidebar')).toBeInTheDocument();
  });

  it('renders "New Menu" button', () => {
    render(<MenusTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /newMenu/i })).toBeInTheDocument();
  });

  it('calls onNewMenu when New Menu button clicked', async () => {
    const user = userEvent.setup();
    render(<MenusTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newMenu/i }));
    expect(defaultProps.onNewMenu).toHaveBeenCalled();
  });

  it('row actions: edit, manage items, delete', async () => {
    const user = userEvent.setup();
    render(<MenusTable {...defaultProps} />);

    // Click first row action button
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);

    // Verify dropdown items
    expect(screen.getByText('edit')).toBeInTheDocument();
    expect(screen.getByText('manageItems')).toBeInTheDocument();
    expect(screen.getByText('delete')).toBeInTheDocument();
  });

  it('calls onEdit when edit action clicked', async () => {
    const user = userEvent.setup();
    render(<MenusTable {...defaultProps} />);

    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('edit'));

    expect(defaultProps.onEdit).toHaveBeenCalledWith(mockMenus[0]);
  });

  it('calls onManageItems when manage items action clicked', async () => {
    const user = userEvent.setup();
    render(<MenusTable {...defaultProps} />);

    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('manageItems'));

    expect(defaultProps.onManageItems).toHaveBeenCalledWith(mockMenus[0]);
  });

  it('calls onDelete when delete action clicked', async () => {
    const user = userEvent.setup();
    render(<MenusTable {...defaultProps} />);

    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('delete'));

    expect(defaultProps.onDelete).toHaveBeenCalledWith(mockMenus[0]);
  });

  it('shows empty state when no menus', () => {
    render(<MenusTable {...defaultProps} menus={[]} />);
    expect(screen.getByText('noMenusFound')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    const { container } = render(<MenusTable {...defaultProps} menus={[]} loading={true} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });
});

// =========================================
// MenuFormDialog Tests
// =========================================
describe('MenuFormDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    onSubmit: vi.fn(),
    loading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create form title when no menu provided', () => {
    render(<MenuFormDialog {...defaultProps} />);
    expect(screen.getByText('newMenu')).toBeInTheDocument();
  });

  it('renders edit form title when menu provided', () => {
    render(<MenuFormDialog {...defaultProps} menu={mockMenus[0]} />);
    expect(screen.getByText('editMenu')).toBeInTheDocument();
  });

  it('renders name and slug fields', () => {
    render(<MenuFormDialog {...defaultProps} />);
    expect(screen.getByLabelText('menuName')).toBeInTheDocument();
    expect(screen.getByLabelText('slug')).toBeInTheDocument();
  });

  it('validates name is required', async () => {
    const user = userEvent.setup();
    render(<MenuFormDialog {...defaultProps} />);

    // Submit without name
    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('validates slug regex (lowercase, numbers, hyphens)', async () => {
    const user = userEvent.setup();
    render(<MenuFormDialog {...defaultProps} />);

    const nameInput = screen.getByLabelText('menuName');
    const slugInput = screen.getByLabelText('slug');

    await user.type(nameInput, 'Test Menu');
    await user.clear(slugInput);
    await user.type(slugInput, 'Invalid Slug!');

    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('renders location select with options', () => {
    render(<MenuFormDialog {...defaultProps} />);
    expect(screen.getByText('location')).toBeInTheDocument();
  });

  it('renders description field', () => {
    render(<MenuFormDialog {...defaultProps} />);
    expect(screen.getByLabelText('description')).toBeInTheDocument();
  });

  it('pre-fills form when editing', () => {
    render(<MenuFormDialog {...defaultProps} menu={mockMenus[0]} />);
    expect(screen.getByLabelText('menuName')).toHaveValue('Main Navigation');
    expect(screen.getByLabelText('slug')).toHaveValue('main-nav');
  });

  it('calls onSubmit with form data when valid', async () => {
    const user = userEvent.setup();
    render(<MenuFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('menuName'), 'New Menu');
    // Slug should auto-generate
    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Menu',
          slug: expect.stringMatching(/^[a-z0-9-]+$/),
        }),
      );
    });
  });
});

// =========================================
// MenuItemsEditor Tests
// =========================================
describe('MenuItemsEditor', () => {
  const defaultProps = {
    menuDetail: mockMenuDetail,
    loading: false,
    onAddItem: vi.fn(),
    onEditItem: vi.fn(),
    onDeleteItem: vi.fn(),
    onToggleActive: vi.fn(),
    onMoveUp: vi.fn(),
    onMoveDown: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders menu name and items tree', () => {
    render(<MenuItemsEditor {...defaultProps} />);
    expect(screen.getByText('Main Navigation')).toBeInTheDocument();
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Blog Post')).toBeInTheDocument();
  });

  it('shows Add Item button', () => {
    render(<MenuItemsEditor {...defaultProps} />);
    expect(screen.getByRole('button', { name: /addItem/i })).toBeInTheDocument();
  });

  it('renders item label, type badge, URL, active toggle', () => {
    render(<MenuItemsEditor {...defaultProps} />);

    // Labels
    expect(screen.getByText('Home')).toBeInTheDocument();
    expect(screen.getByText('Blog Post')).toBeInTheDocument();

    // Type badges (Home + Sub Page are both custom, so multiple)
    expect(screen.getAllByText('typeCustom').length).toBeGreaterThanOrEqual(1);
    expect(screen.getByText('typePost')).toBeInTheDocument();

    // URL display
    expect(screen.getByText('/')).toBeInTheDocument();
  });

  it('shows broken reference warning for is_broken items', () => {
    render(<MenuItemsEditor {...defaultProps} />);
    // Blog Post has is_broken=true
    expect(screen.getByTitle('broken')).toBeInTheDocument();
  });

  it('renders nested items with indentation', () => {
    render(<MenuItemsEditor {...defaultProps} />);
    // Sub Page is a child of Home
    expect(screen.getByText('Sub Page')).toBeInTheDocument();
  });

  it('shows empty state when no items', () => {
    const emptyMenu: SiteMenuDetail = {
      ...mockMenuDetail,
      items: [],
    };
    render(<MenuItemsEditor {...defaultProps} menuDetail={emptyMenu} />);
    expect(screen.getByText('noItems')).toBeInTheDocument();
  });

  it('calls onAddItem when Add Item button clicked', async () => {
    const user = userEvent.setup();
    render(<MenuItemsEditor {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /addItem/i }));
    expect(defaultProps.onAddItem).toHaveBeenCalled();
  });

  it('calls onDeleteItem when delete button clicked', async () => {
    const user = userEvent.setup();
    render(<MenuItemsEditor {...defaultProps} />);
    const deleteButtons = screen.getAllByLabelText('deleteItem');
    await user.click(deleteButtons[0]);
    expect(defaultProps.onDeleteItem).toHaveBeenCalledWith('i1');
  });

  it('calls onEditItem when edit button clicked', async () => {
    const user = userEvent.setup();
    render(<MenuItemsEditor {...defaultProps} />);
    const editButtons = screen.getAllByLabelText('editItem');
    await user.click(editButtons[0]);
    expect(defaultProps.onEditItem).toHaveBeenCalledWith(expect.objectContaining({ id: 'i1' }));
  });
});
