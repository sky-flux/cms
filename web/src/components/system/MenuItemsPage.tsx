import { useState, useCallback } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { MenuItemsEditor } from './MenuItemsEditor';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { siteMenusApi } from '@/lib/system-api';
import type {
  SiteMenuItem,
  CreateMenuItemDTO,
  UpdateMenuItemDTO,
} from '@/lib/system-api';
import {
  Dialog,
  DialogContent,
  DialogHeader,
  DialogTitle,
  DialogFooter,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import { Switch } from '@/components/ui/switch';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';

interface MenuItemsPageProps {
  menuId: string;
}

function MenuItemsPageInner({ menuId }: MenuItemsPageProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [deleteItemId, setDeleteItemId] = useState<string | null>(null);
  const [itemDialogOpen, setItemDialogOpen] = useState(false);
  const [editingItem, setEditingItem] = useState<SiteMenuItem | undefined>();

  // Form state
  const [itemLabel, setItemLabel] = useState('');
  const [itemUrl, setItemUrl] = useState('');
  const [itemType, setItemType] = useState('custom');
  const [itemTarget, setItemTarget] = useState('_self');
  const [itemReferenceId, setItemReferenceId] = useState('');
  const [itemIcon, setItemIcon] = useState('');
  const [itemCssClass, setItemCssClass] = useState('');
  const [itemIsActive, setItemIsActive] = useState(true);

  const { data, isLoading } = useQuery({
    queryKey: ['site-menu-detail', menuId],
    queryFn: () => siteMenusApi.get(menuId),
  });

  const addItemMutation = useMutation({
    mutationFn: (data: CreateMenuItemDTO) => siteMenusApi.addItem(menuId, data),
    onSuccess: () => {
      toast.success(t('messages.createSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menu-detail', menuId] });
      setItemDialogOpen(false);
    },
  });

  const updateItemMutation = useMutation({
    mutationFn: ({ itemId, data }: { itemId: string; data: UpdateMenuItemDTO }) =>
      siteMenusApi.updateItem(menuId, itemId, data),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menu-detail', menuId] });
      setItemDialogOpen(false);
      setEditingItem(undefined);
    },
  });

  const deleteItemMutation = useMutation({
    mutationFn: (itemId: string) => siteMenusApi.deleteItem(menuId, itemId),
    onSuccess: () => {
      toast.success(t('messages.deleteSuccess'));
      queryClient.invalidateQueries({ queryKey: ['site-menu-detail', menuId] });
      setDeleteItemId(null);
    },
  });

  const toggleActiveMutation = useMutation({
    mutationFn: ({ itemId, isActive }: { itemId: string; isActive: boolean }) =>
      siteMenusApi.updateItem(menuId, itemId, { is_active: isActive }),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['site-menu-detail', menuId] });
    },
  });

  const reorderMutation = useMutation({
    mutationFn: (items: { id: string; parent_id: string | null; sort_order: number }[]) =>
      siteMenusApi.reorderItems(menuId, items),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['site-menu-detail', menuId] });
    },
  });

  const resetItemForm = useCallback(() => {
    setItemLabel('');
    setItemUrl('');
    setItemType('custom');
    setItemTarget('_self');
    setItemReferenceId('');
    setItemIcon('');
    setItemCssClass('');
    setItemIsActive(true);
    setEditingItem(undefined);
  }, []);

  const handleAddItem = useCallback(() => {
    resetItemForm();
    setItemDialogOpen(true);
  }, [resetItemForm]);

  const handleEditItem = useCallback((item: SiteMenuItem) => {
    setEditingItem(item);
    setItemLabel(item.label);
    setItemUrl(item.url ?? '');
    setItemType(item.type);
    setItemTarget(item.target);
    setItemReferenceId(item.reference_id ?? '');
    setItemIcon(item.icon ?? '');
    setItemCssClass(item.css_class ?? '');
    setItemIsActive(item.is_active);
    setItemDialogOpen(true);
  }, []);

  const handleItemSubmit = useCallback(() => {
    if (!itemLabel.trim()) return;

    const itemData: CreateMenuItemDTO = {
      label: itemLabel,
      type: itemType,
      url: itemType === 'custom' ? itemUrl || null : null,
      target: itemTarget,
      reference_id: itemType !== 'custom' ? itemReferenceId || null : null,
      icon: itemIcon || null,
      css_class: itemCssClass || null,
      is_active: itemIsActive,
      sort_order: 0,
    };

    if (editingItem) {
      updateItemMutation.mutate({ itemId: editingItem.id, data: itemData });
    } else {
      addItemMutation.mutate(itemData);
    }
  }, [
    itemLabel,
    itemType,
    itemUrl,
    itemTarget,
    itemReferenceId,
    itemIcon,
    itemCssClass,
    itemIsActive,
    editingItem,
    updateItemMutation,
    addItemMutation,
  ]);

  // Simple move up/down by finding flat list and swapping sort_order
  const flattenItems = useCallback(
    (items: SiteMenuItem[], parentId: string | null = null): { id: string; parent_id: string | null; sort_order: number }[] => {
      const result: { id: string; parent_id: string | null; sort_order: number }[] = [];
      items.forEach((item, idx) => {
        result.push({ id: item.id, parent_id: parentId, sort_order: idx });
        if (item.children?.length) {
          result.push(...flattenItems(item.children, item.id));
        }
      });
      return result;
    },
    [],
  );

  const handleMoveUp = useCallback(
    (itemId: string) => {
      if (!data?.data) return;
      const flat = flattenItems(data.data.items);
      const idx = flat.findIndex((f) => f.id === itemId);
      if (idx <= 0) return;
      // Swap sort_order with previous sibling at same level
      const current = flat[idx];
      const siblings = flat.filter((f) => f.parent_id === current.parent_id);
      const siblingIdx = siblings.findIndex((s) => s.id === itemId);
      if (siblingIdx <= 0) return;
      const prev = siblings[siblingIdx - 1];
      const tempOrder = current.sort_order;
      current.sort_order = prev.sort_order;
      prev.sort_order = tempOrder;
      reorderMutation.mutate(flat);
    },
    [data, flattenItems, reorderMutation],
  );

  const handleMoveDown = useCallback(
    (itemId: string) => {
      if (!data?.data) return;
      const flat = flattenItems(data.data.items);
      const current = flat.find((f) => f.id === itemId);
      if (!current) return;
      const siblings = flat.filter((f) => f.parent_id === current.parent_id);
      const siblingIdx = siblings.findIndex((s) => s.id === itemId);
      if (siblingIdx >= siblings.length - 1) return;
      const next = siblings[siblingIdx + 1];
      const tempOrder = current.sort_order;
      current.sort_order = next.sort_order;
      next.sort_order = tempOrder;
      reorderMutation.mutate(flat);
    },
    [data, flattenItems, reorderMutation],
  );

  const menuDetail = data?.data;

  if (isLoading || !menuDetail) {
    return (
      <div className="p-6">
        <p className="text-muted-foreground">{t('common.loading')}</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-6">
        <a href="/dashboard/menus" className="text-sm text-muted-foreground hover:underline">
          {t('common.back')}
        </a>
      </div>

      <MenuItemsEditor
        menuDetail={menuDetail}
        loading={isLoading}
        onAddItem={handleAddItem}
        onEditItem={handleEditItem}
        onDeleteItem={(id) => setDeleteItemId(id)}
        onToggleActive={(id, active) => toggleActiveMutation.mutate({ itemId: id, isActive: active })}
        onMoveUp={handleMoveUp}
        onMoveDown={handleMoveDown}
      />

      {/* Add/Edit Item Dialog */}
      <Dialog open={itemDialogOpen} onOpenChange={setItemDialogOpen}>
        <DialogContent>
          <DialogHeader>
            <DialogTitle>
              {editingItem ? t('system.menus.editItem') : t('system.menus.addItem')}
            </DialogTitle>
          </DialogHeader>
          <div className="space-y-4">
            <div className="space-y-2">
              <Label>{t('system.menus.itemLabel')}</Label>
              <Input
                value={itemLabel}
                onChange={(e) => setItemLabel(e.target.value)}
                placeholder="Menu item label"
              />
            </div>

            <div className="space-y-2">
              <Label>{t('system.menus.itemType')}</Label>
              <Select value={itemType} onValueChange={setItemType}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="custom">{t('system.menus.typeCustom')}</SelectItem>
                  <SelectItem value="post">{t('system.menus.typePost')}</SelectItem>
                  <SelectItem value="category">{t('system.menus.typeCategory')}</SelectItem>
                  <SelectItem value="tag">{t('system.menus.typeTag')}</SelectItem>
                  <SelectItem value="page">{t('system.menus.typePage')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            {itemType === 'custom' ? (
              <div className="space-y-2">
                <Label>{t('system.menus.itemUrl')}</Label>
                <Input
                  value={itemUrl}
                  onChange={(e) => setItemUrl(e.target.value)}
                  placeholder="https://example.com"
                />
              </div>
            ) : (
              <div className="space-y-2">
                <Label>{t('system.menus.referenceId')}</Label>
                <Input
                  value={itemReferenceId}
                  onChange={(e) => setItemReferenceId(e.target.value)}
                  placeholder="Reference ID"
                />
              </div>
            )}

            <div className="space-y-2">
              <Label>{t('system.menus.itemTarget')}</Label>
              <Select value={itemTarget} onValueChange={setItemTarget}>
                <SelectTrigger>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="_self">{t('system.menus.targetSelf')}</SelectItem>
                  <SelectItem value="_blank">{t('system.menus.targetBlank')}</SelectItem>
                </SelectContent>
              </Select>
            </div>

            <div className="space-y-2">
              <Label>{t('system.menus.itemIcon')}</Label>
              <Input
                value={itemIcon}
                onChange={(e) => setItemIcon(e.target.value)}
                placeholder="icon-name"
              />
            </div>

            <div className="space-y-2">
              <Label>{t('system.menus.itemCssClass')}</Label>
              <Input
                value={itemCssClass}
                onChange={(e) => setItemCssClass(e.target.value)}
                placeholder="custom-class"
              />
            </div>

            <div className="flex items-center gap-2">
              <Switch
                checked={itemIsActive}
                onCheckedChange={setItemIsActive}
              />
              <Label>{t('system.menus.itemActive')}</Label>
            </div>
          </div>
          <DialogFooter>
            <Button variant="outline" onClick={() => setItemDialogOpen(false)}>
              {t('common.cancel')}
            </Button>
            <Button
              onClick={handleItemSubmit}
              disabled={addItemMutation.isPending || updateItemMutation.isPending}
            >
              {addItemMutation.isPending || updateItemMutation.isPending
                ? t('common.loading')
                : t('common.save')}
            </Button>
          </DialogFooter>
        </DialogContent>
      </Dialog>

      {/* Delete Confirm Dialog */}
      <ConfirmDialog
        open={deleteItemId !== null}
        onOpenChange={(open) => {
          if (!open) setDeleteItemId(null);
        }}
        title={t('content.confirmDelete')}
        description={t('system.menus.deleteItemConfirm')}
        onConfirm={() => {
          if (deleteItemId) deleteItemMutation.mutate(deleteItemId);
        }}
        loading={deleteItemMutation.isPending}
        variant="danger"
      />
    </div>
  );
}

export function MenuItemsPage({ menuId }: MenuItemsPageProps) {
  return (
    <QueryProvider>
      <I18nProvider>
        <MenuItemsPageInner menuId={menuId} />
      </I18nProvider>
    </QueryProvider>
  );
}
