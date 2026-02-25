import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { DataTable } from '../DataTable';
import type { ColumnDef } from '@tanstack/react-table';

interface TestRow {
  id: string;
  name: string;
}

const columns: ColumnDef<TestRow>[] = [
  { accessorKey: 'name', header: 'Name' },
];

const data: TestRow[] = [
  { id: '1', name: 'Alpha' },
  { id: '2', name: 'Beta' },
];

describe('DataTable', () => {
  it('renders column headers', () => {
    render(<DataTable columns={columns} data={data} />);
    expect(screen.getByText('Name')).toBeInTheDocument();
  });

  it('renders row data', () => {
    render(<DataTable columns={columns} data={data} />);
    expect(screen.getByText('Alpha')).toBeInTheDocument();
    expect(screen.getByText('Beta')).toBeInTheDocument();
  });

  it('shows empty message when no data', () => {
    render(<DataTable columns={columns} data={[]} emptyMessage="No items" />);
    expect(screen.getByText('No items')).toBeInTheDocument();
  });

  it('shows loading skeleton', () => {
    const { container } = render(<DataTable columns={columns} data={[]} loading={true} />);
    expect(container.querySelectorAll('[data-slot="skeleton"]').length).toBeGreaterThan(0);
  });

  it('renders pagination when provided', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByText('1 / 3')).toBeInTheDocument();
  });

  it('calls onPageChange when next page clicked', async () => {
    const onPageChange = vi.fn();
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={onPageChange}
      />,
    );
    await userEvent.click(screen.getByRole('button', { name: /next/i }));
    expect(onPageChange).toHaveBeenCalledWith(2);
  });

  it('disables prev button on first page', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 1, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: /prev/i })).toBeDisabled();
  });

  it('disables next button on last page', () => {
    render(
      <DataTable
        columns={columns}
        data={data}
        pagination={{ page: 3, totalPages: 3 }}
        onPageChange={() => {}}
      />,
    );
    expect(screen.getByRole('button', { name: /next/i })).toBeDisabled();
  });
});
