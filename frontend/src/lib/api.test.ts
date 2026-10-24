import axios, { AxiosError, InternalAxiosRequestConfig } from 'axios';
import { beforeEach, describe, expect, it, vi } from 'vitest';
import { api } from './api';
import { tokenStorage } from './tokenStorage';

const unauthorized = (config: InternalAxiosRequestConfig) =>
  new AxiosError('Unauthorized', 'ERR_BAD_REQUEST', config, null, {
    status: 401,
    statusText: 'Unauthorized',
    data: {},
    headers: {},
    config,
  });

/** Fake server: only the token "fresh" is accepted. Records the Authorization of each call. */
function fakeServer() {
  const seen: string[] = [];
  api.defaults.adapter = async (config) => {
    const auth = String(config.headers.Authorization ?? '');
    seen.push(auth);
    if (auth !== 'Bearer fresh') throw unauthorized(config);
    return { data: { ok: true }, status: 200, statusText: 'OK', headers: {}, config };
  };
  return seen;
}

describe('api client token refresh', () => {
  beforeEach(() => {
    vi.restoreAllMocks();
    window.localStorage.clear();
    tokenStorage.set({ accessToken: 'expired', refreshToken: 'refresh-1' });
  });

  it('refreshes once on 401 and retries the request with the new token', async () => {
    const seen = fakeServer();
    const post = vi
      .spyOn(axios, 'post')
      .mockResolvedValue({ data: { token: 'fresh', refreshToken: 'refresh-2' } });

    const response = await api.get('/rooms');

    expect(response.data).toEqual({ ok: true });
    expect(post).toHaveBeenCalledTimes(1);
    expect(post.mock.calls[0][1]).toEqual({ refreshToken: 'refresh-1' });
    expect(seen).toEqual(['Bearer expired', 'Bearer fresh']);
    expect(tokenStorage.get()).toEqual({ accessToken: 'fresh', refreshToken: 'refresh-2' });
  });

  it('shares a single refresh between concurrent 401s (the refresh token is single-use)', async () => {
    fakeServer();
    const post = vi
      .spyOn(axios, 'post')
      .mockResolvedValue({ data: { token: 'fresh', refreshToken: 'refresh-2' } });

    await Promise.all([api.get('/rooms'), api.get('/users/profile'), api.get('/messages/x/messages')]);

    expect(post).toHaveBeenCalledTimes(1);
  });

  it('ends the session when the refresh fails', async () => {
    fakeServer();
    vi.spyOn(axios, 'post').mockRejectedValue(new Error('refresh rejected'));

    await expect(api.get('/rooms')).rejects.toThrow();

    expect(tokenStorage.get()).toBeNull();
  });

  it('does not try to refresh failed /auth/ calls (e.g. wrong password)', async () => {
    fakeServer();
    const post = vi.spyOn(axios, 'post');

    await expect(api.post('/auth/login', {})).rejects.toBeInstanceOf(AxiosError);

    expect(post).not.toHaveBeenCalled();
    expect(tokenStorage.get()).not.toBeNull();
  });
});
