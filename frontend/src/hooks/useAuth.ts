import { useState, useEffect } from 'react';
import { authService, LoginData, RegisterData } from '@/services/authService';
import { userService } from '@/services/userService';

interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

export const useAuth = () => {
  const [user, setUser] = useState<any>(null);
  const [tokens, setTokens] = useState<AuthTokens | null>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const storedTokens = localStorage.getItem('authTokens');
    if (storedTokens) {
      const parsedTokens = JSON.parse(storedTokens);
      setTokens(parsedTokens);
      userService.getProfile().then(setUser).catch(() => {
        // Handle error, e.g., token expired
        logout();
      });
    }
    setLoading(false);
  }, []);

  const login = async (data: LoginData) => {
    const response = await authService.login(data);
    setTokens(response);
    localStorage.setItem('authTokens', JSON.stringify(response));
    const profile = await userService.getProfile();
    setUser(profile);
  };

  const register = async (data: RegisterData) => {
    await authService.register(data);
  };

  const logout = () => {
    setUser(null);
    setTokens(null);
    localStorage.removeItem('authTokens');
  };

  return { user, tokens, loading, login, register, logout };
};