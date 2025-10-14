'use client';

import { useEffect, useState } from 'react';
import { authService } from '@/services/authService';
import { useParams } from 'next/navigation';
import Link from 'next/link';

export default function VerifyTokenPage() {
  const { token } = useParams();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>(
    'loading'
  );

  useEffect(() => {
    if (token) {
      authService
        .verifyEmail(token as string)
        .then(() => setStatus('success'))
        .catch(() => setStatus('error'));
    }
  }, [token]);

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Email Verification</h1>
        {status === 'loading' && <p className="text-center">Verifying...</p>}
        {status === 'success' && (
          <div className="text-center">
            <p className="text-green-600">Email verified successfully!</p>
            <Link href="/login" className="text-indigo-600 hover:underline">
              Click here to login
            </Link>
          </div>
        )}
        {status === 'error' && (
          <p className="text-center text-red-600">
            Invalid or expired verification link.
          </p>
        )}
      </div>
    </div>
  );
}