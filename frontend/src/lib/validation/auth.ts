import { z } from 'zod';
import type { Translate } from '@/lib/i18n';

// Zod schemas live apart from authService so that only the auth pages ship zod (the root
// layout imports authService through AuthContext).

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
