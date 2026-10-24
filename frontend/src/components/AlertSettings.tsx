'use client';

import { Bell, BellOff } from 'lucide-react';
import { useEffect, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { useAlertPrefs } from '@/hooks/useAlertPrefs';
import { playMessageSound } from '@/lib/sound';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

function Switch({ checked, onChange, label, disabled }: { checked: boolean; onChange: (v: boolean) => void; label: string; disabled?: boolean }) {
  return (
    <label className={cn('flex cursor-pointer items-center justify-between gap-3 py-1.5 text-sm', disabled && 'cursor-not-allowed opacity-50')}>
      <span>{label}</span>
      <input type="checkbox" role="switch" className="peer sr-only" checked={checked} disabled={disabled} onChange={(e) => onChange(e.target.checked)} />
      <span
        aria-hidden="true"
        className="relative h-5 w-9 shrink-0 rounded-full bg-border transition-colors peer-checked:bg-primary peer-focus-visible:ring-2 peer-focus-visible:ring-ring after:absolute after:left-0.5 after:top-0.5 after:h-4 after:w-4 after:rounded-full after:bg-card after:shadow after:transition-transform peer-checked:after:translate-x-4"
      />
    </label>
  );
}

/** Bell button with the two notification switches: message sounds and desktop notifications. */
export default function AlertSettings() {
  const { t } = useT();
  const { prefs, update } = useAlertPrefs();
  const [open, setOpen] = useState(false);
  const [hint, setHint] = useState<'blocked' | 'unsupported' | null>(null);
  const box = useRef<HTMLDivElement>(null);

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => !box.current?.contains(e.target as Node) && setOpen(false);
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && setOpen(false);
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
    };
  }, [open]);

  const setDesktop = async (on: boolean) => {
    setHint(null);
    if (!on) return update({ desktop: false });
    if (typeof Notification === 'undefined') return setHint('unsupported');
    // The permission prompt must come from a click, which is where this runs.
    const permission = Notification.permission === 'default' ? await Notification.requestPermission() : Notification.permission;
    if (permission === 'granted') update({ desktop: true });
    else setHint('blocked');
  };

  const anyOn = prefs.sound || prefs.desktop;
  const Icon = anyOn ? Bell : BellOff;

  return (
    <div ref={box} className="relative">
      <Button variant="ghost" size="icon" onClick={() => setOpen((v) => !v)} aria-label={t('alerts.open')} aria-expanded={open} title={t('alerts.open')}>
        <Icon className="h-4 w-4" />
      </Button>
      {open && (
        <div role="dialog" aria-label={t('alerts.title')} className="absolute right-0 top-full z-30 mt-1 w-64 rounded-lg border bg-card p-3 shadow-lg">
          <h3 className="mb-1 text-[13px] font-semibold">{t('alerts.title')}</h3>
          <Switch label={t('alerts.sound')} checked={prefs.sound} onChange={(sound) => update({ sound })} />
          <Switch label={t('alerts.desktop')} checked={prefs.desktop} onChange={setDesktop} />
          {hint && (
            <p role="alert" className="pt-1 text-xs text-destructive">
              {hint === 'blocked' ? t('alerts.blocked') : t('alerts.unsupported')}
            </p>
          )}
          <Button variant="outline" size="sm" className="mt-2 w-full" disabled={!prefs.sound} onClick={() => playMessageSound('direct')}>
            {t('alerts.test')}
          </Button>
        </div>
      )}
    </div>
  );
}
