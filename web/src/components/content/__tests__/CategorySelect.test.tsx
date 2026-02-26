import { describe, it, expect, vi, beforeEach, beforeAll } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { CategorySelect } from '../CategorySelect';

// cmdk calls scrollIntoView which jsdom doesn't implement
beforeAll(() => {
  Element.prototype.scrollIntoView = vi.fn();
});

const mockTree = vi.fn();

vi.mock('@/lib/content-api', () => ({
  categoriesApi: {
    tree: (...args: unknown[]) => mockTree(...args),
  },
}));

const sampleCategories = [
  {
    id: 'cat-1',
    name: 'Technology',
    slug: 'technology',
    path: '/technology',
    parent_id: null,
    post_count: 5,
    sort_order: 1,
    children: [
      {
        id: 'cat-1-1',
        name: 'Frontend',
        slug: 'frontend',
        path: '/technology/frontend',
        parent_id: 'cat-1',
        post_count: 3,
        sort_order: 1,
        children: [],
      },
    ],
  },
  {
    id: 'cat-2',
    name: 'Design',
    slug: 'design',
    path: '/design',
    parent_id: null,
    post_count: 2,
    sort_order: 2,
    children: [],
  },
];

describe('CategorySelect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockTree.mockResolvedValue({ success: true, data: sampleCategories });
  });

  it('renders with placeholder text', async () => {
    render(<CategorySelect value={[]} onChange={vi.fn()} />);
    await waitFor(() => {
      expect(screen.getByRole('button', { name: /select categories/i })).toBeInTheDocument();
    });
  });

  it('shows selected category names when value has IDs', async () => {
    render(<CategorySelect value={['cat-1', 'cat-2']} onChange={vi.fn()} />);
    await waitFor(() => {
      expect(screen.getByText(/Technology/)).toBeInTheDocument();
      expect(screen.getByText(/Design/)).toBeInTheDocument();
    });
  });

  it('opens popover and shows categories on click', async () => {
    const user = userEvent.setup();
    render(<CategorySelect value={[]} onChange={vi.fn()} />);

    await waitFor(() => {
      expect(mockTree).toHaveBeenCalled();
    });

    const trigger = screen.getByRole('button', { name: /select categories/i });
    await user.click(trigger);

    await waitFor(() => {
      expect(screen.getByText('Technology')).toBeInTheDocument();
      expect(screen.getByText('Frontend')).toBeInTheDocument();
      expect(screen.getByText('Design')).toBeInTheDocument();
    });
  });

  it('calls onChange when a category is toggled on', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<CategorySelect value={[]} onChange={onChange} />);

    await waitFor(() => {
      expect(mockTree).toHaveBeenCalled();
    });

    const trigger = screen.getByRole('button', { name: /select categories/i });
    await user.click(trigger);

    await waitFor(() => {
      expect(screen.getByText('Technology')).toBeInTheDocument();
    });

    // Click on Technology to select it
    await user.click(screen.getByText('Technology'));
    expect(onChange).toHaveBeenCalledWith(['cat-1']);
  });

  it('calls onChange to remove a category when toggled off', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<CategorySelect value={['cat-1']} onChange={onChange} />);

    await waitFor(() => {
      expect(mockTree).toHaveBeenCalled();
    });

    const trigger = screen.getByRole('button', { name: /select categories/i });
    await user.click(trigger);

    await waitFor(() => {
      // Multiple "Technology" elements: trigger text + list item
      expect(screen.getAllByText('Technology').length).toBeGreaterThanOrEqual(2);
    });

    // Click the list item option (role="option") to deselect
    const option = screen.getByRole('option', { name: /technology/i });
    await user.click(option);
    expect(onChange).toHaveBeenCalledWith([]);
  });

  it('shows check icon for selected categories', async () => {
    const user = userEvent.setup();
    render(<CategorySelect value={['cat-1']} onChange={vi.fn()} />);

    await waitFor(() => {
      expect(mockTree).toHaveBeenCalled();
    });

    const trigger = screen.getByRole('button', { name: /select categories/i });
    await user.click(trigger);

    await waitFor(() => {
      const checkIcons = screen.getAllByTestId('category-check');
      expect(checkIcons.length).toBeGreaterThanOrEqual(1);
    });
  });
});
