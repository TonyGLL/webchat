'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import Link from 'next/link';
import { useParams } from 'next/navigation';
import { CheckCircle2 } from 'lucide-react';
import { useMemo, useState } from 'react';
import { authService } from '@/services/authService';
import { makeResetPasswordSchema, ResetPasswordData } from '@/lib/validation/auth';
import { resetErrorKey } from '@/lib/resetErrors';
import { useT } from '@/lib/i18n/useT';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/ui/form-field';
import Spinner from '@/components/Spinner';

export default function ResetPasswordPage() {
  const { token } = useParams<{ token: string }>();
  const { t } = useT();
  const schema = useMemo(() => makeResetPasswordSchema(t), [t]);
  const [done, setDone] = useState(false);
  const [errorKey, setErrorKey] = useState<ReturnType<typeof resetErrorKey> | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors, isSubmitting },
  } = useForm<ResetPasswordData>({ resolver: zodResolver(schema) });

  const onSubmit = async ({ password }: ResetPasswordData) => {
    setErrorKey(null);
    try {
      await authService.resetPassword(decodeURIComponent(token), password);
      setDone(true);
    } catch (err) {
      setErrorKey(resetErrorKey(err));
    }
  };

  if (done) {
    return (
      <div className="space-y-4 text-center">
        <CheckCircle2 className="mx-auto h-8 w-8 text-success" aria-hidden="true" />
        <h1 className="text-2xl font-semibold tracking-tight">{t('reset.doneTitle')}</h1>
        <p className="text-sm text-muted-foreground">{t('reset.doneBody')}</p>
        <Button asChild className="w-full">
          <Link href="/login">{t('reset.signIn')}</Link>
        </Button>
      </div>
    );
  }

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t('reset.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('reset.subtitle')}</p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {errorKey && (
          <div role="alert" className="space-y-1 rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
            <p>{t(errorKey)}</p>
            {errorKey === 'reset.invalidLink' && (
              <Link href="/forgot-password" className="font-medium underline">
                {t('reset.requestNew')}
              </Link>
            )}
          </div>
        )}
        <FormField label={t('reset.password')} type="password" autoComplete="new-password" required hint={t('validation.passwordRules')} error={errors.password?.message} {...register('password')} />
        <FormField label={t('reset.confirm')} type="password" autoComplete="new-password" required error={errors.confirm?.message} {...register('confirm')} />
        <Button type="submit" disabled={isSubmitting} className="w-full">
          {isSubmitting ? <Spinner className="h-4 w-4 text-primary-foreground" /> : t('reset.submit')}
        </Button>
      </form>
    </div>
  );
}
