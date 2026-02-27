import { Pencil, Trash, Eye } from 'lucide-react';
import type { Site } from '../types/sites';

interface SitesTableProps {
  sites: Site[];
  onEdit: (siteId: string) => void;
  onDelete: (siteId: string) => void;
  onView: (siteId: string) => void;
}

export function SitesTable({ sites, onEdit, onDelete, onView }: SitesTableProps) {
  if (sites.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No sites found.
      </div>
    );
  }

  const getStatusBadgeClass = (status: Site['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead className="bg-muted/50">
          <tr>
            <th className="px-4 py-3 text-left text-sm font-medium">Name</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Slug</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Domain</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
            <th className="px-4 py-3 text-right text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {sites.map((site) => (
            <tr key={site.id} className="border-t hover:bg-muted/50">
              <td className="px-4 py-3">
                <span className="font-medium">{site.name}</span>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {site.slug}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {site.domain || '-'}
              </td>
              <td className="px-4 py-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(site.status)}`}>
                  {site.status}
                </span>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(site.createdAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  <button
                    onClick={() => onView(site.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="View"
                    aria-label="View"
                  >
                    <Eye className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onEdit(site.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="Edit"
                    aria-label="Edit"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onDelete(site.id)}
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
