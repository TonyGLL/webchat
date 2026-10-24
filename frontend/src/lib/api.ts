import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { tokenStorage } from './tokenStorage';

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8081/api/v1';

/** Shared axios instance: attaches the access token and renews it once when it expires. */
export const api = axios.create({ baseURL: API_URL });

type RetriableConfig = InternalAxiosRequestConfig & { _retried?: boolean };

interface RefreshResponse {
  token: string;
  refreshToken: string;
}

// Concurrent 401s share a single refresh call (the refresh token is single-use).
let refreshing: Promise<string> | null = null;

async function refreshAccessToken(): Promise<string> {
  const current = tokenStorage.get();
  if (!current?.refreshToken) throw new Error('No refresh token');

  const { data } = await axios.post<RefreshResponse>(`${API_URL}/auth/refresh-token`, {
    refreshToken: current.refreshToken,
  });
  tokenStorage.set({ accessToken: data.token, refreshToken: data.refreshToken });
  return data.token;
}

api.interceptors.request.use((config) => {
  const tokens = tokenStorage.get();
  if (tokens) {
    config.headers.Authorization = `Bearer ${tokens.accessToken}`;
  }
  return config;
});

api.interceptors.response.use(
  (response) => response,
  async (error: AxiosError) => {
    const original = error.config as RetriableConfig | undefined;
    const isAuthRoute = original?.url?.startsWith('/auth/');

    if (error.response?.status !== 401 || !original || original._retried || isAuthRoute) {
      return Promise.reject(error);
    }
    original._retried = true;

    try {
      refreshing ??= refreshAccessToken().finally(() => {
        refreshing = null;
      });
      const newToken = await refreshing;
      original.headers.Authorization = `Bearer ${newToken}`;
      return api(original);
    } catch (refreshError) {
      tokenStorage.clear(); // session is over; AuthContext reacts and sends the user to /login
      return Promise.reject(refreshError);
    }
  }
);
