'use client';

import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import { useEffect } from 'react';
import Spinner from './Spinner';

export default function PrivateRoute({
  children,
}: {
  children: React.ReactNode;
}) {
  const { tokens, loading } = useAuth();
  const router = useRouter();

  useEffect(() => {
    if (!loading && !tokens) {
      router.push('/login');
    }
  }, [loading, tokens, router]);

  if (loading || !tokens) {
    return (
      <div className="flex items-center justify-center min-h-screen">
        <Spinner />
      </div>
    );
  }

  return <>{children}</>;
}