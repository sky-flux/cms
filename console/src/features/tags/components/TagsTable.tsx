import { Pencil, Trash } from 'lucide-react';
import type { Tag } from '../../types/tags';

interface TagsTableProps {
  tags: Tag[];
  onEdit: (tagId: string) => void;
  onDelete: (tagId: string) => void;
}

export function TagsTable({ tags, onEdit, onDelete }: TagsTableProps) {
  if (tags.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No tags found.
      </div>
    );
  }

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead className="bg-muted/50">
          <tr>
            <th className="px-4 py-3 text-left text-sm font-medium">Name</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Slug</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Updated</th>
            <th className="px-4 py-3 text-right text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {tags.map((tag) => (
            <tr key={tag.id} className="border-t hover:bg-muted/50">
              <td className="px-4 py-3">
                <div className="font-medium">{tag.name}</div>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {tag.slug}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(tag.createdAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(tag.updatedAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  <button
                    onClick={() => onEdit(tag.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="Edit"
                    aria-label="Edit"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onDelete(tag.id)}
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
