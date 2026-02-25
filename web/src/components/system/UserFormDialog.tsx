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
  FormDescription,
  FormField,
  FormItem,
  FormLabel,
  FormMessage,
} from '@/components/ui/form';
import { Input } from '@/components/ui/input';
import { Button } from '@/components/ui/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { User, Role, CreateUserDTO, UpdateUserDTO } from '@/lib/system-api';

const createUserSchema = z.object({
  email: z.string().email(),
  display_name: z.string().min(1),
  password: z.string().min(8),
  role: z.string().min(1),
});

const editUserSchema = z.object({
  display_name: z.string().min(1),
  role: z.string().min(1),
  is_active: z.boolean(),
});

type CreateFormValues = z.infer<typeof createUserSchema>;
type EditFormValues = z.infer<typeof editUserSchema>;

interface UserFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onSubmit: (data: CreateUserDTO | UpdateUserDTO) => void;
  roles: Role[];
  loading?: boolean;
  user?: User;
  defaultRole?: string;
}

export function UserFormDialog({
  open,
  onOpenChange,
  onSubmit,
  roles,
  loading = false,
  user,
  defaultRole = '',
}: UserFormDialogProps) {
  const { t } = useTranslation();
  const isEdit = !!user;

  const createForm = useForm<CreateFormValues>({
    resolver: zodResolver(createUserSchema),
    defaultValues: {
      email: '',
      display_name: '',
      password: '',
      role: defaultRole,
    },
  });

  const editForm = useForm<EditFormValues>({
    resolver: zodResolver(editUserSchema),
    defaultValues: {
      display_name: user?.display_name ?? '',
      role: user?.role ?? '',
      is_active: user?.is_active ?? true,
    },
  });

  // Reset forms when user changes or dialog opens
  useEffect(() => {
    if (open) {
      if (isEdit && user) {
        editForm.reset({
          display_name: user.display_name,
          role: user.role,
          is_active: user.is_active,
        });
      } else {
        createForm.reset({
          email: '',
          display_name: '',
          password: '',
          role: defaultRole,
        });
      }
    }
  }, [open, user, isEdit, createForm, editForm]);

  const handleCreateSubmit = (data: CreateFormValues) => {
    onSubmit({
      email: data.email,
      display_name: data.display_name,
      password: data.password,
      role: data.role,
    } as CreateUserDTO);
  };

  const handleEditSubmit = (data: EditFormValues) => {
    onSubmit({
      display_name: data.display_name,
      role: data.role,
      is_active: data.is_active,
    } as UpdateUserDTO);
  };

  if (isEdit) {
    return (
      <Dialog open={open} onOpenChange={onOpenChange}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>{t('system.users.editUser')}</DialogTitle>
          </DialogHeader>
          <Form {...editForm}>
            <form onSubmit={editForm.handleSubmit(handleEditSubmit)} className="space-y-4">
              <FormField
                control={editForm.control}
                name="display_name"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('system.users.displayName')}</FormLabel>
                    <FormControl>
                      <Input {...field} />
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
              <FormField
                control={editForm.control}
                name="role"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>{t('system.users.role')}</FormLabel>
                    <Select onValueChange={field.onChange} value={field.value}>
                      <FormControl>
                        <SelectTrigger data-testid="role-select-trigger">
                          <SelectValue placeholder={t('system.users.role')} />
                        </SelectTrigger>
                      </FormControl>
                      <SelectContent>
                        {roles.map((role) => (
                          <SelectItem key={role.id} value={role.slug}>
                            {role.name}
                          </SelectItem>
                        ))}
                      </SelectContent>
                    </Select>
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
          <DialogTitle>{t('system.users.newUser')}</DialogTitle>
        </DialogHeader>
        <Form {...createForm}>
          <form onSubmit={createForm.handleSubmit(handleCreateSubmit)} className="space-y-4">
            <FormField
              control={createForm.control}
              name="email"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.users.email')}</FormLabel>
                  <FormControl>
                    <Input type="email" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={createForm.control}
              name="display_name"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.users.displayName')}</FormLabel>
                  <FormControl>
                    <Input {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={createForm.control}
              name="password"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.users.password')}</FormLabel>
                  <FormControl>
                    <Input type="password" {...field} />
                  </FormControl>
                  <FormDescription>{t('system.users.passwordHelp')}</FormDescription>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={createForm.control}
              name="role"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>{t('system.users.role')}</FormLabel>
                  <Select onValueChange={field.onChange} value={field.value}>
                    <FormControl>
                      <SelectTrigger data-testid="role-select-trigger">
                        <SelectValue placeholder={t('system.users.role')} />
                      </SelectTrigger>
                    </FormControl>
                    <SelectContent>
                      {roles.map((role) => (
                        <SelectItem key={role.id} value={role.slug}>
                          {role.name}
                        </SelectItem>
                      ))}
                    </SelectContent>
                  </Select>
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
