import { useMutation, useQueryClient } from '@tanstack/react-query';
import { apiClient } from '../../shared';
import type { MediaFile, UploadMediaRequest } from '../types/media';

export function useUploadMedia(siteSlug: string) {
  const queryClient = useQueryClient();
  return useMutation({
    mutationFn: async (data: UploadMediaRequest): Promise<MediaFile> => {
      const formData = new FormData();
      formData.append('file', data.file);
      if (data.alt) {
        formData.append('alt', data.alt);
      }
      if (data.caption) {
        formData.append('caption', data.caption);
      }
      if (data.folderId) {
        formData.append('folderId', data.folderId);
      }
      const response = await apiClient.upload<MediaFile>(
        `/sites/${siteSlug}/media`,
        formData
      );
      return response;
    },
    onSuccess: () => {
      queryClient.invalidateQueries({ queryKey: ['media', siteSlug] });
    },
  });
}
