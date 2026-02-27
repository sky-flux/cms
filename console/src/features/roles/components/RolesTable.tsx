import { Pencil, Trash, Shield } from 'lucide-react';
import type { Role } from '../types/roles';

interface RolesTableProps {
  roles: Role[];
  onEdit: (roleId: string) => void;
  onDelete: (roleId: string) => void;
}

export function RolesTable({ roles, onEdit, onDelete }: RolesTableProps) {
  if (roles.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No roles found.
      </div>
    );
  }

  const getStatusBadgeClass = (status: number) => {
    switch (status) {
      case 1:
        return 'bg-green-100 text-green-800';
      case 2:
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getStatusLabel = (status: number) => {
    switch (status) {
      case 1:
        return 'Active';
      case 2:
        return 'Inactive';
      default:
        return 'Unknown';
    }
  };

  return (
    <div className="rounded-md border">
      <table className="w-full">
        <thead className="bg-muted/50">
          <tr>
            <th className="px-4 py-3 text-left text-sm font-medium">Name</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Slug</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Description</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Permissions</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
            <th className="px-4 py-3 text-right text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {roles.map((role) => (
            <tr key={role.id} className="border-t hover:bg-muted/50">
              <td className="px-4 py-3">
                <div className="flex items-center gap-3">
                  <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center">
                    <Shield className="h-4 w-4" />
                  </div>
                  <span className="font-medium">{role.name}</span>
                  {role.isSystem && (
                    <span className="inline-flex items-center px-2 py-0.5 rounded text-xs font-medium bg-purple-100 text-purple-800">
                      System
                    </span>
                  )}
                </div>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground font-mono">
                {role.slug}
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground max-w-[200px] truncate">
                {role.description || '-'}
              </td>
              <td className="px-4 py-3 text-sm">
                {role.permissions?.length || 0} permissions
              </td>
              <td className="px-4 py-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(role.status)}`}>
                  {getStatusLabel(role.status)}
                </span>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(role.createdAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  <button
                    onClick={() => onEdit(role.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="Edit"
                    aria-label="Edit"
 disabled={role.isSystem}
                  >
                                       <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onDelete(role.id)}
                    className="p-2 hover:bg-muted rounded-md text-destructive"
                    title="Delete"
                    aria-label="Delete"
                    disabled={role.isSystem}
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
