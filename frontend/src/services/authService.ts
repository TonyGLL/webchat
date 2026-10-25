import { api } from '@/lib/api';
import { GlobalRole } from '@/lib/permissions';
import type { LoginData, RegisterData } from '@/lib/validation/auth';

export interface User {
  id: number;
  name: string;
  last_name: string;
  username: string;
  email: string;
  avatar_url: string | null;
  /** Application-wide role; only the server can grant moderator/admin. */
  role: GlobalRole;
  created_at: string;
}

export interface IAuthResponse {
  user: { id: number; email: string };
  token: string;
}

export type { ForgotPasswordData, LoginData, RegisterData, ResetPasswordData } from '@/lib/validation/auth';

export const authService = {
  register: async (data: RegisterData): Promise<{ id: number; email: string }> => {
    const response = await api.post('/auth/register', data);
    return response.data;
  },

  login: async (data: LoginData): Promise<IAuthResponse> => {
    const response = await api.post('/auth/login', data);
    return response.data;
  },

  /** Always resolves the same way, whether or not the address has an account. */
  forgotPassword: async (email: string): Promise<void> => {
    await api.post('/auth/forgot-password', { email });
  },

  resetPassword: async (token: string, password: string): Promise<void> => {
    await api.post('/auth/reset-password', { token, password });
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
  },
};
