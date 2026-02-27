import { useCallback, useState } from 'react';
import { useDropzone } from 'react-dropzone';
import { Upload, X, Image, FileText, Film, Music } from 'lucide-react';

interface MediaUploaderProps {
  onUpload: (files: File[]) => void;
  isUploading?: boolean;
  accept?: Record<string, string[]>;
  maxFiles?: number;
  maxSize?: number;
}

function getFileIcon(file: File) {
  if (file.type.startsWith('image/')) return Image;
  if (file.type.startsWith('video/')) return Film;
  if (file.type.startsWith('audio/')) return Music;
  return FileText;
}

function formatFileSize(bytes: number): string {
  if (bytes === 0) return '0 B';
  const k = 1024;
  const sizes = ['B', 'KB', 'MB', 'GB'];
  const i = Math.floor(Math.log(bytes) / Math.log(k));
  return `${parseFloat((bytes / Math.pow(k, i)).toFixed(1))} ${sizes[i]}`;
}

export function MediaUploader({
  onUpload,
  isUploading = false,
  accept = {
    'image/*': ['.png', '.jpg', '.jpeg', '.gif', '.webp', '.svg'],
    'video/*': ['.mp4', '.webm', '.mov'],
    'audio/*': ['.mp3', '.wav', '.ogg'],
  },
  maxFiles = 10,
  maxSize = 10 * 1024 * 1024, // 10MB
}: MediaUploaderProps) {
  const [previews, setPreviews] = useState<(File & { preview: string })[]>([]);

  const onDrop = useCallback(
    (acceptedFiles: File[]) => {
      const filesWithPreviews = acceptedFiles.map((file) => {
        const preview = file.type.startsWith('image/')
          ? URL.createObjectURL(file)
          : '';
        return Object.assign(file, { preview });
      });
      setPreviews((prev) => [...prev, ...filesWithPreviews]);
    },
    []
  );

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop,
    accept,
    maxFiles,
    maxSize,
    disabled: isUploading,
  });

  const handleUpload = () => {
    if (previews.length > 0) {
      onUpload(previews.map((p) => p));
      setPreviews([]);
    }
  };

  const removePreview = (index: number) => {
    setPreviews((prev) => {
      const file = prev[index];
      if (file.preview) {
        URL.revokeObjectURL(file.preview);
      }
      return prev.filter((_, i) => i !== index);
    });
  };

  return (
    <div className="space-y-4">
      <div
        {...getRootProps()}
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
          isDragActive
            ? 'border-primary bg-primary/5'
            : 'border-border hover:border-muted-foreground'
        } ${isUploading ? 'opacity-50 pointer-events-none' : ''}`}
      >
        <input {...getInputProps()} />
        <Upload className="mx-auto h-10 w-10 text-muted-foreground mb-4" />
        {isDragActive ? (
          <p className="text-sm text-primary">Drop the files here...</p>
        ) : (
          <p className="text-sm text-muted-foreground">
            Drag and drop files here, or click to select files
          </p>
        )}
        <p className="text-xs text-muted-foreground mt-2">
          Supports images, videos, and audio files. Max {maxFiles} files, up to{' '}
          {formatFileSize(maxSize)} each.
        </p>
      </div>

      {previews.length > 0 && (
        <div className="space-y-4">
          <div className="grid grid-cols-2 sm:grid-cols-3 md:grid-cols-4 gap-4">
            {previews.map((file, index) => {
              const FileIcon = getFileIcon(file);
              return (
                <div key={index} className="relative group border rounded-lg overflow-hidden">
                  <div className="h-24 bg-muted/50 flex items-center justify-center">
                    {file.preview ? (
                      <img
                        src={file.preview}
                        alt={file.name}
                        className="max-h-full max-w-full object-contain"
                      />
                    ) : (
                      <FileIcon className="h-8 w-8 text-muted-foreground" />
                    )}
                  </div>
                  <div className="p-2">
                    <p className="text-xs truncate" title={file.name}>
                      {file.name}
                    </p>
                    <p className="text-xs text-muted-foreground">
                      {formatFileSize(file.size)}
                    </p>
                  </div>
                  <button
                    onClick={() => removePreview(index)}
                    className="absolute top-1 right-1 p-1 bg-background/90 rounded-full opacity-0 group-hover:opacity-100 transition-opacity hover:bg-destructive hover:text-destructive-foreground"
                    disabled={isUploading}
                  >
                    <X className="h-3 w-3" />
                  </button>
                </div>
              );
            })}
          </div>
          <div className="flex justify-end gap-2">
            <button
              onClick={() => setPreviews([])}
              className="px-4 py-2 text-sm border rounded-md hover:bg-muted"
              disabled={isUploading}
            >
              Clear All
            </button>
            <button
              onClick={handleUpload}
              className="px-4 py-2 text-sm bg-primary text-primary-foreground rounded-md hover:bg-primary/90"
              disabled={isUploading}
            >
              {isUploading ? 'Uploading...' : `Upload ${previews.length} file${previews.length > 1 ? 's' : ''}`}
            </button>
          </div>
        </div>
      )}
    </div>
  );
}
