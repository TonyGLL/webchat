'use client';

import { createContext, useCallback, useEffect, useMemo, useState } from 'react';
import { authService, LoginData, RegisterData, User } from '@/services/authService';
import { userService } from '@/services/userService';
import { AuthTokens, tokenStorage } from '@/lib/tokenStorage';

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
    tokenStorage.clear();
  }, []);

  // Keep React state in sync with the token store (login, silent refresh, logout, expired session).
  useEffect(() => {
    const unsubscribe = tokenStorage.subscribe((next) => {
      setTokens(next);
      if (!next) setUser(null);
    });

    const stored = tokenStorage.get();
    if (!stored) {
      setLoading(false);
      return unsubscribe;
    }

    setTokens(stored);
    userService
      .getProfile()
      .then(setUser)
      .catch(() => tokenStorage.clear())
      .finally(() => setLoading(false));

    return unsubscribe;
  }, []);

  const login = useCallback(async (data: LoginData) => {
    const { token, refreshToken } = await authService.login(data);
    tokenStorage.set({ accessToken: token, refreshToken });
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
