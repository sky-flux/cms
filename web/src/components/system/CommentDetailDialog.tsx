import { useState } from 'react';
import { useTranslation } from 'react-i18next';
import {
  Dialog,
  DialogContent,
  DialogDescription,
  DialogHeader,
  DialogTitle,
} from '@/components/ui/dialog';
import { Button } from '@/components/ui/button';
import { StatusBadge } from '@/components/shared/StatusBadge';
import type { Comment } from '@/lib/system-api';

interface CommentDetailDialogProps {
  comment: Comment | null;
  open: boolean;
  onOpenChange: (open: boolean) => void;
  onReply: (commentId: string, content: string) => void;
  replyLoading: boolean;
}

function ReplyTree({ replies, level = 1 }: { replies: Comment[]; level?: number }) {
  if (level > 3) return null;
  return (
    <div className="space-y-3">
      {replies.map((reply) => (
        <div
          key={reply.id}
          className="border-l-2 border-muted pl-4"
          style={{ marginLeft: `${(level - 1) * 8}px` }}
        >
          <div className="flex items-center gap-2 text-sm">
            <span className="font-medium">{reply.author_name}</span>
            <span className="text-muted-foreground text-xs">
              {new Date(reply.created_at).toLocaleDateString('en-US', {
                year: 'numeric',
                month: 'short',
                day: 'numeric',
              })}
            </span>
          </div>
          <p className="mt-1 text-sm">{reply.content}</p>
          {reply.replies && reply.replies.length > 0 && (
            <div className="mt-2">
              <ReplyTree replies={reply.replies} level={level + 1} />
            </div>
          )}
        </div>
      ))}
    </div>
  );
}

export function CommentDetailDialog({
  comment,
  open,
  onOpenChange,
  onReply,
  replyLoading,
}: CommentDetailDialogProps) {
  const { t } = useTranslation();
  const [replyContent, setReplyContent] = useState('');

  if (!comment) return null;

  function handleSubmitReply() {
    if (!comment || !replyContent.trim()) return;
    onReply(comment.id, replyContent.trim());
    setReplyContent('');
  }

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl max-h-[80vh] overflow-y-auto">
        <DialogHeader>
          <DialogTitle>{t('system.comments.commentDetail')}</DialogTitle>
          <DialogDescription className="sr-only">
            {t('system.comments.commentDetail')}
          </DialogDescription>
        </DialogHeader>

        {/* Author info */}
        <div className="space-y-2 rounded-md border p-4">
          <div className="flex items-center gap-3">
            <span className="font-medium text-lg">{comment.author_name}</span>
            <StatusBadge status={comment.status} />
          </div>
          <div className="grid grid-cols-2 gap-2 text-sm text-muted-foreground">
            <div>
              <span className="font-medium">{t('common.email')}: </span>
              {comment.author_email}
            </div>
            <div>
              <span className="font-medium">{t('system.audit.ipAddress')}: </span>
              {comment.author_ip}
            </div>
          </div>
        </div>

        {/* Full comment content */}
        <div className="rounded-md border p-4">
          <p className="text-sm leading-relaxed">{comment.content}</p>
        </div>

        {/* Replies tree */}
        {comment.replies && comment.replies.length > 0 && (
          <div className="space-y-2">
            <h4 className="font-medium">{t('system.comments.replies')}</h4>
            <ReplyTree replies={comment.replies} />
          </div>
        )}

        {/* Admin reply form */}
        <div className="space-y-2 border-t pt-4">
          <h4 className="font-medium">{t('system.comments.adminReply')}</h4>
          <textarea
            className="w-full rounded-md border border-input bg-background px-3 py-2 text-sm placeholder:text-muted-foreground focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-ring"
            rows={3}
            placeholder={t('system.comments.replyPlaceholder')}
            value={replyContent}
            onChange={(e) => setReplyContent(e.target.value)}
          />
          <div className="flex justify-end">
            <Button
              size="sm"
              onClick={handleSubmitReply}
              disabled={!replyContent.trim() || replyLoading}
              aria-label={t('system.comments.reply')}
            >
              {replyLoading ? t('common.loading') : t('system.comments.reply')}
            </Button>
          </div>
        </div>
      </DialogContent>
    </Dialog>
  );
}
