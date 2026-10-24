'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '@/hooks/useAuth';
import { makeRegisterSchema, RegisterData } from '@/services/authService';
import { useT } from '@/lib/i18n/useT';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/ui/form-field';
import Spinner from '@/components/Spinner';
import { logError } from '@/lib/logError';
import { setRegisteredFlash } from '@/lib/flash';
import { registerErrorKey } from '@/lib/registerErrors';

export default function RegisterPage() {
  const { register: registerUser } = useAuth();
  const { t } = useT();
  const schema = useMemo(() => makeRegisterSchema(t), [t]);
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<RegisterData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: RegisterData) => {
    setLoading(true);
    setError(null);
    try {
      await registerUser(data);
      setRegisteredFlash(sessionStorage);
      router.push('/login');
    } catch (err) {
      setError(t(registerErrorKey(err)));
      logError('Register failed', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t('register.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('register.subtitle')}</p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {error && (
          <p role="alert" className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        )}
        <div className="grid grid-cols-2 gap-3">
          <FormField
            label={t('register.name')}
            autoComplete="given-name"
            required
            error={errors.name?.message}
            {...register('name')}
          />
          <FormField
            label={t('register.lastName')}
            autoComplete="family-name"
            required
            error={errors.last_name?.message}
            {...register('last_name')}
          />
        </div>
        <FormField
          label={t('register.username')}
          autoComplete="username"
          required
          error={errors.username?.message}
          {...register('username')}
        />
        <FormField
          label={t('common.email')}
          type="email"
          autoComplete="email"
          required
          error={errors.email?.message}
          {...register('email')}
        />
        <FormField
          label={t('register.phone')}
          type="tel"
          autoComplete="tel"
          required
          error={errors.phone?.message}
          {...register('phone')}
        />
        <FormField
          label={t('common.password')}
          type="password"
          autoComplete="new-password"
          required
          hint={t('validation.passwordRules')}
          error={errors.password?.message}
          {...register('password')}
        />
        <Button type="submit" disabled={loading} className="w-full">
          {loading ? <Spinner className="h-4 w-4 text-primary-foreground" /> : t('register.submit')}
        </Button>
      </form>
      <p className="text-center text-sm text-muted-foreground">
        {t('register.haveAccount')}{' '}
        <Link href="/login" className="font-medium text-foreground hover:underline">
          {t('register.signIn')}
        </Link>
      </p>
    </div>
  );
}
