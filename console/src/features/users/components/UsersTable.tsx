import { Pencil, Trash, Eye } from 'lucide-react';
import type { User } from '../types/users';

interface UsersTableProps {
  users: User[];
  onEdit: (userId: string) => void;
  onDelete: (userId: string) => void;
  onView: (userId: string) => void;
}

export function UsersTable({ users, onEdit, onDelete, onView }: UsersTableProps) {
  if (users.length === 0) {
    return (
      <div className="text-center py-8 text-muted-foreground">
        No users found.
      </div>
    );
  }

  const getStatusBadgeClass = (status: User['status']) => {
    switch (status) {
      case 'active':
        return 'bg-green-100 text-green-800';
      case 'inactive':
        return 'bg-gray-100 text-gray-800';
      default:
        return 'bg-gray-100 text-gray-800';
    }
  };

  const getRoleBadgeClass = (role: string) => {
    switch (role) {
      case 'super':
        return 'bg-purple-100 text-purple-800';
      case 'admin':
        return 'bg-blue-100 text-blue-800';
      case 'editor':
        return 'bg-yellow-100 text-yellow-800';
      case 'author':
        return 'bg-green-100 text-green-800';
      case 'subscriber':
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
            <th className="px-4 py-3 text-left text-sm font-medium">User</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Email</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Role</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Status</th>
            <th className="px-4 py-3 text-left text-sm font-medium">Created</th>
            <th className="px-4 py-3 text-right text-sm font-medium">Actions</th>
          </tr>
        </thead>
        <tbody>
          {users.map((user) => (
            <tr key={user.id} className="border-t hover:bg-muted/50">
              <td className="px-4 py-3">
                <div className="flex items-center gap-3">
                  {user.avatar ? (
                    <img
                      src={user.avatar}
                      alt={user.name}
                      className="h-8 w-8 rounded-full object-cover"
                    />
                  ) : (
                    <div className="h-8 w-8 rounded-full bg-muted flex items-center justify-center text-sm font-medium">
                      {user.name.charAt(0).toUpperCase()}
                    </div>
                  )}
                  <span className="font-medium">{user.name}</span>
                </div>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {user.email}
              </td>
              <td className="px-4 py-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getRoleBadgeClass(user.role)}`}>
                  {user.role}
                </span>
              </td>
              <td className="px-4 py-3">
                <span className={`inline-flex items-center px-2 py-1 rounded-full text-xs font-medium ${getStatusBadgeClass(user.status)}`}>
                  {user.status}
                </span>
              </td>
              <td className="px-4 py-3 text-sm text-muted-foreground">
                {new Date(user.createdAt).toLocaleDateString()}
              </td>
              <td className="px-4 py-3 text-right">
                <div className="flex items-center justify-end gap-1">
                  <button
                    onClick={() => onView(user.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="View"
                    aria-label="View"
                  >
                    <Eye className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onEdit(user.id)}
                    className="p-2 hover:bg-muted rounded-md"
                    title="Edit"
                    aria-label="Edit"
                  >
                    <Pencil className="h-4 w-4" />
                  </button>
                  <button
                    onClick={() => onDelete(user.id)}
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
