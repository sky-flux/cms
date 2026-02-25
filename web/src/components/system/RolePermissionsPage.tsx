import { useState, useCallback, useEffect } from 'react';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useTranslation } from 'react-i18next';
import { toast } from 'sonner';

import { QueryProvider } from '@/components/providers/QueryProvider';
import { I18nProvider } from '@/components/providers/I18nProvider';
import { RolePermissions } from './RolePermissions';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { rolesApi, rbacApi, templatesApi } from '@/lib/system-api';

interface RolePermissionsPageProps {
  roleId: string;
}

function RolePermissionsInner({ roleId }: RolePermissionsPageProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();

  const [checkedApiIds, setCheckedApiIds] = useState<string[]>([]);
  const [checkedMenuIds, setCheckedMenuIds] = useState<string[]>([]);
  const [templateDialogOpen, setTemplateDialogOpen] = useState(false);
  const [selectedTemplateId, setSelectedTemplateId] = useState<string>('');
  const [selectedTemplateName, setSelectedTemplateName] = useState<string>('');

  const { data: roleData } = useQuery({
    queryKey: ['role', roleId],
    queryFn: () => rolesApi.get(roleId),
  });

  const { data: apisData } = useQuery({
    queryKey: ['rbac-apis'],
    queryFn: () => rbacApi.listApis(),
  });

  const { data: menusData } = useQuery({
    queryKey: ['rbac-admin-menus'],
    queryFn: () => rbacApi.listAdminMenus(),
  });

  const { data: templatesData } = useQuery({
    queryKey: ['rbac-templates'],
    queryFn: () => templatesApi.list(),
  });

  const { data: roleApisData } = useQuery({
    queryKey: ['role-apis', roleId],
    queryFn: () => rolesApi.getApis(roleId),
  });

  const { data: roleMenusData } = useQuery({
    queryKey: ['role-menus', roleId],
    queryFn: () => rolesApi.getMenus(roleId),
  });

  // Initialize checked IDs from fetched data
  useEffect(() => {
    if (roleApisData?.data) setCheckedApiIds(roleApisData.data);
  }, [roleApisData]);

  useEffect(() => {
    if (roleMenusData?.data) setCheckedMenuIds(roleMenusData.data);
  }, [roleMenusData]);

  const saveApisMutation = useMutation({
    mutationFn: () => rolesApi.setApis(roleId, checkedApiIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['role-apis', roleId] });
    },
  });

  const saveMenusMutation = useMutation({
    mutationFn: () => rolesApi.setMenus(roleId, checkedMenuIds),
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['role-menus', roleId] });
    },
  });

  const applyTemplateMutation = useMutation({
    mutationFn: (templateId: string) => templatesApi.apply(roleId, templateId),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['role-apis', roleId] });
      queryClient.invalidateQueries({ queryKey: ['role-menus', roleId] });
      setTemplateDialogOpen(false);
    },
  });

  const handleSave = useCallback(async () => {
    await Promise.all([
      saveApisMutation.mutateAsync(),
      saveMenusMutation.mutateAsync(),
    ]);
    toast.success(t('messages.saveSuccess'));
  }, [saveApisMutation, saveMenusMutation, t]);

  const handleApplyTemplate = useCallback(
    (templateId: string) => {
      const template = templates.find((t) => t.id === templateId);
      setSelectedTemplateId(templateId);
      setSelectedTemplateName(template?.name ?? '');
      setTemplateDialogOpen(true);
    },
    [],
  );

  const role = roleData?.data;
  const apis = apisData?.data ?? [];
  const menus = menusData?.data ?? [];
  const templates = templatesData?.data ?? [];

  if (!role) {
    return (
      <div className="p-6">
        <p className="text-muted-foreground">{t('common.loading')}</p>
      </div>
    );
  }

  return (
    <div className="p-6">
      <div className="mb-4">
        <a href="/dashboard/roles" className="text-sm text-muted-foreground hover:underline">
          &larr; {t('common.back')}
        </a>
      </div>
      <RolePermissions
        roleId={roleId}
        roleName={role.name}
        apis={apis}
        menus={menus}
        templates={templates}
        checkedApiIds={checkedApiIds}
        checkedMenuIds={checkedMenuIds}
        onApiChange={setCheckedApiIds}
        onMenuChange={setCheckedMenuIds}
        onSave={handleSave}
        onApplyTemplate={handleApplyTemplate}
        saving={saveApisMutation.isPending || saveMenusMutation.isPending}
      />
      <ConfirmDialog
        open={templateDialogOpen}
        onOpenChange={setTemplateDialogOpen}
        title={t('system.roles.applyTemplate')}
        description={t('system.roles.applyTemplateConfirm', { name: selectedTemplateName })}
        onConfirm={() => applyTemplateMutation.mutate(selectedTemplateId)}
        loading={applyTemplateMutation.isPending}
        variant="warning"
      />
    </div>
  );
}

export function RolePermissionsPage({ roleId }: RolePermissionsPageProps) {
  return (
    <QueryProvider>
      <I18nProvider>
        <RolePermissionsInner roleId={roleId} />
      </I18nProvider>
    </QueryProvider>
  );
}
