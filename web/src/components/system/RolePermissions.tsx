import { useMemo } from 'react';
import { useTranslation } from 'react-i18next';

import { Tabs, TabsContent, TabsList, TabsTrigger } from '@/components/ui/tabs';
import { Button } from '@/components/ui/button';
import { PermissionTree } from '@/components/shared/PermissionTree';
import type { TreeNode } from '@/components/shared/PermissionTree';
import type { ApiEndpoint, AdminMenu, RoleTemplate } from '@/lib/system-api';

interface RolePermissionsProps {
  roleId: string;
  roleName: string;
  apis: ApiEndpoint[];
  menus: AdminMenu[];
  templates: RoleTemplate[];
  checkedApiIds: string[];
  checkedMenuIds: string[];
  onApiChange: (ids: string[]) => void;
  onMenuChange: (ids: string[]) => void;
  onSave: () => void;
  onApplyTemplate: (templateId: string) => void;
  saving: boolean;
}

function apisToTreeNodes(apis: ApiEndpoint[]): TreeNode[] {
  return apis.map((api) => ({
    id: api.id,
    label: `${api.method} ${api.path} — ${api.description}`,
    children: [],
  }));
}

function menusToTreeNodes(menus: AdminMenu[]): TreeNode[] {
  return menus.map((menu) => ({
    id: menu.id,
    label: menu.name,
    children: menusToTreeNodes(menu.children),
  }));
}

function getAllApiIds(apis: ApiEndpoint[]): string[] {
  return apis.map((a) => a.id);
}

function getAllMenuIds(menus: AdminMenu[]): string[] {
  const ids: string[] = [];
  for (const menu of menus) {
    ids.push(menu.id);
    ids.push(...getAllMenuIds(menu.children));
  }
  return ids;
}

export function RolePermissions({
  roleId,
  roleName,
  apis,
  menus,
  templates,
  checkedApiIds,
  checkedMenuIds,
  onApiChange,
  onMenuChange,
  onSave,
  onApplyTemplate,
  saving,
}: RolePermissionsProps) {
  const { t } = useTranslation();

  const apiTree = useMemo(() => apisToTreeNodes(apis), [apis]);
  const menuTree = useMemo(() => menusToTreeNodes(menus), [menus]);
  const allApiIds = useMemo(() => getAllApiIds(apis), [apis]);
  const allMenuIds = useMemo(() => getAllMenuIds(menus), [menus]);

  return (
    <div className="space-y-6">
      <div className="flex items-center justify-between">
        <h2 className="text-lg font-semibold">
          {t('system.roles.permissions')} — {roleName}
        </h2>
        <div className="flex items-center gap-2">
          <Button
            variant="outline"
            size="sm"
            onClick={() => {
              if (templates.length > 0) {
                onApplyTemplate(templates[0].id);
              }
            }}
            aria-label={t('system.roles.applyTemplate')}
          >
            {t('system.roles.applyTemplate')}
          </Button>
          <Button
            onClick={onSave}
            disabled={saving}
            aria-label={t('common.save')}
          >
            {saving ? t('common.loading') : t('common.save')}
          </Button>
        </div>
      </div>

      <Tabs defaultValue="api">
        <TabsList>
          <TabsTrigger value="api">{t('system.roles.apiPermissions')}</TabsTrigger>
          <TabsTrigger value="menu">{t('system.roles.menuPermissions')}</TabsTrigger>
        </TabsList>

        <TabsContent value="api" className="space-y-3">
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onApiChange(allApiIds)}
              aria-label={t('system.roles.selectAll')}
            >
              {t('system.roles.selectAll')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onApiChange([])}
              aria-label={t('system.roles.deselectAll')}
            >
              {t('system.roles.deselectAll')}
            </Button>
          </div>
          <PermissionTree
            items={apiTree}
            checkedIds={checkedApiIds}
            onChange={onApiChange}
          />
        </TabsContent>

        <TabsContent value="menu" className="space-y-3">
          <div className="flex items-center gap-2">
            <Button
              variant="outline"
              size="sm"
              onClick={() => onMenuChange(allMenuIds)}
              aria-label={t('system.roles.selectAll')}
            >
              {t('system.roles.selectAll')}
            </Button>
            <Button
              variant="outline"
              size="sm"
              onClick={() => onMenuChange([])}
              aria-label={t('system.roles.deselectAll')}
            >
              {t('system.roles.deselectAll')}
            </Button>
          </div>
          <PermissionTree
            items={menuTree}
            checkedIds={checkedMenuIds}
            onChange={onMenuChange}
          />
        </TabsContent>
      </Tabs>
    </div>
  );
}
