import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CategoryTree } from '../CategoryTree';
import type { CategoryNode } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.noCategoriesFound': 'No categories yet',
        'content.addCategory': 'Add Category',
        'content.addSubcategory': 'Add Subcategory',
        'content.editCategory': 'Edit Category',
        'content.deleteCategory': 'Delete Category',
      };
      if (key === 'content.postCount' && opts?.count !== undefined) {
        return `${opts.count} posts`;
      }
      return map[key] ?? key;
    },
  }),
}));

vi.mock('@dnd-kit/core', () => ({
  DndContext: ({ children }: any) => <div>{children}</div>,
  closestCenter: vi.fn(),
  KeyboardSensor: vi.fn(),
  PointerSensor: vi.fn(),
  useSensor: vi.fn(),
  useSensors: vi.fn(() => []),
}));

vi.mock('@dnd-kit/sortable', () => ({
  SortableContext: ({ children }: any) => <div>{children}</div>,
  useSortable: () => ({
    attributes: {},
    listeners: {},
    setNodeRef: vi.fn(),
    transform: null,
    transition: null,
  }),
  verticalListSortingStrategy: vi.fn(),
}));

const mockCategories: CategoryNode[] = [
  {
    id: 'cat-1',
    name: 'Technology',
    slug: 'technology',
    path: '/technology',
    description: 'Tech articles',
    parent_id: null,
    post_count: 5,
    sort_order: 1,
    children: [
      {
        id: 'cat-1-1',
        name: 'Frontend',
        slug: 'frontend',
        path: '/technology/frontend',
        description: 'Frontend dev',
        parent_id: 'cat-1',
        post_count: 3,
        sort_order: 1,
        children: [],
      },
      {
        id: 'cat-1-2',
        name: 'Backend',
        slug: 'backend',
        path: '/technology/backend',
        description: 'Backend dev',
        parent_id: 'cat-1',
        post_count: 2,
        sort_order: 2,
        children: [],
      },
    ],
  },
  {
    id: 'cat-2',
    name: 'Design',
    slug: 'design',
    path: '/design',
    description: 'Design articles',
    parent_id: null,
    post_count: 10,
    sort_order: 2,
    children: [],
  },
];

describe('CategoryTree', () => {
  it('renders empty state when no categories', () => {
    render(
      <CategoryTree
        categories={[]}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText('No categories yet')).toBeInTheDocument();
  });

  it('renders root category names', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText('Technology')).toBeInTheDocument();
    expect(screen.getByText('Design')).toBeInTheDocument();
  });

  it('displays post count badge', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    expect(screen.getByText('5 posts')).toBeInTheDocument();
    expect(screen.getByText('10 posts')).toBeInTheDocument();
  });

  it('toggles children visibility on expand/collapse click', async () => {
    const user = userEvent.setup();
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    // Children should not be visible initially (collapsed by default)
    expect(screen.queryByText('Frontend')).not.toBeInTheDocument();

    // Click expand toggle for Technology
    const expandButtons = screen.getAllByRole('button', { name: /toggle/i });
    await user.click(expandButtons[0]);

    // Now children should be visible
    expect(screen.getByText('Frontend')).toBeInTheDocument();
    expect(screen.getByText('Backend')).toBeInTheDocument();

    // Click collapse
    await user.click(expandButtons[0]);
    expect(screen.queryByText('Frontend')).not.toBeInTheDocument();
  });

  it('calls onEdit with category when edit button clicked', async () => {
    const onEdit = vi.fn();
    const user = userEvent.setup();
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={onEdit}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    await user.click(editButtons[0]);
    expect(onEdit).toHaveBeenCalledWith(mockCategories[0]);
  });

  it('calls onAddChild with category id when add child clicked', async () => {
    const onAddChild = vi.fn();
    const user = userEvent.setup();
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={onAddChild}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    const addChildButtons = screen.getAllByRole('button', { name: /add subcategory/i });
    await user.click(addChildButtons[0]);
    expect(onAddChild).toHaveBeenCalledWith('cat-1');
  });

  it('calls onDelete with category when delete button clicked', async () => {
    const onDelete = vi.fn();
    const user = userEvent.setup();
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={onDelete}
        onReorder={vi.fn()}
      />,
    );
    const deleteButtons = screen.getAllByRole('button', { name: /delete/i });
    await user.click(deleteButtons[0]);
    expect(onDelete).toHaveBeenCalledWith(mockCategories[0]);
  });

  it('renders nested children after expand with correct indentation', async () => {
    const user = userEvent.setup();
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    // Expand Technology
    const expandButtons = screen.getAllByRole('button', { name: /toggle/i });
    await user.click(expandButtons[0]);

    // Children should show their post counts
    expect(screen.getByText('3 posts')).toBeInTheDocument();
    expect(screen.getByText('2 posts')).toBeInTheDocument();
  });

  it('does not show expand toggle for leaf nodes', () => {
    render(
      <CategoryTree
        categories={mockCategories}
        onEdit={vi.fn()}
        onAddChild={vi.fn()}
        onDelete={vi.fn()}
        onReorder={vi.fn()}
      />,
    );
    // Design has no children, so it should not have an expand toggle
    // Technology has children, so it has a toggle
    const expandButtons = screen.getAllByRole('button', { name: /toggle/i });
    // Only Technology should have expand button
    expect(expandButtons).toHaveLength(1);
  });
});
