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
import type { Site, CreateSiteRequest, UpdateSiteRequest } from '../types/sites';

const siteSchema = z.object({
  name: z.string().min(1, 'Name is required'),
  slug: z.string().min(1, 'Slug is required').regex(/^[a-z0-9_]{3,50}$/, 'Slug must be 3-50 characters, lowercase letters, numbers, and underscores only'),
  domain: z.string().optional(),
  status: z.enum(['active', 'inactive']).optional(),
});

type SiteFormValues = z.infer<typeof siteSchema>;

interface SiteFormDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  site?: Site | null;
  onSubmit: (data: CreateSiteRequest | UpdateSiteRequest) => void;
  isLoading?: boolean;
}

export function SiteFormDialog({
  open,
  onOpenChange,
  site,
  onSubmit,
  isLoading = false,
}: SiteFormDialogProps) {
  const isEdit = !!site;

  const form = useForm<SiteFormValues>({
    resolver: zodResolver(siteSchema),
    defaultValues: {
      name: site?.name || '',
      slug: site?.slug || '',
      domain: site?.domain || '',
      status: site?.status || 'active',
    },
  });

  const handleSubmit = (data: SiteFormValues) => {
    if (isEdit) {
      const updateData: UpdateSiteRequest = {
        name: data.name,
        domain: data.domain || undefined,
        status: data.status,
      };
      onSubmit(updateData);
    } else {
      onSubmit({
        name: data.name,
        slug: data.slug,
        domain: data.domain || undefined,
      } as CreateSiteRequest);
    }
  };

  const handleOpenChange = (newOpen: boolean) => {
    if (!newOpen) {
      form.reset();
    }
    onOpenChange(newOpen);
  };

  return (
    <Dialog open={open} onOpenChange={handleOpenChange}>
      <DialogContent className="sm:max-w-[500px]">
        <DialogHeader>
          <DialogTitle>{isEdit ? 'Edit Site' : 'Create Site'}</DialogTitle>
          <DialogDescription>
            {isEdit
              ? 'Update the site details below.'
              : 'Fill in the details to create a new site.'}
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
                    <Input placeholder="My Site" {...field} />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="slug"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Slug</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="my-site"
                      {...field}
                      disabled={isEdit}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            <FormField
              control={form.control}
              name="domain"
              render={({ field }) => (
                <FormItem>
                  <FormLabel>Domain (optional)</FormLabel>
                  <FormControl>
                    <Input
                      placeholder="example.com"
                      {...field}
                      value={field.value || ''}
                    />
                  </FormControl>
                  <FormMessage />
                </FormItem>
              )}
            />
            {isEdit && (
              <FormField
                control={form.control}
                name="status"
                render={({ field }) => (
                  <FormItem>
                    <FormLabel>Status</FormLabel>
                    <FormControl>
                      <select
                        className="flex h-10 w-full items-center justify-between rounded-md border border-input bg-background px-3 py-2 text-sm ring-offset-background placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-ring focus:ring-offset-2 disabled:cursor-not-allowed disabled:opacity-50"
                        {...field}
                        value={field.value || 'active'}
                        onChange={(e) => field.onChange(e.target.value)}
                      >
                        <option value="active">Active</option>
                        <option value="inactive">Inactive</option>
                      </select>
                    </FormControl>
                    <FormMessage />
                  </FormItem>
                )}
              />
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
