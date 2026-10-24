import { afterEach, describe, expect, it, vi } from 'vitest';
import { describeError, logError } from './logError';

const axiosLike = {
  message: 'Request failed with status code 401',
  response: { status: 401, data: { error: 'Invalid credentials' } },
  config: {
    url: 'http://localhost:8081/api/v1/auth/login?token=secret-token',
    method: 'post',
    data: '{"email":"a@b.c","password":"hunter2-password"}',
    headers: { Authorization: 'Bearer eyJ.secret.jwt' },
  },
};

describe('describeError', () => {
  it('keeps message, status and endpoint path only', () => {
    expect(describeError(axiosLike)).toBe('Request failed with status code 401 | status 401 | POST /api/v1/auth/login');
  });

  it('never includes request data, headers or query strings', () => {
    const out = describeError(axiosLike);
    for (const secret of ['hunter2', 'Bearer', 'eyJ', 'secret-token', 'a@b.c']) expect(out).not.toContain(secret);
  });

  it('handles plain errors and odd values', () => {
    expect(describeError(new Error('boom'))).toBe('boom');
    expect(describeError('oops')).toBe('oops');
    expect(describeError(null)).toBe('unknown error');
    expect(describeError({ config: { url: '/rooms?x=1' } })).toBe('unknown error | /rooms');
  });
});

describe('logError', () => {
  afterEach(() => vi.restoreAllMocks());

  it('logs a single safe string', () => {
    const spy = vi.spyOn(console, 'error').mockImplementation(() => {});
    logError('Login failed', axiosLike);
    expect(spy).toHaveBeenCalledTimes(1);
    expect(spy.mock.calls[0]).toHaveLength(1);
    expect(String(spy.mock.calls[0][0])).not.toContain('hunter2');
  });
});
