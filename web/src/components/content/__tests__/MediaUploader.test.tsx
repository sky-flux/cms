import { describe, it, expect, vi } from 'vitest';
import { render, screen } from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import { MediaUploader } from '../MediaUploader';

vi.mock('react-i18next', () => ({
  useTranslation: () => ({
    t: (key: string, opts?: Record<string, unknown>) => {
      const map: Record<string, string> = {
        'content.dropzone': 'Drop files here or click to upload',
        'content.uploading': 'Uploading...',
        'content.uploadMedia': 'Upload Files',
      };
      if (key === 'content.uploadProgress' && opts?.percent !== undefined) {
        return `${opts.percent}%`;
      }
      return map[key] ?? key;
    },
  }),
}));

vi.mock('react-dropzone', () => ({
  useDropzone: (opts: any) => ({
    getRootProps: () => ({ onClick: vi.fn(), 'data-testid': 'dropzone' }),
    getInputProps: () => ({ 'data-testid': 'dropzone-input' }),
    isDragActive: false,
    open: opts?.onDrop ? () => opts.onDrop([]) : vi.fn(),
  }),
}));

describe('MediaUploader', () => {
  it('renders dropzone area with text', () => {
    render(
      <MediaUploader onUpload={vi.fn()} />,
    );
    expect(screen.getByText('Drop files here or click to upload')).toBeInTheDocument();
  });

  it('renders upload button', () => {
    render(
      <MediaUploader onUpload={vi.fn()} />,
    );
    expect(screen.getByTestId('dropzone')).toBeInTheDocument();
  });

  it('shows uploading status when files are in progress', () => {
    render(
      <MediaUploader
        onUpload={vi.fn()}
        uploadingFiles={[
          { name: 'test.jpg', progress: 50 },
        ]}
      />,
    );
    expect(screen.getByText('test.jpg')).toBeInTheDocument();
    expect(screen.getByText('50%')).toBeInTheDocument();
  });

  it('shows multiple uploading files', () => {
    render(
      <MediaUploader
        onUpload={vi.fn()}
        uploadingFiles={[
          { name: 'test1.jpg', progress: 30 },
          { name: 'test2.pdf', progress: 70 },
        ]}
      />,
    );
    expect(screen.getByText('test1.jpg')).toBeInTheDocument();
    expect(screen.getByText('test2.pdf')).toBeInTheDocument();
    expect(screen.getByText('30%')).toBeInTheDocument();
    expect(screen.getByText('70%')).toBeInTheDocument();
  });

  it('does not show upload list when no files uploading', () => {
    render(
      <MediaUploader onUpload={vi.fn()} />,
    );
    // Should only show dropzone, no progress bars
    expect(screen.queryByRole('progressbar')).not.toBeInTheDocument();
  });

  it('renders with custom className', () => {
    const { container } = render(
      <MediaUploader onUpload={vi.fn()} className="my-custom-class" />,
    );
    expect(container.firstChild).toHaveClass('my-custom-class');
  });
});
