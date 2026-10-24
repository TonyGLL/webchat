'use client';

import { useUser } from '@/hooks/useUser';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import PrivateRoute from '@/components/PrivateRoute';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';

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
      <main className="mx-auto flex min-h-dvh max-w-md flex-col justify-center px-4 py-10">
        <Link
          href="/dashboard"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          Back to chats
        </Link>
        {loading ? (
          <Spinner />
        ) : profile ? (
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-full bg-muted text-xl font-medium">
                {profile.username?.charAt(0).toUpperCase()}
              </div>
              <div className="min-w-0">
                <h1 className="truncate text-xl font-semibold tracking-tight">{profile.username}</h1>
                <p className="truncate text-sm text-muted-foreground">{profile.email}</p>
              </div>
            </div>
            <div className="space-y-3 rounded-lg border p-4">
              <h2 className="text-sm font-medium">Danger zone</h2>
              <p className="text-sm text-muted-foreground">
                Deactivating your account signs you out and disables your access.
              </p>
              <Button variant="destructive" onClick={handleDeactivate} className="w-full">
                Deactivate account
              </Button>
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">Could not load profile.</p>
        )}
      </main>
    </PrivateRoute>
  );
}