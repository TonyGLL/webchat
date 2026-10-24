'use client';

import { useEffect, useRef, useState } from 'react';
import { authService } from '@/services/authService';
import { useParams } from 'next/navigation';
import Link from 'next/link';
import { CheckCircle2, XCircle } from 'lucide-react';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';

export default function VerifyTokenPage() {
  const { token } = useParams();
  const [status, setStatus] = useState<'loading' | 'success' | 'error'>('loading');
  // The token is single-use; React dev mode runs effects twice, so guard the request.
  const requested = useRef(false);

  useEffect(() => {
    if (token && !requested.current) {
      requested.current = true;
      authService
        .verifyEmail(token as string)
        .then(() => setStatus('success'))
        .catch(() => setStatus('error'));
    }
  }, [token]);

  if (status === 'loading') {
    return (
      <div className="space-y-3 text-center">
        <Spinner className="h-6 w-6" />
        <p className="text-sm text-muted-foreground">Verifying your email…</p>
      </div>
    );
  }

  const ok = status === 'success';
  const Icon = ok ? CheckCircle2 : XCircle;

  return (
    <div className="space-y-4 text-center">
      <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted">
        <Icon className={`h-5 w-5 ${ok ? 'text-foreground' : 'text-destructive'}`} />
      </div>
      <h1 className="text-2xl font-semibold tracking-tight">
        {ok ? 'Email verified' : 'Link not valid'}
      </h1>
      <p className="text-sm text-muted-foreground">
        {ok ? 'Your account is ready.' : 'This verification link is invalid or has expired.'}
      </p>
      <Button asChild className="w-full">
        <Link href="/login">{ok ? 'Continue to sign in' : 'Back to sign in'}</Link>
      </Button>
    </div>
  );
}
