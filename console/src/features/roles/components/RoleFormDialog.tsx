import { useState } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogFooter,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import type { Role, CreateRoleRequest, UpdateRoleRequest, Permission } from '../types/roles';
import { RolePermissions } from './RolePermissions';

const roleSchema = z.object({
  name: z.string().min(1, 'Name is required').max(50, 'Name must be at most 50 characters'),
  slug: z.string().min(1, 'Slug is required').max(50, 'Slug must be at most 50 characters').regex(/^[a-z0-9_]+$/, 'Slug must contain only lowercase letters, numbers, and underscores'),
  description: z.string().optional(),
});

type RoleFormValues = z.infer<typeof roleSchema>;

interface RoleFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  role?: Role | null;
  permissions?: Permission[];
  selectedPermissionIds?: string[];
  onPermissionChange?: (permissionIds: string[]) => void;
  onSubmit: (data: CreateRoleRequest | UpdateRoleRequest, permissionIds: string[]) => void;
  isLoading?: boolean;
}

export function RoleFormDialog({
  open,
  onOpenChange,
  role,
  permissions = [],
  selectedPermissionIds = [],
  onPermissionChange,
  onSubmit,
  isLoading = false,
}: RoleFormDialogProps) {
  const isEdit = !!role;

  const form = useForm<RoleFormValues>({
    resolver: zodResolver(roleSchema),
    defaultValues: {
      name: role?.name || '',
      slug: role?.slug || '',
      description: role?.description || '',
    },
  });

  const [localPermissionIds, setLocalPermissionIds] = useState<string[]>(selectedPermissionIds);

  const handleSubmit = (data: RoleFormValues) => {
    const permissionIds = onPermissionChange ? selectedPermissionIds : localPermissionIds;
    if (isEdit) {
      const updateData: UpdateRoleRequest = {
        name: data.name,
        description: data.description,
      };
      onSubmit(updateData, permissionIds);
    } else {
      const createData: CreateRoleRequest = {
        name: data.name,
        slug: data.slug,
        description: data.description,
      };
      onSubmit(createData, permissionIds);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      form.reset();
      setLocalPermissionIds([]);
    }
    onOpenChange(newOpen);
  };

  const handlePermissionChange = (permissionIds: string[]) => {
    setLocalPermissionIds(permissionIds);
    if (onPermissionChange) {
      onPermissionChange(permissionIds);
    }
  };

  // Auto-generate slug from name when creating
  const handleNameChange = (value: string) => {
    if (!isEdit && value) {
      const slug = value.toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '');
      const currentSlug = form.getValues('slug');
      if (!currentSlug || currentSlug === form.getValues('name').toLowerCase().replace(/[^a-z0-9]+/g, '_').replace(/^_|_$/g, '')) {
        form.setValue('slug', slug);
      }
    }
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[600px] max-h-[90vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Role' : 'Create Role'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the role details below.'
              : 'Fill in the details to create a new role.'}
          </DialogDescription>
        </DialogHeader>
        <Form {...form}>
          <form onSubmit={form.handleSubmit(handleSubmit)} className="space-y-4">
            <FormField
              control={form.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Name</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Editor"
                      {...field}
                      onChange={(e) => {
                        field.onChange(e);
                        handleNameChange(e.target.value);
                      }}
                      disabled={isEdit}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            {!isEdit && (
              <FormField
                control={form.control}
                name="slug"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Slug</FormLabel>
                    <FormControl>
                      <Input
                        placeholder="editor"
                        {...field}
                        disabled={isEdit}
                      />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
            )}
            <FormField
              control={form.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Description</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="Role description (optional)"
                      {...field}
                      value={field.value || ''}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            {permissions.length > 0 && (
              <div className="space-y-2">
                <FormLabel>Permissions</FormLabel>
                <RolePermissions
                  permissions={permissions}
                  selectedPermissionIds={onPermissionChange ? selectedPermissionIds : localPermissionIds}
                  onPermissionChange={handlePermissionChange}
                  disabled={isLoading}
                />
              </div>
            )}
            <DialogFooter>
              <Button
                type="button"
                variant="outline"
                onClick={() => handleOpenChange(false)}
              >
                Cancel
              </Button>
              <Button type="submit" disabled={isLoading}>
                {isLoading ? 'Saving...' : isEdit ? 'Update' : 'Create'}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
