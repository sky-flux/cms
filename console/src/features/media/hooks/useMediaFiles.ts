import { createQuery } from '@tanstack/react-query';
import { apiClient, type ListResponse } from '../../shared';
import type { MediaFile, MediaListParams } from '../types/media';

export function useMediaFiles(siteSlug: string, params: MediaListParams = {}) {
  return createQuery({
    queryKey: ['media', siteSlug, params],
    queryFn: async () => {
      const response = await apiClient.get<ListResponse<MediaFile>>(
        `/sites/${siteSlug}/media`,
        { params }
      );
      return response;
    },
  });
}
