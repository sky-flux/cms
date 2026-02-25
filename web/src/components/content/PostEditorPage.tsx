import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { Toaster } from 'sonner';
import { queryClient } from '@/lib/query-client';
import i18n from '@/i18n/config';
import { PostEditor } from './PostEditor';

interface PostEditorPageProps {
  mode: 'create' | 'edit';
  postId?: string;
}

export function PostEditorPage({ mode, postId }: PostEditorPageProps) {
  const handleCreated = (newPostId: string) => {
    window.location.href = `/dashboard/posts/${newPostId}/edit`;
  };

  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <div className="p-6">
          <PostEditor mode={mode} postId={postId} onCreated={handleCreated} />
        </div>
        <Toaster />
      </QueryClientProvider>
    </I18nextProvider>
  );
}
