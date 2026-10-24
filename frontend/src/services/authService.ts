import { z } from 'zod';
import { api } from '@/lib/api';
import { GlobalRole } from '@/lib/permissions';

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

export const RegisterSchema = z.object({
  name: z.string().min(1, { message: 'Name is required' }),
  last_name: z.string().min(1, { message: 'Last name is required' }),
  username: z
    .string()
    .min(3, { message: 'Username must be at least 3 characters' })
    .max(30, { message: 'Username must be at most 30 characters' }),
  email: z.string().email({ message: 'Invalid email address' }),
  phone: z.string().min(1, { message: 'Phone is required' }),
  password: z
    .string()
    .min(8, { message: 'Password must be at least 8 characters' })
    .max(72, { message: 'Password must be at most 72 characters' }),
});

export const LoginSchema = z.object({
  user: z.string().email(),
  password: z.string().min(6),
});

export type RegisterData = z.infer<typeof RegisterSchema>;
export type LoginData = z.infer<typeof LoginSchema>;

export const authService = {
  register: async (data: RegisterData): Promise<{ id: number; email: string }> => {
    const response = await api.post('/auth/register', data);
    return response.data;
  },

  login: async (data: LoginData): Promise<IAuthResponse> => {
    const response = await api.post('/auth/login', data);
    return response.data;
  },

  sendVerifyEmail: async (userId: number, email: string) => {
    const response = await api.post(`/auth/send-verify-email/${userId}`, { email });
    return response.data;
  },

  logout: async (): Promise<void> => {
    await api.post('/auth/logout');
  },

  verifyEmail: async (token: string) => {
    const response = await api.post(`/auth/verify-email/${encodeURIComponent(token)}`);
    return response.data;
  },
};
