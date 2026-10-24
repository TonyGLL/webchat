'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { MailCheck } from 'lucide-react';
import { useEffect, useMemo, useRef, useState } from 'react';
import { authService, ForgotPasswordData, makeForgotPasswordSchema } from '@/services/authService';
import { useT } from '@/lib/i18n/useT';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/ui/form-field';
import Spinner from '@/components/Spinner';

export default function ForgotPasswordPage() {
  const { t } = useT();
  const schema = useMemo(() => makeForgotPasswordSchema(t), [t]);
  const [sentTo, setSentTo] = useState<string | null>(null);
  const [error, setError] = useState(false);
  const sentHeading = useRef<HTMLHeadingElement>(null);
  // The form is replaced by the confirmation: move focus to its heading so it is announced.
  useEffect(() => {
    if (sentTo) sentHeading.current?.focus();
  }, [sentTo]);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ForgotPasswordData>({ resolver: zodResolver(schema) });

  const onSubmit = async ({ email }: ForgotPasswordData) => {
    setError(false);
    try {
      await authService.forgotPassword(email);
      setSentTo(email);
    } catch {
      setError(true);
    }
  };

  // The same screen appears whether or not the address has an account.
  if (sentTo) {
    return (
      <div className="space-y-4 text-center">
        <MailCheck className="mx-auto h-8 w-8 text-primary" aria-hidden="true" />
        <h1 ref={sentHeading} tabIndex={-1} className="text-2xl font-semibold tracking-tight outline-none">{t('forgot.sentTitle')}</h1>
        <p className="text-sm text-muted-foreground">{t('forgot.sentBody', { email: sentTo })}</p>
        <p className="text-xs text-muted-foreground">{t('forgot.sentHint')}</p>
        <Button asChild variant="outline" className="w-full">
          <Link href="/login">{t('forgot.back')}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t('forgot.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('forgot.subtitle')}</p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {error && (
          <p role="alert" className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {t('forgot.failed')}
          </p>
        )}
        <FormField label={t('common.email')} type="email" autoComplete="email" required error={errors.email?.message} {...register('email')} />
        <Button type="submit" disabled={isSubmitting} className="w-full">
          {isSubmitting ? <Spinner className="h-4 w-4 text-primary-foreground" /> : t('forgot.submit')}
        </Button>
      </form>
      <p className="text-center text-sm">
        <Link href="/login" className="font-medium text-foreground hover:underline">
          {t('forgot.back')}
        </Link>
      </p>
    </div>
  );
}
