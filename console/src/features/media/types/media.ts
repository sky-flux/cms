export interface MediaFile {
  id: string;
  filename: string;
  url: string;
  size: number;
  mimeType: string;
  width?: number;
  height?: number;
  alt?: string;
  caption?: string;
  siteId: string;
  folderId?: string;
  createdAt: string;
  updatedAt: string;
}

export interface UploadMediaRequest {
  file: File;
  alt?: string;
  caption?: string;
  folderId?: string;
}

export interface MediaListParams {
  page?: number;
  pageSize?: number;
  folderId?: string;
  mimeType?: string;
  search?: string;
}
