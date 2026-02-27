import { ApiError, type RequestOptions } from './types/api-client';

const API_BASE_URL = import.meta.env.VITE_API_URL || 'http://localhost:8080/api';

class ApiClient {
  private baseURL: string;

  constructor(baseURL: string = API_BASE_URL) {
    this.baseURL = baseURL;
  }

  private buildURL(endpoint: string, params?: Record<string, string | number | boolean>): string {
    const url = new URL(`${this.baseURL}${endpoint}`);
    if (params) {
      Object.entries(params).forEach(([key, value]) => {
        url.searchParams.append(key, String(value));
      });
    }
    return url.toString();
  }

  private async handleResponse<T>(response: Response): Promise<T> {
    if (!response.ok) {
      const errorData = await response.json().catch(() => ({}));
      throw new ApiError(
        errorData.message || `HTTP error ${response.status}`,
        response.status,
        errorData.code
      );
    }
    return response.json();
  }

  async get<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'GET',
      credentials: 'include',
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async post<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'POST',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async put<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'PUT',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async patch<T>(endpoint: string, data?: unknown, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'PATCH',
      credentials: 'include',
      headers: { 'Content-Type': 'application/json' },
      body: data ? JSON.stringify(data) : undefined,
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async delete<T>(endpoint: string, options?: RequestOptions): Promise<T> {
    const response = await fetch(this.buildURL(endpoint, options?.params), {
      method: 'DELETE',
      credentials: 'include',
      ...options,
    });
    return this.handleResponse<T>(response);
  }

  async upload<T>(endpoint: string, formData: FormData): Promise<T> {
    const response = await fetch(this.buildURL(endpoint), {
      method: 'POST',
      credentials: 'include',
      body: formData,
    });
    return this.handleResponse<T>(response);
  }
}

export const apiClient = new ApiClient();
export { ApiError };
