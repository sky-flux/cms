// Detect if running in Docker production environment
const isDockerProduction = () => {
  try {
    // Check if we're in a container by looking for Docker indicators
    return typeof window === 'undefined' &&
           process.env.NODE_ENV === 'production';
  } catch {
    return false;
  }
};

// Use internal API URL for server-side calls in Docker, public URL for browser
const getApiBase = () => {
  // Server-side in Docker production: use internal API URL
  if (isDockerProduction()) {
    return 'http://api:8080'; // Docker container-to-container communication
  }
  // Server-side in local dev or client-side: use public URL
  return import.meta.env.PUBLIC_API_URL || '/api';
};

const API_BASE = getApiBase();

export interface SetupStatusResponse {
  installed: boolean;
}

export interface InitializeSetupRequest {
  site_name: string;
  super_email: string;
  super_password: string;
  super_name: string;
  db_host?: string;
  db_port?: number;
  db_name?: string;
  db_user?: string;
  db_password?: string;
}

export interface InitializeSetupResponse {
  user_id: string;
  site_id: string;
}

/**
 * Check installation status
 */
export async function fetchSetupStatus(): Promise<boolean> {
  const response = await fetch(`${API_BASE}/v1/setup/check`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
  });

  if (!response.ok) {
    throw new Error('Failed to check setup status');
  }

  const result = await response.json();
  return result.data.installed;
}

/**
 * Initialize the CMS (first-time setup)
 */
export async function initializeSetup(
  data: InitializeSetupRequest
): Promise<InitializeSetupResponse> {
  const response = await fetch(`${API_BASE}/v1/setup/initialize`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/json' },
    credentials: 'include',
    body: JSON.stringify(data),
  });

  if (!response.ok) {
    const error = await response.json();
    throw new Error(error.message || 'Failed to initialize setup');
  }

  const result = await response.json();
  return result.data;
}
