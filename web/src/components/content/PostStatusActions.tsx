import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import { Button } from '@/components/ui/button';
import { postsApi } from '@/lib/content-api';

interface PostStatusActionsProps {
  postId: string;
  status: string;
  onStatusChange: () => void;
}

export function PostStatusActions({
  postId,
  status,
  onStatusChange,
}: PostStatusActionsProps) {
  const { t } = useTranslation();
  const [loading, setLoading] = useState(false);

  async function handleAction(
    action: (id: string) => Promise<unknown>,
  ) {
    setLoading(true);
    try {
      await action(postId);
      onStatusChange();
    } finally {
      setLoading(false);
    }
  }

  if (status === 'draft') {
    return (
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          disabled={loading}
          onClick={() => handleAction(postsApi.publish)}
          aria-label={t('content.publish')}
        >
          {t('content.publish')}
        </Button>
      </div>
    );
  }

  if (status === 'published') {
    return (
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          variant="outline"
          disabled={loading}
          onClick={() => handleAction(postsApi.unpublish)}
          aria-label={t('content.unpublish')}
        >
          {t('content.unpublish')}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          disabled={loading}
          onClick={() => handleAction(postsApi.revertToDraft)}
          aria-label={t('content.revertToDraft')}
        >
          {t('content.revertToDraft')}
        </Button>
      </div>
    );
  }

  if (status === 'scheduled') {
    return (
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          disabled={loading}
          onClick={() => handleAction(postsApi.publish)}
          aria-label={t('content.publish')}
        >
          {t('content.publish')}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          disabled={loading}
          onClick={() => handleAction(postsApi.revertToDraft)}
          aria-label={t('content.revertToDraft')}
        >
          {t('content.revertToDraft')}
        </Button>
      </div>
    );
  }

  if (status === 'archived') {
    return (
      <div className="flex items-center gap-2">
        <Button
          size="sm"
          disabled={loading}
          onClick={() => handleAction(postsApi.restore)}
          aria-label={t('content.restore')}
        >
          {t('content.restore')}
        </Button>
        <Button
          size="sm"
          variant="ghost"
          disabled={loading}
          onClick={() => handleAction(postsApi.revertToDraft)}
          aria-label={t('content.revertToDraft')}
        >
          {t('content.revertToDraft')}
        </Button>
      </div>
    );
  }

  return null;
}
