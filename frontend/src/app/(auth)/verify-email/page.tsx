'use client';

import { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import Link from 'next/link';
import { Mail } from 'lucide-react';
import { useT } from '@/lib/i18n/useT';

function VerifyEmailContent() {
  const email = useSearchParams().get('email');
  const { t } = useT();

  return (
    <div className="space-y-4 text-center">
      <div className="mx-auto flex h-12 w-12 items-center justify-center rounded-full bg-muted">
        <Mail className="h-5 w-5 text-muted-foreground" />
      </div>
      <h1 className="text-2xl font-semibold tracking-tight">{t('verifyEmail.title')}</h1>
      <p className="text-sm text-muted-foreground">
        {t('verifyEmail.sent', { email: email ?? t('verifyEmail.yourEmail') })}
      </p>
      <Link href="/login" className="inline-block text-sm font-medium hover:underline">
        {t('verifyEmail.back')}
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
