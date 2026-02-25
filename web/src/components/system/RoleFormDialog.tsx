import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { z } from 'zod';
import { useTranslation } from 'react-i18next';

import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import {
  Form,
  FormControl,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import type { Role, CreateRoleDTO, UpdateRoleDTO } from '@/lib/system-api';

const slugRegex = /^[a-z0-9-]+$/;

const createRoleSchema = z.object({
  name: z.string().min(1),
  slug: z.string().min(1).regex(slugRegex, 'Only lowercase letters, numbers, and hyphens'),
  description: z.string().optional(),
});

const editRoleSchema = z.object({
  name: z.string().min(1),
  slug: z.string(), // read-only in edit mode
  description: z.string().optional(),
});

type CreateFormValues = z.infer<typeof createRoleSchema>;
type EditFormValues = z.infer<typeof editRoleSchema>;

interface RoleFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateRoleDTO | UpdateRoleDTO) => void;
  loading?: boolean;
  role?: Role;
}

export function RoleFormDialog({
  open,
  onOpenChange,
  onSubmit,
  loading = false,
  role,
}: RoleFormDialogProps) {
  const { t } = useTranslation();
  const isEdit = !!role;

  const createForm = useForm<CreateFormValues>({
    resolver: zodResolver(createRoleSchema),
    defaultValues: {
      name: '',
      slug: '',
      description: '',
    },
  });

  const editForm = useForm<EditFormValues>({
    resolver: zodResolver(editRoleSchema),
    defaultValues: {
      name: role?.name ?? '',
      slug: role?.slug ?? '',
      description: role?.description ?? '',
    },
  });

  useEffect(() => {
    if (open) {
      if (isEdit && role) {
        editForm.reset({
          name: role.name,
          slug: role.slug,
          description: role.description ?? '',
        });
      } else {
        createForm.reset({
          name: '',
          slug: '',
          description: '',
        });
      }
    }
  }, [open, role, isEdit, createForm, editForm]);

  const handleCreateSubmit = (data: CreateFormValues) => {
    onSubmit({
      name: data.name,
      slug: data.slug,
      description: data.description,
    } as CreateRoleDTO);
  };

  const handleEditSubmit = (data: EditFormValues) => {
    onSubmit({
      name: data.name,
      description: data.description,
    } as UpdateRoleDTO);
  };

  if (isEdit) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('system.roles.editRole')}</DialogTitle>
          </DialogHeader>
          <Form {...editForm}>
            <form onSubmit={editForm.handleSubmit(handleEditSubmit)} className="space-y-4">
              <FormField
                control={editForm.control}
                name="name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('system.roles.roleName')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={editForm.control}
                name="slug"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('system.roles.roleSlug')}</FormLabel>
                    <FormControl>
                      <Input {...field} disabled />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={editForm.control}
                name="description"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('system.roles.description')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <DialogFooter>
                <Button type="submit" disabled={loading}>
                  {loading ? t('common.loading') : t('common.save')}
                </Button>
              </DialogFooter>
            </form>
          </Form>
        </DialogContent>
      </Dialog>
    );
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent>
        <DialogHeader>
          <DialogTitle>{t('system.roles.newRole')}</DialogTitle>
        </DialogHeader>
        <Form {...createForm}>
          <form onSubmit={createForm.handleSubmit(handleCreateSubmit)} className="space-y-4">
            <FormField
              control={createForm.control}
              name="name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.roles.roleName')}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={createForm.control}
              name="slug"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.roles.roleSlug')}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={createForm.control}
              name="description"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.roles.description')}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <DialogFooter>
              <Button type="submit" disabled={loading}>
                {loading ? t('common.loading') : t('common.save')}
              </Button>
            </DialogFooter>
          </form>
        </Form>
      </DialogContent>
    </Dialog>
  );
}
