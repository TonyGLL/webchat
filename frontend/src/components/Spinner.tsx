'use client';

import { Loader2 } from 'lucide-react';
import { useT } from '@/lib/i18n/useT';

export default function Spinner({ className = 'h-5 w-5' }: { className?: string }) {
  const { t } = useT();
  return (
    <div className="flex items-center justify-center" role="status" aria-label={t('common.loading')}>
      <Loader2 className={`animate-spin motion-reduce:animate-none text-muted-foreground ${className}`} />
    </div>
  );
}
