'use client';

import Link from 'next/link';
import { LogOut } from 'lucide-react';
import { useAuth } from '@/hooks/useAuth';
import { useUser } from '@/hooks/useUser';
import { useRouter } from 'next/navigation';
import ThemeToggle from '@/components/ThemeToggle';
import { Button } from '@/components/ui/button';

export default function UserProfile() {
  const { logout } = useAuth();
  const { profile } = useUser();
  const router = useRouter();

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  return (
    <div className="flex h-14 items-center justify-between border-b px-4">
      <Link href="/profile" className="flex min-w-0 items-center gap-3 rounded-md hover:opacity-80">
        <div className="flex h-8 w-8 shrink-0 items-center justify-center rounded-full bg-muted text-sm font-medium">
          {profile?.username?.charAt(0).toUpperCase()}
        </div>
        <p className="truncate text-sm font-medium">{profile?.username}</p>
      </Link>
      <div className="flex shrink-0 items-center">
      <ThemeToggle />
      <Button variant="ghost" size="icon" onClick={handleLogout} aria-label="Log out" title="Log out">
        <LogOut className="h-4 w-4" />
      </Button>
      </div>
    </div>
  );
}
