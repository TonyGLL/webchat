import { beforeEach, describe, expect, it, vi } from 'vitest';
import { tokenStorage } from './tokenStorage';

describe('tokenStorage', () => {
  beforeEach(() => window.localStorage.clear());

  it('returns null when nothing is stored or the value is corrupt', () => {
    expect(tokenStorage.get()).toBeNull();
    window.localStorage.setItem('authTokens', '{not json');
    expect(tokenStorage.get()).toBeNull();
  });

  it('stores and clears tokens, notifying subscribers', () => {
    const listener = vi.fn();
    const unsubscribe = tokenStorage.subscribe(listener);
    const tokens = { accessToken: 'a', refreshToken: 'r' };

    tokenStorage.set(tokens);
    expect(tokenStorage.get()).toEqual(tokens);
    expect(listener).toHaveBeenLastCalledWith(tokens);

    tokenStorage.clear();
    expect(tokenStorage.get()).toBeNull();
    expect(listener).toHaveBeenLastCalledWith(null);

    unsubscribe();
    tokenStorage.set(tokens);
    expect(listener).toHaveBeenCalledTimes(2);
  });
});
