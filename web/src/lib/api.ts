const API_BASE_URL = import.meta.env.PUBLIC_API_URL || '/api';

export interface APIResponse<T> {
  data: T;
  message?: string;
}

interface APIError {
  code: number;
  message: string;
}

export async function fetchAPI<T>(endpoint: string, options: RequestInit = {}): Promise<T> {
  const url = `${API_BASE_URL}${endpoint}`;

  const response = await fetch(url, {
    headers: {
      'Content-Type': 'application/json',
      ...options.headers,
    },
    ...options,
  });

  if (!response.ok) {
    const error: APIError = await response.json().catch(() => ({
      code: response.status,
      message: response.statusText,
    }));
    throw new Error(error.message);
  }

  return response.json();
}
