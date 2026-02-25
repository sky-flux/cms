import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { UserMinus } from 'lucide-react';

import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { Label } from '@/components/ui/label';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/ui/select';
import type { SiteUser } from '@/lib/system-api';

interface SiteUsersDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  siteUsers: SiteUser[];
  loading: boolean;
  onAssignRole: (userId: string, role: string) => void;
  onRemoveUser: (userId: string) => void;
  assignLoading: boolean;
}

const ROLES = ['admin', 'editor', 'author', 'contributor'];

export function SiteUsersDialog({
  open,
  onOpenChange,
  siteUsers,
  loading,
  onAssignRole,
  onRemoveUser,
  assignLoading,
}: SiteUsersDialogProps) {
  const { t } = useTranslation();
  const [newUserId, setNewUserId] = useState('');
  const [newRole, setNewRole] = useState('editor');

  const handleAssign = () => {
    if (!newUserId.trim()) return;
    onAssignRole(newUserId.trim(), newRole);
    setNewUserId('');
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="sm:max-w-lg">
        <DialogHeader>
          <DialogTitle>{t('system.sites.siteUsers')}</DialogTitle>
          <DialogDescription className="sr-only">
            {t('system.sites.manageUsers')}
          </DialogDescription>
        </DialogHeader>

        {/* Assign user form */}
        <div className="flex items-end gap-2 border-b pb-4">
          <div className="flex-1 space-y-1">
            <Label>{t('system.sites.assignUser')}</Label>
            <Input
              placeholder={t('common.email')}
              value={newUserId}
              onChange={(e) => setNewUserId(e.target.value)}
            />
          </div>
          <div className="w-32 space-y-1">
            <Label>{t('system.users.role')}</Label>
            <Select value={newRole} onValueChange={setNewRole}>
              <SelectTrigger>
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                {ROLES.map((role) => (
                  <SelectItem key={role} value={role}>
                    {role}
                  </SelectItem>
                ))}
              </SelectContent>
            </Select>
          </div>
          <Button
            onClick={handleAssign}
            disabled={assignLoading || !newUserId.trim()}
            size="sm"
          >
            {t('system.sites.assignUser')}
          </Button>
        </div>

        {/* User list */}
        <div className="space-y-2 max-h-[300px] overflow-y-auto">
          {loading ? (
            <p className="text-center text-sm text-muted-foreground py-4">
              {t('common.loading')}
            </p>
          ) : siteUsers.length === 0 ? (
            <p className="text-center text-sm text-muted-foreground py-4">
              {t('system.users.noUsersFound')}
            </p>
          ) : (
            siteUsers.map((su) => (
              <div
                key={su.user.id}
                className="flex items-center justify-between rounded-md border px-3 py-2"
              >
                <div className="flex flex-col">
                  <div className="flex items-center gap-2">
                    <span className="font-medium text-sm">{su.user.display_name}</span>
                    <Badge variant="secondary" className="text-xs">
                      {su.role}
                    </Badge>
                  </div>
                  <span className="text-xs text-muted-foreground">
                    {su.user.email}
                  </span>
                </div>
                <Button
                  variant="ghost"
                  size="sm"
                  aria-label={t('system.sites.removeUser')}
                  onClick={() => onRemoveUser(su.user.id)}
                >
                  <UserMinus className="h-4 w-4 text-destructive" />
                </Button>
              </div>
            ))
          )}
        </div>
      </DialogContent>
    </Dialog>
  );
}
