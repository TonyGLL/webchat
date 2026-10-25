'use client';

import { createContext, useCallback, useEffect, useMemo, useState } from 'react';
import { authService, LoginData, RegisterData, User } from '@/services/authService';
import { userService } from '@/services/userService';
import { refreshAccessToken } from '@/lib/api';
import { AuthTokens, tokenStorage } from '@/lib/tokenStorage';
import { restoreSession } from '@/lib/sessionRestore';

interface AuthContextValue {
  /** Profile of the signed-in user; null while loading or signed out. */
  user: User | null;
  tokens: AuthTokens | null;
  /** True until the stored session has been read and the profile request settled. */
  loading: boolean;
  login: (data: LoginData) => Promise<void>;
  register: (data: RegisterData) => Promise<void>;
  logout: () => void;
}

export const AuthContext = createContext<AuthContextValue | null>(null);

export function AuthProvider({ children }: { children: React.ReactNode }) {
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const [user, setUser] = useState<User | null>(null);
  const [loading, setLoading] = useState(true);

  const logout = useCallback(() => {
    // End the local session first (it must not wait on the network), then revoke the refresh
    // cookie on the server without waiting for the answer.
    tokenStorage.clear();
    authService.logout().catch(() => undefined);
  }, []);

  // Keep React state in sync with the token store (login, silent refresh, logout, expired session).
  useEffect(() => {
    const unsubscribe = tokenStorage.subscribe((next) => {
      setTokens(next);
      if (!next) setUser(null);
    });
    const stopWatching = tokenStorage.watchRemoteSignOut();

    // Restore the session after a reload: the access token is memory-only, so ask for a new
    // one with the httpOnly refresh cookie.
    let cancelled = false;
    const restore = async () => {
      if (!tokenStorage.get() && tokenStorage.mayHaveSession()) {
        // Only an explicit rejection ends the session; an unreachable server keeps the stored
        // hint, so the next load tries again instead of asking the user to sign in.
        if ((await restoreSession(refreshAccessToken)) === 'rejected') tokenStorage.clear();
      }
      if (cancelled) return;
      if (!tokenStorage.get()) {
        setLoading(false);
        return;
      }
      setTokens(tokenStorage.get());
      try {
        const profile = await userService.getProfile();
        if (!cancelled) setUser(profile);
      } catch {
        tokenStorage.clear();
      } finally {
        if (!cancelled) setLoading(false);
      }
    };
    void restore();

    return () => {
      cancelled = true;
      unsubscribe();
      stopWatching();
    };
  }, []);


  const login = useCallback(async (data: LoginData) => {
    const { token } = await authService.login(data);
    tokenStorage.set({ accessToken: token });
    setUser(await userService.getProfile());
  }, []);

  // Email verification is paused for now: registering only creates the account.
  const register = useCallback(async (data: RegisterData) => {
    await authService.register(data);
  }, []);

  const value = useMemo(
    () => ({ user, tokens, loading, login, register, logout }),
    [user, tokens, loading, login, register, logout]
  );

  return <AuthContext.Provider value={value}>{children}</AuthContext.Provider>;
}
