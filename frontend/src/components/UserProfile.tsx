'use client';

import { useAuth } from '@/hooks/useAuth';
import { useUser } from '@/hooks/useUser';
import { useRouter } from 'next/navigation';

export default function UserProfile() {
  const { logout } = useAuth();
  const { profile } = useUser();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <div className="flex items-center justify-between p-4 bg-white dark:bg-gray-800 border-b border-gray-200 dark:border-gray-700">
      <div className="flex items-center">
        <div className="w-10 h-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center font-bold text-gray-600 dark:text-gray-300">
          {profile?.username?.charAt(0).toUpperCase()}
        </div>
        <div className="ml-3">
          <p className="text-sm font-semibold text-gray-800 dark:text-gray-200">{profile?.username}</p>
        </div>
      </div>
      <button
        onClick={handleLogout}
        className="px-3 py-2 text-sm font-medium text-white bg-red-600 rounded-md hover:bg-red-700 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-red-500"
      >
        Logout
      </button>
    </div>
  );
}