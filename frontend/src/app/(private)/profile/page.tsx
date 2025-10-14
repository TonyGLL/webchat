'use client';

import { useUser } from '@/hooks/useUser';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import PrivateRoute from '@/components/PrivateRoute';
import Spinner from '@/components/Spinner';
import { Card, CardContent, CardHeader, CardTitle } from '@/components/ui/card';
import { Button } from '@/components/ui/button';

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
      <div className="container mx-auto p-4 flex justify-center items-center h-full">
        {loading ? (
          <Spinner />
        ) : profile ? (
          <Card className="w-full max-w-md">
            <CardHeader>
              <CardTitle>User Profile</CardTitle>
            </CardHeader>
            <CardContent className="space-y-4">
              <div>
                <p className="font-semibold">ID</p>
                <p className="text-muted-foreground">{profile.id}</p>
              </div>
              <div>
                <p className="font-semibold">Username</p>
                <p className="text-muted-foreground">{profile.username}</p>
              </div>
              <div>
                <p className="font-semibold">Email</p>
                <p className="text-muted-foreground">{profile.email}</p>
              </div>
              <Button variant="destructive" onClick={handleDeactivate} className="w-full">
                Deactivate Account
              </Button>
            </CardContent>
          </Card>
        ) : (
          <p>Could not load profile.</p>
        )}
      </div>
    </PrivateRoute>
  );
}