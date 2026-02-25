import { useTranslation } from 'react-i18next';
import { useDropzone } from 'react-dropzone';
import { Upload } from 'lucide-react';

export interface UploadingFile {
  name: string;
  progress: number;
}

interface MediaUploaderProps {
  onUpload: (files: File[]) => void;
  uploadingFiles?: UploadingFile[];
  className?: string;
}

export function MediaUploader({
  onUpload,
  uploadingFiles = [],
  className = '',
}: MediaUploaderProps) {
  const { t } = useTranslation();

  const { getRootProps, getInputProps, isDragActive } = useDropzone({
    onDrop: (acceptedFiles) => {
      if (acceptedFiles.length > 0) {
        onUpload(acceptedFiles);
      }
    },
    accept: {
      'image/*': [],
      'video/*': [],
      'application/pdf': [],
      'application/msword': [],
      'application/vnd.openxmlformats-officedocument.wordprocessingml.document': [],
    },
    maxSize: 50 * 1024 * 1024, // 50MB
    multiple: true,
  });

  return (
    <div className={className}>
      <div
        {...getRootProps()}
        className={`border-2 border-dashed rounded-lg p-8 text-center cursor-pointer transition-colors ${
          isDragActive
            ? 'border-primary bg-primary/5'
            : 'border-muted-foreground/25 hover:border-primary/50'
        }`}
      >
        <input {...getInputProps()} />
        <Upload className="h-8 w-8 mx-auto mb-3 text-muted-foreground" />
        <p className="text-sm text-muted-foreground">
          {t('content.dropzone')}
        </p>
      </div>

      {uploadingFiles.length > 0 && (
        <div className="mt-4 space-y-2">
          {uploadingFiles.map((file, index) => (
            <div key={`${file.name}-${index}`} className="space-y-1">
              <div className="flex items-center justify-between text-sm">
                <span className="truncate">{file.name}</span>
                <span className="text-muted-foreground">
                  {t('content.uploadProgress', { percent: file.progress })}
                </span>
              </div>
              <div className="h-2 rounded-full bg-muted overflow-hidden" role="progressbar" aria-valuenow={file.progress} aria-valuemin={0} aria-valuemax={100}>
                <div
                  className="h-full bg-primary transition-all"
                  style={{ width: `${file.progress}%` }}
                />
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
}
