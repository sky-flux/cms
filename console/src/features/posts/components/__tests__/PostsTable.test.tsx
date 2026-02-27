import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import { PostsTable } from '../PostsTable';
import type { Post } from '../../types/posts';

const mockPosts: Post[] = [
  {
    id: '1',
    title: 'Test Post 1',
    slug: 'test-post-1',
    content: 'Content 1',
    excerpt: 'Excerpt 1',
    status: 'published',
    authorId: 'author-1',
    siteId: 'site-1',
    publishedAt: '2026-01-01T00:00:00Z',
    createdAt: '2026-01-01T00:00:00Z',
    updatedAt: '2026-01-01T00:00:00Z',
  },
  {
    id: '2',
    title: 'Test Post 2',
    slug: 'test-post-2',
    content: 'Content 2',
    status: 'draft',
    authorId: 'author-1',
    siteId: 'site-1',
    createdAt: '2026-01-02T00:00:00Z',
    updatedAt: '2026-01-02T00:00:00Z',
  },
];

describe('PostsTable', () => {
  const mockOnEdit = vi.fn();
  const mockOnDelete = vi.fn();
  const mockOnView = vi.fn();

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders posts table with data', () => {
    render(
      <PostsTable
        posts={mockPosts}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    expect(screen.getByText('Test Post 1')).toBeInTheDocument();
    expect(screen.getByText('Test Post 2')).toBeInTheDocument();
  });

  it('displays correct status badges', () => {
    render(
      <PostsTable
        posts={mockPosts}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    expect(screen.getByText('published')).toBeInTheDocument();
    expect(screen.getByText('draft')).toBeInTheDocument();
  });

  it('calls onEdit when edit button is clicked', () => {
    render(
      <PostsTable
        posts={mockPosts}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    const editButtons = screen.getAllByRole('button', { name: /edit/i });
    editButtons[0].click();

    expect(mockOnEdit).toHaveBeenCalledWith('1');
  });

  it('calls onDelete when delete button is clicked', () => {
    render(
      <PostsTable
        posts={mockPosts}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    const deleteButtons = screen.getAllByRole('button', { name: /delete/i });
    deleteButtons[0].click();

    expect(mockOnDelete).toHaveBeenCalledWith('1');
  });

  it('calls onView when view button is clicked', () => {
    render(
      <PostsTable
        posts={mockPosts}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    const viewButtons = screen.getAllByRole('button', { name: /view/i });
    viewButtons[0].click();

    expect(mockOnView).toHaveBeenCalledWith('1');
  });

  it('renders empty state when no posts', () => {
    render(
      <PostsTable
        posts={[]}
        siteSlug="test-site"
        onEdit={mockOnEdit}
        onDelete={mockOnDelete}
        onView={mockOnView}
      />
    );

    expect(screen.getByText(/no posts found/i)).toBeInTheDocument();
  });
});
