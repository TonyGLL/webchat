'use client';

import { useUser } from '@/hooks/useUser';
import { useAuth } from '@/hooks/useAuth';
import { useRouter } from 'next/navigation';
import PrivateRoute from '@/components/PrivateRoute';
import Spinner from '@/components/Spinner';
import { Button } from '@/components/ui/button';
import { FormField } from '@/components/ui/form-field';
import Link from 'next/link';
import { ArrowLeft } from 'lucide-react';
import { useState } from 'react';
import { userService } from '@/services/userService';
import { useT } from '@/lib/i18n/useT';
import { logError } from '@/lib/logError';

export default function ProfilePage() {
  const { profile, loading, deactivate } = useUser();
  const { logout } = useAuth();
  const router = useRouter();
  const { t } = useT();

  const handleDeactivate = async () => {
    try {
      await deactivate();
      logout();
      router.push('/login');
    } catch (error) {
      logError('Failed to deactivate account', error);
    }
  };

  const [password, setPassword] = useState('');
  const [confirming, setConfirming] = useState(false);
  const [message, setMessage] = useState<string | null>(null);

  const handleExport = async () => {
    try {
      const url = URL.createObjectURL(await userService.exportData());
      const link = document.createElement('a');
      link.href = url;
      link.download = 'webchat-my-data.json';
      link.click();
      URL.revokeObjectURL(url);
    } catch {
      setMessage(t('profile.exportError'));
    }
  };

  const handleDelete = async () => {
    try {
      await userService.deleteAccount(password);
      logout();
      router.push('/login');
    } catch {
      setMessage(t('profile.deleteError'));
    }
  };

  return (
    <PrivateRoute>
      <main className="mx-auto flex min-h-dvh max-w-md flex-col justify-center px-4 py-10">
        <Link
          href="/dashboard"
          className="mb-6 inline-flex items-center gap-1.5 text-sm text-muted-foreground hover:text-foreground"
        >
          <ArrowLeft className="h-4 w-4" />
          {t('profile.back')}
        </Link>
        {loading ? (
          <Spinner />
        ) : profile ? (
          <div className="space-y-6">
            <div className="flex items-center gap-4">
              <div className="flex h-14 w-14 items-center justify-center rounded-full bg-muted text-xl font-medium">
                {profile.username?.charAt(0).toUpperCase()}
              </div>
              <div className="min-w-0">
                <h1 className="truncate text-xl font-semibold tracking-tight">{profile.username}</h1>
                <p className="truncate text-sm text-muted-foreground">{profile.email}</p>
              </div>
            </div>
            <div className="space-y-3 rounded-lg border p-4">
              <h2 className="text-sm font-medium">{t('profile.dataTitle')}</h2>
              <p className="text-sm text-muted-foreground">{t('profile.dataBody')}</p>
              <Button variant="outline" onClick={handleExport} className="w-full">
                {t('profile.download')}
              </Button>
            </div>
            <div className="space-y-3 rounded-lg border p-4">
              <h2 className="text-sm font-medium">{t('profile.dangerTitle')}</h2>
              <p className="text-sm text-muted-foreground">
                {t('profile.deactivateBody')}
              </p>
              <Button variant="destructive" onClick={handleDeactivate} className="w-full">
                {t('profile.deactivate')}
              </Button>
              <hr />
              <p id="delete-warning" className="text-sm text-muted-foreground">
                {t('profile.deleteBody')}
              </p>
              {confirming ? (
                <div className="space-y-2">
                  <FormField
                    label={t('profile.confirmPassword')}
                    type="password"
                    value={password}
                    onChange={(e) => setPassword(e.target.value)}
                    aria-describedby="delete-warning"
                    autoComplete="current-password"
                  />
                  <Button variant="destructive" disabled={!password} onClick={handleDelete} className="w-full">
                    {t('profile.deleteForever')}
                  </Button>
                </div>
              ) : (
                <Button variant="outline" onClick={() => setConfirming(true)} className="w-full">
                  {t('profile.delete')}
                </Button>
              )}
              {message && (
                <p className="text-sm text-destructive" role="alert">
                  {message}
                </p>
              )}
            </div>
          </div>
        ) : (
          <p className="text-sm text-muted-foreground">{t('profile.loadError')}</p>
        )}
      </main>
    </PrivateRoute>
  );
}