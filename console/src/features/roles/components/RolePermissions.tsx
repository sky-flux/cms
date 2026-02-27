import { Checkbox } from '@/components/ui/checkbox';
import { Label } from '@/components/ui/label';
import type { Permission } from '../types/roles';

interface RolePermissionsProps {
  permissions: Permission[];
  selectedPermissionIds: string[];
  onPermissionChange: (permissionIds: string[]) => void;
  disabled?: boolean;
}

export function RolePermissions({
  permissions,
  selectedPermissionIds,
  onPermissionChange,
  disabled = false,
}: RolePermissionsProps) {
  // Group permissions by category
  const groupedPermissions = permissions.reduce<Record<string, Permission[]>>(
    (acc, permission) => {
      const category = permission.category || 'Other';
      if (!acc[category]) {
        acc[category] = [];
      }
      acc[category].push(permission);
      return acc;
    },
    {}
  );

  const categories = Object.keys(groupedPermissions).sort();

  const handlePermissionToggle = (permissionId: string) => {
    const newSelected = selectedPermissionIds.includes(permissionId)
      ? selectedPermissionIds.filter((id) => id !== permissionId)
      : [...selectedPermissionIds, permissionId];
    onPermissionChange(newSelected);
  };

  const handleCategoryToggle = (category: string, checked: boolean) => {
    const categoryPermissions = groupedPermissions[category] || [];
    const categoryIds = categoryPermissions.map((p) => p.id);

    if (checked) {
      const newSelected = [...new Set([...selectedPermissionIds, ...categoryIds])];
      onPermissionChange(newSelected);
    } else {
      const newSelected = selectedPermissionIds.filter(
        (id) => !categoryIds.includes(id)
      );
      onPermissionChange(newSelected);
    }
  };

  const isCategoryFullySelected = (category: string) => {
    const categoryPermissions = groupedPermissions[category] || [];
    return categoryPermissions.every((p) => selectedPermissionIds.includes(p.id));
  };

  const isCategoryPartiallySelected = (category: string) => {
    const categoryPermissions = groupedPermissions[category] || [];
    const selectedCount = categoryPermissions.filter((p) =>
      selectedPermissionIds.includes(p.id)
    ).length;
    return selectedCount > 0 && selectedCount < categoryPermissions.length;
  };

  if (permissions.length === 0) {
    return (
      <div className="text-center py-4 text-muted-foreground text-sm">
        No permissions available.
      </div>
    );
  }

  return (
    <div className="space-y-6">
      {categories.map((category) => (
        <div key={category} className="space-y-3">
          <div className="flex items-center gap-2">
            <Checkbox
              id={`category-${category}`}
              checked={isCategoryFullySelected(category)}
              indeterminate={isCategoryPartiallySelected(category)}
              onCheckedChange={(checked) =>
                handleCategoryToggle(category, checked === true)
              }
              disabled={disabled}
            />
            <Label
              htmlFor={`category-${category}`}
              className="font-medium text-sm cursor-pointer"
            >
              {category}
            </Label>
          </div>
          <div className="ml-6 space-y-2">
            {(groupedPermissions[category] || []).map((permission) => (
              <div
                key={permission.id}
                className="flex items-start gap-2"
              >
                <Checkbox
                  id={`permission-${permission.id}`}
                  checked={selectedPermissionIds.includes(permission.id)}
                  onCheckedChange={() => handlePermissionToggle(permission.id)}
                  disabled={disabled}
                />
                <Label
                  htmlFor={`permission-${permission.id}`}
                  className="cursor-pointer leading-snug"
                >
                  <span className="font-medium">{permission.name}</span>
                  {permission.description && (
                    <span className="text-muted-foreground text-xs ml-2">
                      {permission.description}
                    </span>
                  )}
                </Label>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
}
