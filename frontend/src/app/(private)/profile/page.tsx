'use client';

import { useUser } from '@/hooks/useUser';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import PrivateRoute from '@/components/PrivateRoute';
import Spinner from '@/components/Spinner';

export default function ProfilePage() {
  const { profile, loading, deactivate } = useUser();
  const { logout } = useAuth();
  const router = useRouter();

  const handleDeactivate = async () => {
    try {
      await deactivate();
      logout();
      router.push('/login');
    } catch (error) {
      console.error('Failed to deactivate account', error);
    }
  };

  return (
    <PrivateRoute>
      <div className="container mx-auto p-4">
        <h1 className="text-2xl font-bold mb-4">User Profile</h1>
        {loading ? (
          <Spinner />
        ) : profile ? (
          <div className="bg-white shadow-md rounded-lg p-6">
            <p>
              <strong>ID:</strong> {profile.id}
            </p>
            <p>
              <strong>Username:</strong> {profile.username}
            </p>
            <p>
              <strong>Email:</strong> {profile.email}
            </p>
            <button
              onClick={handleDeactivate}
              className="mt-4 px-4 py-2 bg-red-600 text-white rounded hover:bg-red-700"
            >
              Deactivate Account
            </button>
          </div>
        ) : (
          <p>Could not load profile.</p>
        )}
      </div>
    </PrivateRoute>
  );
}