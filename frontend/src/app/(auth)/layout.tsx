import Link from 'next/link';
import LanguageToggle from '@/components/LanguageToggle';

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="relative flex min-h-dvh flex-col items-center justify-center px-4 py-10">
      <div className="absolute right-3 top-3">
        <LanguageToggle />
      </div>
      <Link href="/" className="mb-8 text-lg font-semibold tracking-tight">
        WebChat
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </main>
  );
}
