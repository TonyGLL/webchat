import { describe, expect, it, vi } from 'vitest';
import { isSessionRejected, restoreSession } from './sessionRestore';

const rejected = Object.assign(new Error('401'), { response: { status: 401 } });
const noSleep = async () => undefined;

describe('restoreSession', () => {
  it('restores on the first try', async () => {
    const refresh = vi.fn().mockResolvedValue('ok');
    expect(await restoreSession(refresh, { sleep: noSleep })).toBe('restored');
    expect(refresh).toHaveBeenCalledTimes(1);
  });

  it('retries a network failure and then succeeds', async () => {
    const refresh = vi.fn().mockRejectedValueOnce(new Error('Network Error')).mockResolvedValue('ok');
    expect(await restoreSession(refresh, { sleep: noSleep })).toBe('restored');
    expect(refresh).toHaveBeenCalledTimes(2);
  });

  it('gives up as unavailable (not rejected) when the server stays unreachable', async () => {
    const refresh = vi.fn().mockRejectedValue(new Error('Network Error'));
    expect(await restoreSession(refresh, { attempts: 3, sleep: noSleep })).toBe('unavailable');
    expect(refresh).toHaveBeenCalledTimes(3);
  });

  it('stops at once when the server rejects the session', async () => {
    const refresh = vi.fn().mockRejectedValue(rejected);
    expect(await restoreSession(refresh, { sleep: noSleep })).toBe('rejected');
    expect(refresh).toHaveBeenCalledTimes(1);
  });
});

describe('isSessionRejected', () => {
  it('only treats 401 and 403 as a rejection', () => {
    expect(isSessionRejected(rejected)).toBe(true);
    expect(isSessionRejected({ response: { status: 403 } })).toBe(true);
    expect(isSessionRejected({ response: { status: 500 } })).toBe(false);
    expect(isSessionRejected(new Error('Network Error'))).toBe(false);
    expect(isSessionRejected(null)).toBe(false);
  });
});
