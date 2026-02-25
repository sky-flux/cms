import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CategoryForm } from '../CategoryForm';
import type { CategoryNode } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const map: Record<string, string> = {
        'content.addCategory': 'Add Category',
        'content.editCategory': 'Edit Category',
        'content.categoryName': 'Name',
        'content.categorySlug': 'Slug',
        'content.categoryDescription': 'Description',
        'content.categoryParent': 'Parent Category',
        'content.categoryNone': 'None (Root)',
        'common.save': 'Save',
        'common.cancel': 'Cancel',
        'common.loading': 'Loading...',
      };
      return map[key] ?? key;
    },
  }),
}));

const parentOptions: { id: string; name: string; depth: number }[] = [
  { id: 'cat-1', name: 'Technology', depth: 0 },
  { id: 'cat-1-1', name: 'Frontend', depth: 1 },
  { id: 'cat-2', name: 'Design', depth: 0 },
];

const existingCategory: CategoryNode = {
  id: 'cat-1',
  name: 'Technology',
  slug: 'technology',
  path: '/technology',
  description: 'Tech articles',
  parent_id: null,
  post_count: 5,
  sort_order: 1,
  children: [],
};

describe('CategoryForm', () => {
  it('renders create mode with empty fields', () => {
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
      />,
    );
    expect(screen.getByText('Add Category')).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toHaveValue('');
    expect(screen.getByLabelText('Slug')).toHaveValue('');
  });

  it('renders edit mode with pre-filled fields', () => {
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
        category={existingCategory}
      />,
    );
    expect(screen.getByText('Edit Category')).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toHaveValue('Technology');
    expect(screen.getByLabelText('Slug')).toHaveValue('technology');
  });

  it('auto-generates slug from name in create mode', async () => {
    const user = userEvent.setup();
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
      />,
    );

    const nameInput = screen.getByLabelText('Name');
    await user.type(nameInput, 'Hello World');

    await waitFor(() => {
      expect(screen.getByLabelText('Slug')).toHaveValue('hello-world');
    });
  });

  it('does not render when closed', () => {
    render(
      <CategoryForm
        open={false}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
      />,
    );
    expect(screen.queryByText('Add Category')).not.toBeInTheDocument();
  });

  it('shows validation error when name is empty on submit', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        parentOptions={parentOptions}
      />,
    );

    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSubmit).not.toHaveBeenCalled();
    });
  });

  it('calls onSubmit with form data in create mode', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        parentOptions={parentOptions}
      />,
    );

    await user.type(screen.getByLabelText('Name'), 'New Category');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Category',
          slug: 'new-category',
        }),
      );
    });
  });

  it('calls onSubmit with form data in edit mode', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        parentOptions={parentOptions}
        category={existingCategory}
      />,
    );

    const nameInput = screen.getByLabelText('Name');
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Name');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Updated Name',
        }),
      );
    });
  });

  it('allows editing slug manually', async () => {
    const user = userEvent.setup();
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
      />,
    );

    const slugInput = screen.getByLabelText('Slug');
    await user.type(slugInput, 'custom-slug');
    expect(slugInput).toHaveValue('custom-slug');
  });

  it('shows description textarea', () => {
    render(
      <CategoryForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        parentOptions={parentOptions}
      />,
    );
    expect(screen.getByLabelText('Description')).toBeInTheDocument();
  });
});
