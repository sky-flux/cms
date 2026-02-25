import { describe, it, expect, vi } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TagForm } from '../TagForm';
import type { Tag } from '@/lib/content-api';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string) => {
      const map: Record<string, string> = {
        'content.addTag': 'Add Tag',
        'content.editTag': 'Edit Tag',
        'content.tagName': 'Name',
        'content.tagSlug': 'Slug',
        'common.save': 'Save',
        'common.cancel': 'Cancel',
        'common.loading': 'Loading...',
      };
      return map[key] ?? key;
    },
  }),
}));

const existingTag: Tag = {
  id: 'tag-1',
  name: 'JavaScript',
  slug: 'javascript',
  post_count: 15,
  created_at: '2026-01-15T00:00:00Z',
};

describe('TagForm', () => {
  it('renders create mode with empty fields', () => {
    render(
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );
    expect(screen.getByText('Add Tag')).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toHaveValue('');
    expect(screen.getByLabelText('Slug')).toHaveValue('');
  });

  it('renders edit mode with pre-filled fields', () => {
    render(
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
        tag={existingTag}
      />,
    );
    expect(screen.getByText('Edit Tag')).toBeInTheDocument();
    expect(screen.getByLabelText('Name')).toHaveValue('JavaScript');
    expect(screen.getByLabelText('Slug')).toHaveValue('javascript');
  });

  it('auto-generates slug from name in create mode', async () => {
    const user = userEvent.setup();
    render(
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );

    await user.type(screen.getByLabelText('Name'), 'Hello World');

    await waitFor(() => {
      expect(screen.getByLabelText('Slug')).toHaveValue('hello-world');
    });
  });

  it('does not render when closed', () => {
    render(
      <TagForm
        open={false}
        onOpenChange={vi.fn()}
        onSubmit={vi.fn()}
      />,
    );
    expect(screen.queryByText('Add Tag')).not.toBeInTheDocument();
  });

  it('shows validation error when name is empty on submit', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn();
    render(
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
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
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
      />,
    );

    await user.type(screen.getByLabelText('Name'), 'New Tag');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'New Tag',
          slug: 'new-tag',
        }),
      );
    });
  });

  it('calls onSubmit with form data in edit mode', async () => {
    const user = userEvent.setup();
    const onSubmit = vi.fn().mockResolvedValue(undefined);
    render(
      <TagForm
        open={true}
        onOpenChange={vi.fn()}
        onSubmit={onSubmit}
        tag={existingTag}
      />,
    );

    const nameInput = screen.getByLabelText('Name');
    await user.clear(nameInput);
    await user.type(nameInput, 'Updated Tag');
    await user.click(screen.getByRole('button', { name: 'Save' }));

    await waitFor(() => {
      expect(onSubmit).toHaveBeenCalledWith(
        expect.objectContaining({
          name: 'Updated Tag',
        }),
      );
    });
  });
});
