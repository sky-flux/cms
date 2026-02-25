import { QueryClientProvider } from '@tanstack/react-query';
import { I18nextProvider } from 'react-i18next';
import { Toaster } from 'sonner';
import { queryClient } from '@/lib/query-client';
import i18n from '@/i18n/config';
import { RevisionHistory } from './RevisionHistory';

interface RevisionsPageProps {
  postId: string;
}

export function RevisionsPage({ postId }: RevisionsPageProps) {
  return (
    <I18nextProvider i18n={i18n}>
      <QueryClientProvider client={queryClient}>
        <div className="p-6">
          <RevisionHistory postId={postId} />
        </div>
        <Toaster />
      </QueryClientProvider>
    </I18nextProvider>
  );
}
