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
    <div className="flex items-center justify-between p-3 bg-whatsapp-deep-sea-green border-b border-gray-700">
      <div className="flex items-center">
        <div className="w-10 h-10 rounded-full bg-whatsapp-surfie-green flex items-center justify-center font-bold text-white">
          {profile?.username?.charAt(0).toUpperCase()}
        </div>
        <div className="ml-3">
          <p className="text-sm font-semibold text-white">{profile?.username}</p>
        </div>
      </div>
      <button
        onClick={handleLogout}
        className="px-3 py-2 text-sm font-medium text-white bg-whatsapp-surfie-green rounded-md hover:bg-opacity-80 focus:outline-none focus:ring-2 focus:ring-offset-2 focus:ring-offset-whatsapp-deep-sea-green focus:ring-white"
      >
        Logout
      </button>
    </div>
  );
}