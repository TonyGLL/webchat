'use client';

import { Button } from '@/components/ui/button';
import { Locale } from '@/lib/i18n';
import { setLocale, useT } from '@/lib/i18n/useT';

export default function LanguageToggle() {
  const { t, locale } = useT();
  const next: Locale = locale === 'es' ? 'en' : 'es';

  return (
    <Button
      variant="ghost"
      size="icon"
      className="text-xs font-semibold uppercase"
      onClick={() => setLocale(next)}
      aria-label={t('nav.language')}
      title={t('nav.language')}
    >
      {next}
    </Button>
  );
}
