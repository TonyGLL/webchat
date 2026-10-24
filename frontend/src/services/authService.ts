import axios from 'axios';
import { z } from 'zod';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

export interface IAuthResponse {
  user: User;
  token: string;
  refreshToken: string;
}

export interface User {
  id: number;
  name: string;
  lastName: string;
  username: string;
  email: string;
  phone: string;
  avatarUrl: string;
  createdAt: Date;
  updatedAt: Date;
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

const LoginSchema = z.object({
  user: z.string().email(),
  password: z.string().min(6),
});

export type RegisterData = z.infer<typeof RegisterSchema>;
export type LoginData = z.infer<typeof LoginSchema>;

export const authService = {
  register: async (data: RegisterData): Promise<{ id: number; email: string }> => {
    const response = await axios.post(`${API_URL}/auth/register`, data);
    return response.data;
  },

  login: async (data: LoginData): Promise<IAuthResponse> => {
    const response = await axios.post(`${API_URL}/auth/login`, data);
    return response.data;
  },

  refreshToken: async (token: string) => {
    const response = await axios.post(`${API_URL}/auth/refresh`, { token });
    return response.data;
  },

  sendVerifyEmail: async (userId: number, email: string) => {
    const response = await axios.post(`${API_URL}/auth/send-verify-email/${userId}`, { email });
    return response.data;
  },

  verifyEmail: async (token: string) => {
    const response = await axios.post(`${API_URL}/auth/verify-email/${encodeURIComponent(token)}`);
    return response.data;
  },
};