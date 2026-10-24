'use client';

import { Hash, MessageSquareHeart, Users } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { clampStep, ONBOARDING_STEPS, OnboardingStep } from '@/lib/onboarding';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

const icons: Record<OnboardingStep, typeof Users> = {
  people: Users,
  rooms: Hash,
  talk: MessageSquareHeart,
};

/** First-run welcome: three short steps. Skipping or finishing records it so it does not come back. */
export default function OnboardingDialog({ name, onFinish }: { name: string; onFinish: () => void }) {
  const { t } = useT();
  const [index, setIndex] = useState(0);
  const primary = useRef<HTMLButtonElement>(null);
  const step = ONBOARDING_STEPS[clampStep(index)];
  const last = index >= ONBOARDING_STEPS.length - 1;
  const Icon = icons[step];

  useEffect(() => {
    primary.current?.focus();
  }, [index]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onFinish();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onFinish]);

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/40 p-4">
      <div
        role="dialog"
        aria-modal="true"
        aria-labelledby="onboarding-title"
        className="w-full max-w-md rounded-xl border bg-card p-6 shadow-xl"
      >
        <p className="text-xs text-muted-foreground">{t('onboarding.welcome', { name })}</p>
        <div className="mt-4 flex h-11 w-11 items-center justify-center rounded-lg bg-soft text-primary">
          <Icon className="h-5 w-5" aria-hidden="true" />
        </div>
        <h2 id="onboarding-title" className="mt-4 text-lg font-semibold tracking-tight">
          {t(`onboarding.${step}.title`)}
        </h2>
        <p className="mt-2 text-sm text-muted-foreground">{t(`onboarding.${step}.body`)}</p>

        <div className="mt-6 flex items-center justify-between">
          <div className="flex items-center gap-1.5" role="img" aria-label={t('onboarding.step', { current: index + 1, total: ONBOARDING_STEPS.length })}>
            {ONBOARDING_STEPS.map((id, i) => (
              <span key={id} className={cn('h-1.5 rounded-full transition-all', i === index ? 'w-5 bg-primary' : 'w-1.5 bg-border')} />
            ))}
          </div>
          <div className="flex items-center gap-2">
            {!last && (
              <Button variant="ghost" size="sm" onClick={onFinish}>
                {t('onboarding.skip')}
              </Button>
            )}
            {index > 0 && (
              <Button variant="outline" size="sm" onClick={() => setIndex((i) => clampStep(i - 1))}>
                {t('onboarding.back')}
              </Button>
            )}
            <Button ref={primary} size="sm" onClick={() => (last ? onFinish() : setIndex((i) => clampStep(i + 1)))}>
              {last ? t('onboarding.start') : t('onboarding.next')}
            </Button>
          </div>
        </div>
      </div>
    </div>
  );
}
