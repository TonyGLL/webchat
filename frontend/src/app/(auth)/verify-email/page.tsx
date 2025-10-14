'use client';

import { useSearchParams } from 'next/navigation';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const email = searchParams.get('email');

  return (
    <div className="flex items-center justify-center min-h-screen bg-gray-100">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-md">
        <h1 className="text-2xl font-bold text-center">Verify Your Email</h1>
        <p className="text-center">
          An email has been sent to <strong>{email}</strong>. Please check your
          inbox and click the link to verify your account.
        </p>
      </div>
    </div>
  );
}