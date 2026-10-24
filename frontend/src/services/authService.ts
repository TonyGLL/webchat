import { z } from 'zod';
import { api } from '@/lib/api';
import { GlobalRole } from '@/lib/permissions';
import type { Translate } from '@/lib/i18n';

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

export const makeRegisterSchema = (t: Translate) =>
  z.object({
    name: z.string().min(1, { message: t('validation.nameRequired') }),
    last_name: z.string().min(1, { message: t('validation.lastNameRequired') }),
    username: z
      .string()
      .min(3, { message: t('validation.usernameMin', { min: 3 }) })
      .max(30, { message: t('validation.usernameMax', { max: 30 }) })
      .regex(/^[a-zA-Z0-9_.-]+$/, { message: t('validation.usernameChars') }),
    email: z.string().email({ message: t('validation.invalidEmail') }),
    phone: z.string().min(1, { message: t('validation.phoneRequired') }),
    password: z
      .string()
      .min(8, { message: t('validation.passwordMin', { min: 8 }) })
      .max(72, { message: t('validation.passwordMax', { max: 72 }) }),
  });

export const makeLoginSchema = (t: Translate) =>
  z.object({
    user: z.string().email({ message: t('validation.invalidEmail') }),
    password: z.string().min(6, { message: t('validation.passwordMin', { min: 6 }) }),
  });

export const makeForgotPasswordSchema = (t: Translate) =>
  z.object({ email: z.string().email({ message: t('validation.invalidEmail') }) });

export const makeResetPasswordSchema = (t: Translate) =>
  z
    .object({
      password: z
        .string()
        .min(8, { message: t('validation.passwordMin', { min: 8 }) })
        .max(72, { message: t('validation.passwordMax', { max: 72 }) }),
      confirm: z.string(),
    })
    .refine((v) => v.password === v.confirm, { path: ['confirm'], message: t('validation.passwordMatch') });

export type ForgotPasswordData = z.infer<ReturnType<typeof makeForgotPasswordSchema>>;
export type ResetPasswordData = z.infer<ReturnType<typeof makeResetPasswordSchema>>;
export type RegisterData = z.infer<ReturnType<typeof makeRegisterSchema>>;
export type LoginData = z.infer<ReturnType<typeof makeLoginSchema>>;

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

  verifyEmail: async (token: string) => {
    const response = await api.post(`/auth/verify-email/${encodeURIComponent(token)}`);
    return response.data;
  },
};
