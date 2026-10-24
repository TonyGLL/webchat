'use client';

import { useForm } from 'react-hook-form';
import { zodResolver } from '@hookform/resolvers/zod';
import { useAuth } from '@/hooks/useAuth';
import { useT } from '@/lib/i18n/useT';
import { LoginData, makeLoginSchema } from '@/services/authService';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { useMemo, useState } from 'react';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/ui/form-field';
import Spinner from '@/components/Spinner';
import { logError } from '@/lib/logError';

export default function LoginPage() {
  const { login } = useAuth();
  const { t } = useT();
  const schema = useMemo(() => makeLoginSchema(t), [t]);
  const router = useRouter();
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const {
    register,
    handleSubmit,
    formState: { errors },
  } = useForm<LoginData>({
    resolver: zodResolver(schema),
  });

  const onSubmit = async (data: LoginData) => {
    setLoading(true);
    setError(null);
    try {
      await login(data);
      router.push('/dashboard');
    } catch (err) {
      setError(t('login.invalid'));
      logError('Login failed', err);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="space-y-6">
      <div className="space-y-1">
        <h1 className="text-2xl font-semibold tracking-tight">{t('login.title')}</h1>
        <p className="text-sm text-muted-foreground">{t('login.subtitle')}</p>
      </div>
      <form onSubmit={handleSubmit(onSubmit)} className="space-y-4" noValidate>
        {error && (
          <p role="alert" className="rounded-md bg-destructive/10 px-3 py-2 text-sm text-destructive">
            {error}
          </p>
        )}
        <FormField
          label={t('common.email')}
          type="email"
          autoComplete="email"
          error={errors.user?.message}
          {...register('user')}
        />
        <FormField
          label={t('common.password')}
          type="password"
          autoComplete="current-password"
          error={errors.password?.message}
          {...register('password')}
        />
        <p className="text-right text-sm">
          <Link href="/forgot-password" className="text-muted-foreground hover:text-foreground hover:underline">
            {t('login.forgot')}
          </Link>
        </p>
        <Button type="submit" disabled={loading} className="w-full">
          {loading ? <Spinner className="h-4 w-4 text-primary-foreground" /> : t('login.submit')}
        </Button>
      </form>
      <p className="text-center text-sm text-muted-foreground">
        {t('login.noAccount')}{' '}
        <Link href="/register" className="font-medium text-foreground hover:underline">
          {t('login.createOne')}
        </Link>
      </p>
    </div>
  );
}
