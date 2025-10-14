import axios from 'axios';
import { z } from 'zod';

const API_URL = process.env.NEXT_PUBLIC_API_URL;

const RegisterSchema = z.object({
  username: z.string().min(3),
  email: z.string().email(),
  password: z.string().min(6),
});

const LoginSchema = z.object({
  email: z.string().email(),
  password: z.string().min(6),
});

export type RegisterData = z.infer<typeof RegisterSchema>;
export type LoginData = z.infer<typeof LoginSchema>;

export const authService = {
  register: async (data: RegisterData) => {
    const response = await axios.post(`${API_URL}/auth/register`, data);
    return response.data;
  },

  login: async (data: LoginData) => {
    const response = await axios.post(`${API_URL}/auth/login`, data);
    return response.data;
  },

  refreshToken: async (token: string) => {
    const response = await axios.post(`${API_URL}/auth/refresh`, { token });
    return response.data;
  },

  verifyEmail: async (token: string) => {
    const response = await axios.get(`${API_URL}/auth/verify/${token}`);
    return response.data;
  },
};