import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tokenStorage } from './tokenStorage';

describe('tokenStorage', () => {
  beforeEach(() => {
    tokenStorage.clear();
    window.localStorage.clear();
  });

  it('keeps the access token in memory only, never in localStorage', () => {
    tokenStorage.set({ accessToken: 'secret-access' });

    expect(tokenStorage.get()).toEqual({ accessToken: 'secret-access' });
    expect(JSON.stringify({ ...window.localStorage })).not.toContain('secret-access');
  });

  it('remembers that a session may exist, without storing credentials', () => {
    expect(tokenStorage.mayHaveSession()).toBe(false);
    tokenStorage.set({ accessToken: 'a' });
    expect(tokenStorage.mayHaveSession()).toBe(true);
    tokenStorage.clear();
    expect(tokenStorage.mayHaveSession()).toBe(false);
  });

  it('notifies subscribers on set and clear until they unsubscribe', () => {
    const listener = vi.fn();
    const unsubscribe = tokenStorage.subscribe(listener);

    tokenStorage.set({ accessToken: 'a' });
    expect(listener).toHaveBeenLastCalledWith({ accessToken: 'a' });
    tokenStorage.clear();
    expect(listener).toHaveBeenLastCalledWith(null);

    unsubscribe();
    tokenStorage.set({ accessToken: 'b' });
    expect(listener).toHaveBeenCalledTimes(2);
  });
});
