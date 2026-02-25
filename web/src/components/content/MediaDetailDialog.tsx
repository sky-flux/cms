import { useEffect } from 'react';
import { useForm } from 'react-hook-form';
import { useTranslation } from 'react-i18next';
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
import { Separator } from '@/components/ui/separator';
import { FileText, Trash2 } from 'lucide-react';
import type { MediaFileDetail, UpdateMediaDTO } from '@/lib/content-api';

interface MediaDetailDialogProps {
  open: boolean;
  onOpenChange: (open: boolean) => void;
  media: MediaFileDetail | null;
  onSave: (data: UpdateMediaDTO) => void | Promise<void>;
  onDelete: (id: string, force: boolean) => void;
  loading?: boolean;
}

function formatFileSize(bytes: number): string {
  if (bytes < 1024) return `${bytes} B`;
  if (bytes < 1024 * 1024) return `${(bytes / 1024).toFixed(1)} KB`;
  return `${(bytes / (1024 * 1024)).toFixed(1)} MB`;
}

interface MetaFormValues {
  alt_text: string;
  title: string;
}

export function MediaDetailDialog({
  open,
  onOpenChange,
  media,
  onSave,
  onDelete,
  loading = false,
}: MediaDetailDialogProps) {
  const { t } = useTranslation();

  const form = useForm<MetaFormValues>({
    defaultValues: {
      alt_text: media?.alt_text ?? '',
      title: media?.title ?? '',
    },
  });

  useEffect(() => {
    if (open && media) {
      form.reset({
        alt_text: media.alt_text ?? '',
        title: media.title ?? '',
      });
    }
  }, [open, media, form]);

  if (!media) return null;

  const isImage = media.media_type === 'image';
  const hasReferences = media.reference_count > 0;

  const handleSave = form.handleSubmit(async (values) => {
    await onSave({
      alt_text: values.alt_text,
      title: values.title,
    });
  });

  const handleDelete = () => {
    onDelete(media.id, hasReferences);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent className="max-w-2xl">
        <DialogHeader>
          <DialogTitle>{t('content.mediaDetail')}</DialogTitle>
        </DialogHeader>

        <div className="grid gap-6 md:grid-cols-2">
          {/* Preview */}
          <div className="flex items-center justify-center rounded-md border bg-muted p-4">
            {isImage && media.thumbnail_urls?.md ? (
              <img
                src={media.thumbnail_urls.md}
                alt={media.alt_text || media.file_name}
                className="max-h-64 rounded object-contain"
              />
            ) : (
              <FileText className="h-16 w-16 text-muted-foreground" />
            )}
          </div>

          {/* Info */}
          <div className="space-y-3">
            <div>
              <span className="text-sm font-medium text-muted-foreground">{t('content.mediaFileName')}</span>
              <p className="text-sm">{media.file_name}</p>
            </div>
            <div>
              <span className="text-sm font-medium text-muted-foreground">{t('content.mediaType')}</span>
              <p className="text-sm">{media.mime_type}</p>
            </div>
            <div>
              <span className="text-sm font-medium text-muted-foreground">{t('content.mediaFileSize')}</span>
              <p className="text-sm">{formatFileSize(media.file_size)}</p>
            </div>
            {isImage && media.width > 0 && media.height > 0 && (
              <div>
                <span className="text-sm font-medium text-muted-foreground">{t('content.mediaDimensions')}</span>
                <p className="text-sm">{media.width} x {media.height}</p>
              </div>
            )}
          </div>
        </div>

        <Separator />

        {/* Editable fields */}
        <form onSubmit={handleSave} className="space-y-4">
          <div className="space-y-2">
            <Label htmlFor="media-alt-text">{t('content.mediaAltText')}</Label>
            <Input
              id="media-alt-text"
              aria-label={t('content.mediaAltText')}
              {...form.register('alt_text')}
              placeholder="Describe this image"
            />
          </div>
          <div className="space-y-2">
            <Label htmlFor="media-title">{t('content.mediaTitle')}</Label>
            <Input
              id="media-title"
              aria-label={t('content.mediaTitle')}
              {...form.register('title')}
              placeholder="File title"
            />
          </div>

          {/* Referencing posts */}
          {media.referencing_posts.length > 0 && (
            <div className="space-y-2">
              <span className="text-sm font-medium">{t('content.mediaReferences')}</span>
              <ul className="space-y-1">
                {media.referencing_posts.map((post) => (
                  <li key={post.id} className="text-sm text-muted-foreground">
                    {post.title}
                  </li>
                ))}
              </ul>
            </div>
          )}

          <DialogFooter className="gap-2">
            <Button
              type="button"
              variant="destructive"
              size="sm"
              onClick={handleDelete}
              aria-label="Delete File"
            >
              <Trash2 className="h-4 w-4 mr-1" />
              {t('content.deleteMedia')}
            </Button>
            <div className="flex-1" />
            <Button type="submit" disabled={loading}>
              {loading ? t('common.loading') : t('common.save')}
            </Button>
          </DialogFooter>
        </form>
      </DialogContent>
    </Dialog>
  );
}
