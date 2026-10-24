import Link from 'next/link';
import { Hash } from 'lucide-react';
import LanguageToggle from '@/components/LanguageToggle';

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="relative flex min-h-dvh flex-col items-center justify-center gap-8 bg-background px-4 py-10">
      <div className="absolute right-3 top-3">
        <LanguageToggle />
      </div>
      <Link href="/" className="flex items-center gap-2.5 text-lg font-semibold tracking-tight">
        <span className="flex h-8 w-8 items-center justify-center rounded-lg bg-primary text-primary-foreground">
          <Hash className="h-[18px] w-[18px]" aria-hidden="true" />
        </span>
        WebChat
      </Link>
      <div className="w-full max-w-sm rounded-xl border bg-card p-7">{children}</div>
    </main>
  );
}
