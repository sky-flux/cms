import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { RedirectsTable } from '../RedirectsTable';
import { RedirectFormDialog } from '../RedirectFormDialog';
import { CsvImportDialog } from '../CsvImportDialog';
import type { Redirect, PaginationMeta } from '@/lib/system-api';

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
const mockRedirects: Redirect[] = [
  {
    id: 'r1',
    source_path: '/old-page',
    target_url: 'https://example.com/new-page',
    status_code: 301,
    is_active: true,
    hit_count: 42,
    last_hit_at: '2026-01-20T10:00:00Z',
    created_by: { id: 'u1', display_name: 'Alice' },
    created_at: '2026-01-10T10:00:00Z',
    updated_at: '2026-01-10T10:00:00Z',
  },
  {
    id: 'r2',
    source_path: '/temp-redirect',
    target_url: 'https://example.com/target',
    status_code: 302,
    is_active: false,
    hit_count: 0,
    last_hit_at: null,
    created_by: null,
    created_at: '2026-01-12T10:00:00Z',
    updated_at: '2026-01-12T10:00:00Z',
  },
  {
    id: 'r3',
    source_path: '/another-old',
    target_url: '/another-new',
    status_code: 301,
    is_active: true,
    hit_count: 100,
    last_hit_at: '2026-01-25T15:00:00Z',
    created_by: { id: 'u2', display_name: 'Bob' },
    created_at: '2026-01-15T10:00:00Z',
    updated_at: '2026-01-15T10:00:00Z',
  },
];

const mockPagination: PaginationMeta = {
  page: 1,
  per_page: 20,
  total: 3,
  total_pages: 1,
};

// =========================================
// RedirectsTable Tests
// =========================================
describe('RedirectsTable', () => {
  const defaultProps = {
    redirects: mockRedirects,
    pagination: mockPagination,
    loading: false,
    onPageChange: vi.fn(),
    onSearch: vi.fn(),
    onStatusCodeFilter: vi.fn(),
    onEdit: vi.fn(),
    onDelete: vi.fn(),
    onToggleActive: vi.fn(),
    onNewRedirect: vi.fn(),
    onImportCsv: vi.fn(),
    onExportCsv: vi.fn(),
    onBatchDelete: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders source path, target URL, status code, hit count columns', () => {
    render(<RedirectsTable {...defaultProps} />);

    // Source paths (monospace)
    expect(screen.getByText('/old-page')).toBeInTheDocument();
    expect(screen.getByText('/temp-redirect')).toBeInTheDocument();
    expect(screen.getByText('/another-old')).toBeInTheDocument();

    // Target URLs
    expect(screen.getByText('https://example.com/new-page')).toBeInTheDocument();
    expect(screen.getByText('https://example.com/target')).toBeInTheDocument();

    // Hit counts
    expect(screen.getByText('42')).toBeInTheDocument();
    expect(screen.getByText('0')).toBeInTheDocument();
    expect(screen.getByText('100')).toBeInTheDocument();
  });

  it('renders status code badges (301 and 302)', () => {
    render(<RedirectsTable {...defaultProps} />);
    const badges301 = screen.getAllByText('301');
    const badges302 = screen.getAllByText('302');
    expect(badges301.length).toBe(2); // r1 and r3
    expect(badges302.length).toBe(1); // r2
  });

  it('renders filter bar with status code select and search', () => {
    render(<RedirectsTable {...defaultProps} />);
    expect(screen.getByPlaceholderText('searchPlaceholder')).toBeInTheDocument();
  });

  it('renders New Redirect, Import CSV, Export CSV buttons', () => {
    render(<RedirectsTable {...defaultProps} />);
    expect(screen.getByRole('button', { name: /newRedirect/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /importCsv/i })).toBeInTheDocument();
    expect(screen.getByRole('button', { name: /exportCsv/i })).toBeInTheDocument();
  });

  it('calls onNewRedirect when New Redirect button clicked', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /newRedirect/i }));
    expect(defaultProps.onNewRedirect).toHaveBeenCalled();
  });

  it('calls onImportCsv when Import CSV button clicked', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /importCsv/i }));
    expect(defaultProps.onImportCsv).toHaveBeenCalled();
  });

  it('calls onExportCsv when Export CSV button clicked', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);
    await user.click(screen.getByRole('button', { name: /exportCsv/i }));
    expect(defaultProps.onExportCsv).toHaveBeenCalled();
  });

  it('checkbox selection enables batch delete bar', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);

    // Click first checkbox
    const checkboxes = screen.getAllByRole('checkbox');
    await user.click(checkboxes[1]); // First row checkbox (index 0 is header)

    // Batch delete button should appear
    await waitFor(() => {
      expect(screen.getByText(/selected/i)).toBeInTheDocument();
    });
  });

  it('row actions: edit and delete', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);

    // Click first row action button
    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);

    expect(screen.getByText('edit')).toBeInTheDocument();
    expect(screen.getByText('delete')).toBeInTheDocument();
  });

  it('calls onEdit when edit action clicked', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);

    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('edit'));

    expect(defaultProps.onEdit).toHaveBeenCalledWith(mockRedirects[0]);
  });

  it('calls onDelete when delete action clicked', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);

    const actionButtons = screen.getAllByRole('button', { name: /actions/i });
    await user.click(actionButtons[0]);
    await user.click(screen.getByText('delete'));

    expect(defaultProps.onDelete).toHaveBeenCalledWith(mockRedirects[0]);
  });

  it('shows empty state when no redirects', () => {
    render(<RedirectsTable {...defaultProps} redirects={[]} />);
    expect(screen.getByText('noRedirectsFound')).toBeInTheDocument();
  });

  it('shows loading state', () => {
    const { container } = render(<RedirectsTable {...defaultProps} redirects={[]} loading={true} />);
    const skeletons = container.querySelectorAll('[data-slot="skeleton"]');
    expect(skeletons.length).toBeGreaterThan(0);
  });

  it('calls onSearch when search input changes', async () => {
    const user = userEvent.setup();
    render(<RedirectsTable {...defaultProps} />);
    const searchInput = screen.getByPlaceholderText('searchPlaceholder');
    await user.type(searchInput, 'test');
    expect(defaultProps.onSearch).toHaveBeenCalledWith('test');
  });
});

