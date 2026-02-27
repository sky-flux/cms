import { MoreHorizontal, Pencil, Trash, Eye } from 'lucide-react';
import type { Post } from '../../types/posts';

interface PostsTableProps {
  posts: Post[];
  siteSlug: string;
  onEdit: (postId: string) => void;
  onDelete: (postId: string) => void;
  onView: (postId: string) => void;
}

export function PostsTable({ posts, siteSlug, onEdit, onDelete, onView }: PostsTableProps) {
  if (posts.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No posts found.
      </div>
    );
  }

  const getStatusBadgeClass = (status: Post['status']) => {
    switch (status) {
      case 'published':
        return 'bg-green-100 text-green-800';
      case 'draft':
        return 'bg-gray-100 text-gray-800';
      case 'scheduled':
        return 'bg-yellow-100 text-yellow-800';
      case 'private':
        return 'bg-purple-100 text-purple-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead className="bg-muted/50">
          <tr>
            <th className="px-4 py-3 text-left text-sm font-medium">Title</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Slug</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Updated</th>
            <th className="px-4 py-3 text-right text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {posts.map((post) => (
            <tr key={post.id} className="border-t hover:bg-muted/50">
              <td className="px-4 py-3">
                <div className="font-medium">{post.title}</div>
                {post.excerpt && (
                  <div className="text-sm text-muted-foreground truncate max-w-xs">
                    {post.excerpt}
                  </div>
                )}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {post.slug}
              </td>
              <td className="px-4 py-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(post.status)}`}>
                  {post.status}
                </span>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(post.createdAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(post.updatedAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  <button
                    onClick={() => onView(post.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="View"
                    aria-label="View"
                  >
                    <Eye className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onEdit(post.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="Edit"
                    aria-label="Edit"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onDelete(post.id)}
                    className="p-2 hover:bg-muted rounded-md text-destructive"
                    title="Delete"
                    aria-label="Delete"
                  >
                    <Trash className="h-4 w-4" />
                  </button>
                </div>
              </td>
            </tr>
          ))}
        </tbody>
      </table>
    </div>
  );
}
