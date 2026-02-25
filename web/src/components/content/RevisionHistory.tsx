import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { toast } from 'sonner';
import { postsApi } from '@/lib/content-api';
import type { Revision } from '@/lib/content-api';
import { ConfirmDialog } from '@/components/shared/ConfirmDialog';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card } from '@/components/ui/card';

interface RevisionHistoryProps {
  postId: string;
}

function formatDate(dateStr: string): string {
  return new Date(dateStr).toLocaleString();
}

export function RevisionHistory({ postId }: RevisionHistoryProps) {
  const { t } = useTranslation();
  const queryClient = useQueryClient();
  const [rollbackTarget, setRollbackTarget] = useState<Revision | null>(null);

  const { data, isLoading } = useQuery({
    queryKey: ['revisions', postId],
    queryFn: () => postsApi.getRevisions(postId),
  });

  const rollbackMutation = useMutation({
    mutationFn: ({ revisionId }: { revisionId: string }) =>
      postsApi.rollback(postId, revisionId),
    onSuccess: () => {
      toast.success(t('messages.updateSuccess'));
      queryClient.invalidateQueries({ queryKey: ['revisions', postId] });
      queryClient.invalidateQueries({ queryKey: ['post', postId] });
      setRollbackTarget(null);
    },
    onError: (error: Error) => {
      toast.error(error.message);
    },
  });

  if (isLoading) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">{t('common.loading')}</p>
      </div>
    );
  }

  const revisions = data?.data ?? [];
  const maxVersion = revisions.length > 0 ? Math.max(...revisions.map((r) => r.version)) : 0;

  if (revisions.length === 0) {
    return (
      <div className="flex items-center justify-center p-8">
        <p className="text-muted-foreground">No revisions found</p>
      </div>
    );
  }

  return (
    <div className="flex flex-col gap-3">
      <h2 className="text-lg font-semibold">{t('content.revisions')}</h2>

      <div className="flex flex-col gap-2">
        {revisions.map((revision) => {
          const isCurrent = revision.version === maxVersion;
          return (
            <Card
              key={revision.id}
              data-testid={`revision-${revision.id}`}
              className={`p-4 flex items-start justify-between ${isCurrent ? 'current border-primary bg-primary/5' : ''}`}
            >
              <div className="flex flex-col gap-1">
                <div className="flex items-center gap-2">
                  <Badge variant={isCurrent ? 'default' : 'secondary'}>
                    {t('content.revisionVersion', { version: revision.version })}
                  </Badge>
                  <span className="text-sm font-medium">{revision.editor.display_name}</span>
                </div>
                <p className="text-sm text-muted-foreground">{revision.diff_summary}</p>
                <p className="text-xs text-muted-foreground">{formatDate(revision.created_at)}</p>
              </div>
              {!isCurrent && (
                <Button
                  variant="outline"
                  size="sm"
                  onClick={() => setRollbackTarget(revision)}
                >
                  {t('content.rollback')}
                </Button>
              )}
            </Card>
          );
        })}
      </div>

      {/* Rollback Confirm Dialog */}
      <ConfirmDialog
        open={!!rollbackTarget}
        onOpenChange={(open) => {
          if (!open) setRollbackTarget(null);
        }}
        title={t('content.rollback')}
        description={
          rollbackTarget
            ? t('content.rollbackConfirm', { version: rollbackTarget.version })
            : ''
        }
        onConfirm={() => {
          if (rollbackTarget) {
            rollbackMutation.mutate({ revisionId: rollbackTarget.id });
          }
        }}
        loading={rollbackMutation.isPending}
        variant="warning"
      />
    </div>
  );
}
