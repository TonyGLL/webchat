'use client';

import { Button } from '@/components/ui/button';
import { Locale } from '@/lib/i18n';
import { setLocale, useT } from '@/lib/i18n/useT';
import { cn } from '@/lib/utils';

export default function LanguageToggle({ className }: { className?: string }) {
  const { t, locale } = useT();
  const next: Locale = locale === 'es' ? 'en' : 'es';

  return (
    <Button
      variant="ghost"
      size="icon"
      className={cn('text-xs font-semibold uppercase', className)}
      onClick={() => setLocale(next)}
      aria-label={`${t('nav.language')}: ${next.toUpperCase()}`}
      title={t('nav.language')}
    >
      <span lang={next}>{next}</span>
    </Button>
  );
}
