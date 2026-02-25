import { describe, it, expect, vi, beforeEach } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { PermissionTree } from '../PermissionTree';
import type { TreeNode } from '../PermissionTree';

const mockItems: TreeNode[] = [
  {
    id: '1',
    label: 'Posts',
    children: [
      { id: '1-1', label: 'Create Post', children: [] },
      { id: '1-2', label: 'Edit Post', children: [] },
      { id: '1-3', label: 'Delete Post', children: [] },
    ],
  },
  {
    id: '2',
    label: 'Users',
    children: [
      { id: '2-1', label: 'Create User', children: [] },
      { id: '2-2', label: 'Edit User', children: [] },
    ],
  },
  { id: '3', label: 'Settings', children: [] },
];

describe('PermissionTree', () => {
  const defaultProps = {
    items: mockItems,
    checkedIds: [] as string[],
    onChange: vi.fn(),
  };

  beforeEach(() => {
    vi.clearAllMocks();
  });

  it('renders all top-level items', () => {
    render(<PermissionTree {...defaultProps} />);
    expect(screen.getByText('Posts')).toBeInTheDocument();
    expect(screen.getByText('Users')).toBeInTheDocument();
    expect(screen.getByText('Settings')).toBeInTheDocument();
  });

  it('renders child items', () => {
    render(<PermissionTree {...defaultProps} />);
    expect(screen.getByText('Create Post')).toBeInTheDocument();
    expect(screen.getByText('Edit Post')).toBeInTheDocument();
  });

  it('checks items that are in checkedIds', () => {
    render(<PermissionTree {...defaultProps} checkedIds={['1-1', '3']} />);
    const checkboxes = screen.getAllByRole('checkbox');
    const createPostCb = checkboxes.find(
      (cb) => cb.closest('label')?.textContent?.includes('Create Post'),
    );
    expect(createPostCb).toBeTruthy();
  });

  it('calls onChange when a leaf node is toggled', async () => {
    const user = userEvent.setup();
    render(<PermissionTree {...defaultProps} checkedIds={[]} />);
    const settingsCb = screen.getByText('Settings').closest('label')?.querySelector('button') ||
      screen.getByText('Settings').parentElement?.querySelector('[role="checkbox"]');
    if (settingsCb) {
      await user.click(settingsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith(['3']);
    }
  });

  it('selects all children when parent is checked', async () => {
    const user = userEvent.setup();
    render(<PermissionTree {...defaultProps} checkedIds={[]} />);
    const postsCb = screen.getByText('Posts').parentElement?.querySelector('[role="checkbox"]');
    if (postsCb) {
      await user.click(postsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith(
        expect.arrayContaining(['1', '1-1', '1-2', '1-3']),
      );
    }
  });

  it('deselects all children when parent is unchecked', async () => {
    const user = userEvent.setup();
    render(
      <PermissionTree {...defaultProps} checkedIds={['1', '1-1', '1-2', '1-3']} />,
    );
    const postsCb = screen.getByText('Posts').parentElement?.querySelector('[role="checkbox"]');
    if (postsCb) {
      await user.click(postsCb);
      expect(defaultProps.onChange).toHaveBeenCalledWith([]);
    }
  });

  it('renders empty message when no items', () => {
    render(<PermissionTree {...defaultProps} items={[]} />);
    expect(screen.getByText('No permissions available')).toBeInTheDocument();
  });
});
