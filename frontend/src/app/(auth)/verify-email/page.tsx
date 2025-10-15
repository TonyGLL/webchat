'use client';

import { useSearchParams } from 'next/navigation';

export default function VerifyEmailPage() {
  const searchParams = useSearchParams();
  const email = searchParams.get('email');

  return (
    <div className="flex items-center justify-center min-h-screen bg-whatsapp-deep-sea-green">
      <div className="w-full max-w-md p-8 space-y-6 bg-white rounded-lg shadow-lg text-center">
        <h1 className="text-3xl font-bold text-whatsapp-surfie-green">Verify Your Email</h1>
        <p className="text-gray-600">
          An email has been sent to <strong className="font-semibold text-gray-800">{email}</strong>. Please check your
          inbox and click the link to verify your account.
        </p>
      </div>
    </div>
  );
}