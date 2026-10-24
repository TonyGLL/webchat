'use client';

import Link from 'next/link';
import { LogOut } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useUser } from '@/hooks/useUser';
import { useRouter } from 'next/navigation';
import ThemeToggle from '@/components/ThemeToggle';
import LanguageToggle from '@/components/LanguageToggle';
import { useT } from '@/lib/i18n/useT';
import { Button } from '@/components/ui/button';
import Avatar from '@/components/Avatar';

export default function UserProfile() {
  const { logout } = useAuth();
  const { profile } = useUser();
  const router = useRouter();
  const { t } = useT();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <div className="flex h-14 items-center justify-between border-b px-4 pr-3">
      <Link href="/profile" className="flex min-w-0 items-center gap-2.5 rounded-md hover:opacity-80">
        <Avatar name={profile?.username} tone="me" size="sm" />
        <div className="min-w-0 leading-tight">
          <p className="truncate text-[13px] font-semibold">{profile?.username}</p>
          <p className="text-[11px] text-muted-foreground">{t('nav.online')}</p>
        </div>
      </Link>
      <div className="flex shrink-0 items-center">
      <LanguageToggle />
      <ThemeToggle />
      <Button variant="ghost" size="icon" onClick={handleLogout} aria-label={t('nav.logOut')} title={t('nav.logOut')}>
        <LogOut className="h-4 w-4" />
      </Button>
      </div>
    </div>
  );
}
