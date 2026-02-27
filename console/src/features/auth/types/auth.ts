export interface LoginRequest {
  email: string;
  password: string;
}

export interface LoginResponse {
  token: string;
  user: {
    id: string;
    email: string;
    name: string;
  };
  requires?: 'totp';
  tempToken?: string;
}

export interface VerifyTOTPRequest {
  tempToken: string;
  code: string;
}

export interface ForgotPasswordRequest {
  email: string;
}

export interface ResetPasswordRequest {
  token: string;
  password: string;
}

export interface MeResponse {
  id: string;
  email: string;
  name: string;
  avatar?: string;
  role: string;
  siteIds: string[];
}
