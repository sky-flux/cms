import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, act } from '@testing-library/react';
import { CategoryTree } from '../../components/CategoryTree';
import type { Category } from '../../types/categories';

const mockCategories: Category[] = [
  {
    id: '1',
    name: 'Technology',
    slug: 'technology',
    description: 'Tech posts',
    siteId: 'site-1',
    order: 1,
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
    children: [
      {
        id: '2',
        name: 'Programming',
        slug: 'programming',
        description: 'Programming posts',
        parentId: '1',
        siteId: 'site-1',
        order: 1,
        createdAt: '2026-01-01T00:00:00Z',
        updatedAt: '2026-01-01T00:00:00Z',
      },
      {
        id: '3',
        name: 'AI',
        slug: 'ai',
        description: 'AI posts',
        parentId: '1',
        siteId: 'site-1',
        order: 2,
        createdAt: '2026-01-02T00:00:00Z',
        updatedAt: '2026-01-02T00:00:00Z',
      },
    ],
  },
  {
    id: '4',
    name: 'Design',
    slug: 'design',
    description: 'Design posts',
    siteId: 'site-1',
    order: 2,
    createdAt: '2026-01-03T00:00:00Z',
    updatedAt: '2026-01-03T00:00:00Z',
  },
];

describe('CategoryTree', () => {
  const mockOnEdit = vi.fn();
  const mockOnDelete = vi.fn();
  const mockOnSelect = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders categories tree with data', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    expect(screen.getByText('Technology')).toBeInTheDocument();
    expect(screen.getByText('Design')).toBeInTheDocument();
  });

  it('expands and shows children when toggle is clicked', async () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    // Initially, children should not be visible (collapsed)
    expect(screen.queryByText('Programming')).not.toBeInTheDocument();

    // Find and click the expand button for Technology
    const expandButtons = screen.getAllByRole('button', { name: /expand/i });
    await act(async () => {
      expandButtons[0].click();
    });

    // Now children should be visible
    expect(screen.getByText('Programming')).toBeInTheDocument();
    expect(screen.getByText('AI')).toBeInTheDocument();
  });

  it('calls onEdit when edit button is clicked', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    editButtons[0].click();

    expect(mockOnEdit).toHaveBeenCalledWith('1');
  });

  it('calls onDelete when delete button is clicked', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i });
    deleteButtons[0].click();

    expect(mockOnDelete).toHaveBeenCalledWith('1');
  });

  it('calls onSelect when category is clicked', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    const categoryNames = screen.getAllByText('Technology');
    categoryNames[0].click();

    expect(mockOnSelect).toHaveBeenCalledWith('1');
  });

  it('renders empty state when no categories', () => {
    render(
      <CategoryTree
        categories={[]}
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onSelect={mockOnSelect}
      />
    );

    expect(screen.getByText(/no categories found/i)).toBeInTheDocument();
  });
});
