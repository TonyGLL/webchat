'use client';

import { useSyncExternalStore } from 'react';
import { Moon, Sun } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { applyTheme, Theme } from '@/lib/theme';
import { useT } from '@/lib/i18n/useT';
import { cn } from '@/lib/utils';

// The <html class="dark"> flag is the source of truth, so the button just watches it.
function subscribe(onChange: () => void) {
  const observer = new MutationObserver(onChange);
  observer.observe(document.documentElement, { attributes: true, attributeFilter: ['class'] });
  return () => observer.disconnect();
}
const getTheme = (): Theme => (document.documentElement.classList.contains('dark') ? 'dark' : 'light');
const getServerTheme = (): Theme => 'light';

export default function ThemeToggle({ className }: { className?: string }) {
  const theme = useSyncExternalStore(subscribe, getTheme, getServerTheme);
  const next: Theme = theme === 'dark' ? 'light' : 'dark';
  const { t } = useT();
  const label = t('nav.switchTheme', { theme: t(`theme.${next}`) });

  return (
    <Button variant="ghost" size="icon" className={cn(className)} onClick={() => applyTheme(next)} aria-label={label} title={label}>
      {theme === 'dark' ? <Sun className="h-4 w-4" /> : <Moon className="h-4 w-4" />}
    </Button>
  );
}
