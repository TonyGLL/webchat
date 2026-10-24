import { useAuth } from './useAuth';
import { userService } from '@/services/userService';

/** The signed-in user's profile, shared through AuthContext (one request for the whole app). */
export const useUser = () => {
  const { user, loading } = useAuth();

  const deactivate = async (password: string) => {
    await userService.deactivate(password);
  };

  return { profile: user, loading, deactivate };
};
