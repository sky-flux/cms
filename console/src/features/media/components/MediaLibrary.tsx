import { FileImage, FileVideo, FileAudio, File, Trash2, Download } from 'lucide-react';
import type { MediaFile } from '../types/media';

interface MediaLibraryProps {
  mediaFiles: MediaFile[];
  onDelete: (mediaId: string) => void;
  onSelect?: (mediaFile: MediaFile) => void;
  selectedIds?: string[];
}

function getFileIcon(mimeType: string) {
  if (mimeType.startsWith('image/')) return FileImage;
  if (mimeType.startsWith('video/')) return FileVideo;
  if (mimeType.startsWith('audio/')) return FileAudio;
  return File;
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

function isImage(mimeType: string): boolean {
  return mimeType.startsWith('image/');
}

export function MediaLibrary({ mediaFiles, onDelete, onSelect, selectedIds = [] }: MediaLibraryProps) {
  if (mediaFiles.length === 0) {
    return (
      <div className="text-center py-12 text-muted-foreground">
        No media files found. Upload some files to get started.
      </div>
    );
  }

  return (
    <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 lg:grid-cols-5 gap-4">
      {mediaFiles.map((media) => {
        const isSelected = selectedIds.includes(media.id);
        const FileIcon = getFileIcon(media.mimeType);

        return (
          <div
            key={media.id}
            className={`group relative border rounded-lg overflow-hidden hover:shadow-md transition-shadow ${
              isSelected ? 'ring-2 ring-primary' : ''
            }`}
          >
            <button
              onClick={() => onSelect?.(media)}
              className="block w-full h-32 bg-muted/50 flex items-center justify-center"
            >
              {isImage(media.mimeType) ? (
                <img
                  src={media.url}
                  alt={media.alt || media.filename}
                  className="max-h-full max-w-full object-contain"
                />
              ) : (
                <FileIcon className="h-12 w-12 text-muted-foreground" />
              )}
            </button>
            <div className="p-2">
              <p className="text-sm font-medium truncate" title={media.filename}>
                {media.filename}
              </p>
              <p className="text-xs text-muted-foreground">
                {formatFileSize(media.size)}
              </p>
            </div>
            <div className="absolute top-1 right-1 opacity-0 group-hover:opacity-100 transition-opacity flex gap-1">
              {isImage(media.mimeType) && (
                <a
                  href={media.url}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="p-1 bg-background/90 rounded hover:bg-background"
                  title="Download"
                  onClick={(e) => e.stopPropagation()}
                >
                  <Download className="h-3 w-3" />
                </a>
              )}
              <button
                onClick={(e) => {
                  e.stopPropagation();
                  onDelete(media.id);
                }}
                className="p-1 bg-background/90 rounded hover:bg-destructive hover:text-destructive-foreground"
                title="Delete"
              >
                <Trash2 className="h-3 w-3" />
              </button>
            </div>
          </div>
        );
      })}
    </div>
  );
}
