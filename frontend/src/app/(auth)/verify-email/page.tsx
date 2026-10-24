'use client';

import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Mail } from 'lucide-react';

function VerifyEmailContent() {
  const email = useSearchParams().get('email');

  return (
    <div className="space-y-4 text-center">
      <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted">
        <Mail className="h-5 w-5 text-muted-foreground" />
      </div>
      <h1 className="text-2xl font-semibold tracking-tight">Check your inbox</h1>
      <p className="text-sm text-muted-foreground">
        We sent a verification link to{' '}
        {email ? <strong className="font-medium text-foreground">{email}</strong> : 'your email'}.
        It expires in 10 minutes.
      </p>
      <Link href="/login" className="inline-block text-sm font-medium hover:underline">
        Back to sign in
      </Link>
    </div>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense fallback={null}>
      <VerifyEmailContent />
    </Suspense>
  );
}
