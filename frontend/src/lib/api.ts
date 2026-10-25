import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { tokenStorage } from './tokenStorage';
import { isSessionRejected } from './sessionRestore';

export const API_URL = process.env.NEXT_PUBLIC_API_URL ?? 'http://localhost:8081/api/v1';

/** A stalled request must not hang forever (and keep a signed-out session's calls alive). */
const REQUEST_TIMEOUT_MS = 20_000;

/** Shared axios instance: attaches the access token and renews it once when it expires. */
export const api = axios.create({ baseURL: API_URL, withCredentials: true, timeout: REQUEST_TIMEOUT_MS });

type RetriableConfig = InternalAxiosRequestConfig & { _retried?: boolean };

interface RefreshResponse {
  token: string;
}

// Concurrent callers (401s, React strict-mode double effects) share a single refresh call:
// the refresh token is single-use, so a second request would be rejected.
let refreshing: Promise<string> | null = null;

/** Exchanges the httpOnly refresh cookie for a new access token. */
export function refreshAccessToken(): Promise<string> {
  const epoch = tokenStorage.epoch();
  refreshing ??= axios
    .post<RefreshResponse>(`${API_URL}/auth/refresh-token`, null, { withCredentials: true, timeout: REQUEST_TIMEOUT_MS })
    .then(({ data }) => {
      // A sign-out while this call was in flight must not bring the session back.
      if (tokenStorage.epoch() !== epoch) throw new Error('Session ended while refreshing');
      tokenStorage.set({ accessToken: data.token });
      return data.token;
    })
    .finally(() => {
      refreshing = null;
    });
  return refreshing;
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
      const newToken = await refreshAccessToken();
      original.headers.Authorization = `Bearer ${newToken}`;
      return api(original);
    } catch (refreshError) {
      // Only a rejected refresh token ends the session (AuthContext then sends the user to
      // /login); a network blip must not sign anyone out.
      if (isSessionRejected(refreshError)) tokenStorage.clear();
      return Promise.reject(refreshError);
    }
  }
);