// =========================================
// RedirectFormDialog Tests
// =========================================
describe('RedirectFormDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    onSubmit: vi.fn(),
    loading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders create form title', () => {
    render(<RedirectFormDialog {...defaultProps} />);
    expect(screen.getByText('newRedirect')).toBeInTheDocument();
  });

  it('renders edit form title when redirect provided', () => {
    render(<RedirectFormDialog {...defaultProps} redirect={mockRedirects[0]} />);
    expect(screen.getByText('editRedirect')).toBeInTheDocument();
  });

  it('renders source_path and target_url fields', () => {
    render(<RedirectFormDialog {...defaultProps} />);
    expect(screen.getByLabelText('sourcePath')).toBeInTheDocument();
    expect(screen.getByLabelText('targetUrl')).toBeInTheDocument();
  });

  it('validates source_path starts with "/"', async () => {
    const user = userEvent.setup();
    render(<RedirectFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('sourcePath'), 'no-slash');
    await user.type(screen.getByLabelText('targetUrl'), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('validates source_path has no "?"', async () => {
    const user = userEvent.setup();
    render(<RedirectFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('sourcePath'), '/path?query=1');
    await user.type(screen.getByLabelText('targetUrl'), 'https://example.com');
    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).not.toHaveBeenCalled();
    });
  });

  it('renders status_code select with 301 and 302 options', () => {
    render(<RedirectFormDialog {...defaultProps} />);
    expect(screen.getByText('statusCode')).toBeInTheDocument();
  });

  it('renders is_active switch', () => {
    render(<RedirectFormDialog {...defaultProps} />);
    expect(screen.getByText('active')).toBeInTheDocument();
  });

  it('pre-fills form when editing', () => {
    render(<RedirectFormDialog {...defaultProps} redirect={mockRedirects[0]} />);
    expect(screen.getByLabelText('sourcePath')).toHaveValue('/old-page');
    expect(screen.getByLabelText('targetUrl')).toHaveValue('https://example.com/new-page');
  });

  it('calls onSubmit with valid data', async () => {
    const user = userEvent.setup();
    render(<RedirectFormDialog {...defaultProps} />);

    await user.type(screen.getByLabelText('sourcePath'), '/old');
    await user.type(screen.getByLabelText('targetUrl'), '/new');
    await user.click(screen.getByRole('button', { name: /save/i }));

    await waitFor(() => {
      expect(defaultProps.onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          source_path: '/old',
          target_url: '/new',
        }),
      );
    });
  });
});

// =========================================
// CsvImportDialog Tests
// =========================================
describe('CsvImportDialog', () => {
  const defaultProps = {
    open: true,
    onOpenChange: vi.fn(),
    onImport: vi.fn(),
    loading: false,
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders file input', () => {
    render(<CsvImportDialog {...defaultProps} />);
    const fileInput = screen.getByLabelText('csvFile');
    expect(fileInput).toBeInTheDocument();
    expect(fileInput).toHaveAttribute('accept', '.csv');
  });

  it('renders Import button', () => {
    render(<CsvImportDialog {...defaultProps} />);
    expect(screen.getByRole('button', { name: /importCsv/i })).toBeInTheDocument();
  });

  it('shows CSV format help text', () => {
    render(<CsvImportDialog {...defaultProps} />);
    expect(screen.getByText('csvFormat')).toBeInTheDocument();
  });

  it('Import button is disabled when no file selected', () => {
    render(<CsvImportDialog {...defaultProps} />);
    expect(screen.getByRole('button', { name: /importCsv/i })).toBeDisabled();
  });

  it('calls onImport when file selected and Import clicked', async () => {
    const user = userEvent.setup();
    render(<CsvImportDialog {...defaultProps} />);

    const file = new File(
      ['source_path,target_url,status_code\n/old,/new,301'],
      'redirects.csv',
      { type: 'text/csv' },
    );

    const fileInput = screen.getByLabelText('csvFile');
    await user.upload(fileInput as HTMLInputElement, file);

    await waitFor(() => {
      expect(screen.getByRole('button', { name: /importCsv/i })).not.toBeDisabled();
    });

    await user.click(screen.getByRole('button', { name: /importCsv/i }));

    await waitFor(() => {
      expect(defaultProps.onImport).toHaveBeenCalledWith(expect.any(File));
    });
  });
});
