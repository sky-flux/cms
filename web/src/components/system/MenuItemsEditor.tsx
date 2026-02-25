import { useTranslation } from 'react-i18next';
import {
  Plus,
  Pencil,
  Trash2,
  ChevronUp,
  ChevronDown,
  AlertTriangle,
} from 'lucide-react';

import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Switch } from '@/components/ui/switch';
import type { SiteMenuDetail, SiteMenuItem } from '@/lib/system-api';

interface MenuItemsEditorProps {
  menuDetail: SiteMenuDetail;
  loading: boolean;
  onAddItem: () => void;
  onEditItem: (item: SiteMenuItem) => void;
  onDeleteItem: (itemId: string) => void;
  onToggleActive: (itemId: string, isActive: boolean) => void;
  onMoveUp: (itemId: string) => void;
  onMoveDown: (itemId: string) => void;
}

const typeColors: Record<string, string> = {
  custom: 'bg-gray-100 text-gray-700 dark:bg-gray-800 dark:text-gray-300',
  post: 'bg-blue-100 text-blue-700 dark:bg-blue-900 dark:text-blue-300',
  category: 'bg-green-100 text-green-700 dark:bg-green-900 dark:text-green-300',
  tag: 'bg-yellow-100 text-yellow-700 dark:bg-yellow-900 dark:text-yellow-300',
  page: 'bg-purple-100 text-purple-700 dark:bg-purple-900 dark:text-purple-300',
};

interface ItemRowProps {
  item: SiteMenuItem;
  depth: number;
  onEditItem: (item: SiteMenuItem) => void;
  onDeleteItem: (itemId: string) => void;
  onToggleActive: (itemId: string, isActive: boolean) => void;
  onMoveUp: (itemId: string) => void;
  onMoveDown: (itemId: string) => void;
  t: (key: string, params?: Record<string, unknown>) => string;
}

function ItemRow({
  item,
  depth,
  onEditItem,
  onDeleteItem,
  onToggleActive,
  onMoveUp,
  onMoveDown,
  t,
}: ItemRowProps) {
  const typeKey = `type${item.type.charAt(0).toUpperCase()}${item.type.slice(1)}`;

  return (
    <>
      <div
        className="flex items-center gap-3 rounded-md border p-3 hover:bg-muted/50"
        style={{ marginLeft: depth * 24 }}
      >
        {/* Move buttons */}
        <div className="flex flex-col gap-0.5">
          <Button
            variant="ghost"
            size="sm"
            className="h-5 w-5 p-0"
            onClick={() => onMoveUp(item.id)}
            aria-label="Move up"
          >
            <ChevronUp className="h-3 w-3" />
          </Button>
          <Button
            variant="ghost"
            size="sm"
            className="h-5 w-5 p-0"
            onClick={() => onMoveDown(item.id)}
            aria-label="Move down"
          >
            <ChevronDown className="h-3 w-3" />
          </Button>
        </div>

        {/* Label */}
        <span className="font-medium">{item.label}</span>

        {/* Type badge */}
        <Badge variant="outline" className={typeColors[item.type] ?? ''}>
          {t(`system.menus.${typeKey}`)}
        </Badge>

        {/* URL */}
        {item.url && (
          <code className="text-sm text-muted-foreground">{item.url}</code>
        )}

        {/* Broken reference warning */}
        {item.is_broken && (
          <span title={t('system.menus.broken')}>
            <AlertTriangle className="h-4 w-4 text-destructive" />
          </span>
        )}

        {/* Spacer */}
        <div className="flex-1" />

        {/* Active toggle */}
        <Switch
          checked={item.is_active}
          onCheckedChange={(checked) => onToggleActive(item.id, checked)}
          aria-label={t('system.menus.itemActive')}
        />

        {/* Edit/Delete */}
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 p-0"
          onClick={() => onEditItem(item)}
          aria-label={t('system.menus.editItem')}
        >
          <Pencil className="h-3.5 w-3.5" />
        </Button>
        <Button
          variant="ghost"
          size="sm"
          className="h-7 w-7 p-0 text-destructive hover:text-destructive"
          onClick={() => onDeleteItem(item.id)}
          aria-label={t('system.menus.deleteItem')}
        >
          <Trash2 className="h-3.5 w-3.5" />
        </Button>
      </div>

      {/* Recursive children */}
      {item.children?.map((child) => (
        <ItemRow
          key={child.id}
          item={child}
          depth={depth + 1}
          onEditItem={onEditItem}
          onDeleteItem={onDeleteItem}
          onToggleActive={onToggleActive}
          onMoveUp={onMoveUp}
          onMoveDown={onMoveDown}
          t={t}
        />
      ))}
    </>
  );
}

export function MenuItemsEditor({
  menuDetail,
  loading,
  onAddItem,
  onEditItem,
  onDeleteItem,
  onToggleActive,
  onMoveUp,
  onMoveDown,
}: MenuItemsEditorProps) {
  const { t } = useTranslation();

  return (
    <div className="space-y-4">
      <div className="flex items-center justify-between">
        <h2 className="text-xl font-semibold">{menuDetail.name}</h2>
        <Button onClick={onAddItem} aria-label={t('system.menus.addItem')}>
          <Plus className="mr-2 h-4 w-4" />
          {t('system.menus.addItem')}
        </Button>
      </div>

      {menuDetail.items.length === 0 ? (
        <div className="flex flex-col items-center gap-2 rounded-md border py-8">
          <p className="text-muted-foreground">{t('system.menus.noItems')}</p>
        </div>
      ) : (
        <div className="space-y-2">
          {menuDetail.items.map((item) => (
            <ItemRow
              key={item.id}
              item={item}
              depth={0}
              onEditItem={onEditItem}
              onDeleteItem={onDeleteItem}
              onToggleActive={onToggleActive}
              onMoveUp={onMoveUp}
              onMoveDown={onMoveDown}
              t={t}
            />
          ))}
        </div>
      )}
    </div>
  );
}
