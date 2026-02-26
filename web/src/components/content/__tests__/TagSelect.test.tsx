import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen, waitFor } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { TagSelect } from '../TagSelect';

const mockSuggest = vi.fn();
const mockCreate = vi.fn();

vi.mock('@/lib/content-api', () => ({
  tagsApi: {
    suggest: (...args: unknown[]) => mockSuggest(...args),
    create: (...args: unknown[]) => mockCreate(...args),
  },
}));

const sampleTags = [
  { id: 'tag-1', name: 'JavaScript', slug: 'javascript', post_count: 15, created_at: '2026-01-01T00:00:00Z' },
  { id: 'tag-2', name: 'React', slug: 'react', post_count: 10, created_at: '2026-01-01T00:00:00Z' },
  { id: 'tag-3', name: 'TypeScript', slug: 'typescript', post_count: 8, created_at: '2026-01-01T00:00:00Z' },
];

describe('TagSelect', () => {
  beforeEach(() => {
    vi.clearAllMocks();
    mockSuggest.mockResolvedValue({ success: true, data: sampleTags });
    mockCreate.mockResolvedValue({ success: true, data: { id: 'tag-new', name: 'NewTag', slug: 'newtag', post_count: 0, created_at: '2026-01-01T00:00:00Z' } });
  });

  it('renders add tag button', () => {
    render(<TagSelect value={[]} onChange={vi.fn()} allTags={[]} />);
    expect(screen.getByRole('button', { name: /add tag/i })).toBeInTheDocument();
  });

  it('shows selected tags as Badge components', () => {
    render(
      <TagSelect
        value={['tag-1', 'tag-2']}
        onChange={vi.fn()}
        allTags={sampleTags}
      />,
    );
    expect(screen.getByText('JavaScript')).toBeInTheDocument();
    expect(screen.getByText('React')).toBeInTheDocument();
  });

  it('can remove a tag by clicking X button', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(
      <TagSelect
        value={['tag-1', 'tag-2']}
        onChange={onChange}
        allTags={sampleTags}
      />,
    );

    const removeButtons = screen.getAllByRole('button', { name: /remove/i });
    await user.click(removeButtons[0]);
    expect(onChange).toHaveBeenCalledWith(['tag-2']);
  });

  it('opens popover and shows suggestions on search', async () => {
    const user = userEvent.setup();
    render(<TagSelect value={[]} onChange={vi.fn()} allTags={[]} />);

    await user.click(screen.getByRole('button', { name: /add tag/i }));

    await waitFor(() => {
      const input = screen.getByPlaceholderText(/search tags/i);
      expect(input).toBeInTheDocument();
    });
  });

  it('calls onChange when a suggested tag is selected', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    render(<TagSelect value={[]} onChange={onChange} allTags={[]} />);

    await user.click(screen.getByRole('button', { name: /add tag/i }));

    const input = screen.getByPlaceholderText(/search tags/i);
    await user.type(input, 'Java');

    await waitFor(() => {
      expect(mockSuggest).toHaveBeenCalled();
    });

    await waitFor(() => {
      expect(screen.getByText('JavaScript')).toBeInTheDocument();
    });

    await user.click(screen.getByText('JavaScript'));
    expect(onChange).toHaveBeenCalledWith(['tag-1']);
  });

  it('shows create option when search has no exact match', async () => {
    const user = userEvent.setup();
    mockSuggest.mockResolvedValue({ success: true, data: [] });

    render(<TagSelect value={[]} onChange={vi.fn()} allTags={[]} />);

    await user.click(screen.getByRole('button', { name: /add tag/i }));

    const input = screen.getByPlaceholderText(/search tags/i);
    await user.type(input, 'new-tag');

    await waitFor(() => {
      expect(screen.getByText(/create "new-tag"/i)).toBeInTheDocument();
    });
  });

  it('creates a new tag and selects it', async () => {
    const user = userEvent.setup();
    const onChange = vi.fn();
    mockSuggest.mockResolvedValue({ success: true, data: [] });

    render(<TagSelect value={[]} onChange={onChange} allTags={[]} />);

    await user.click(screen.getByRole('button', { name: /add tag/i }));

    const input = screen.getByPlaceholderText(/search tags/i);
    await user.type(input, 'NewTag');

    await waitFor(() => {
      expect(screen.getByText(/create "NewTag"/i)).toBeInTheDocument();
    });

    await user.click(screen.getByText(/create "NewTag"/i));

    await waitFor(() => {
      expect(mockCreate).toHaveBeenCalledWith({ name: 'NewTag' });
      expect(onChange).toHaveBeenCalledWith(['tag-new']);
    });
  });
});
