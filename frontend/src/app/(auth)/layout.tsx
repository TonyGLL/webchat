import Link from 'next/link';

export default function AuthLayout({ children }: { children: React.ReactNode }) {
  return (
    <main className="flex min-h-dvh flex-col items-center justify-center px-4 py-10">
      <Link href="/" className="mb-8 text-lg font-semibold tracking-tight">
        WebChat
      </Link>
      <div className="w-full max-w-sm">{children}</div>
    </main>
  );
}
