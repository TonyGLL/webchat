import { useState, useEffect } from 'react';
import { userService } from '@/services/userService';

export const useUser = () => {
  const [profile, setProfile] = useState<any>(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const fetchProfile = async () => {
      try {
        const data = await userService.getProfile();
        setProfile(data);
      } catch (error) {
        console.error('Failed to fetch profile', error);
      } finally {
        setLoading(false);
      }
    };

    fetchProfile();
  }, []);

  const deactivate = async () => {
    await userService.deactivate();
    // Handle logout after deactivation
  };

  return { profile, loading, deactivate };
};