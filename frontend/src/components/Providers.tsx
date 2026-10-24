'use client';

import { AuthProvider } from '@/contexts/AuthContext';
import { useHtmlLang } from '@/lib/i18n/useT';

export default function Providers({ children }: { children: React.ReactNode }) {
  useHtmlLang();
  return <AuthProvider>{children}</AuthProvider>;
}
