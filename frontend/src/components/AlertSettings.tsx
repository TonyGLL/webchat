'use client';

import { Bell, BellOff } from 'lucide-react';
import { useEffect, useId, useRef, useState } from 'react';
import { Button } from '@/components/ui/button';
import { useAlertPrefs } from '@/hooks/useAlertPrefs';
import { popoverPosition } from '@/lib/popover';
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
        className="relative h-5 w-9 shrink-0 rounded-full bg-muted-foreground/70 transition-colors motion-reduce:transition-none peer-checked:bg-primary peer-focus-visible:ring-2 peer-focus-visible:ring-ring after:absolute after:left-0.5 after:top-0.5 after:h-4 after:w-4 after:rounded-full after:bg-card after:shadow after:transition-transform motion-reduce:after:transition-none peer-checked:after:translate-x-4"
      />
    </label>
  );
}

const PANEL_WIDTH = 256;

/** Bell button with the two notification switches: message sounds and desktop notifications. */
export default function AlertSettings({ className }: { className?: string }) {
  const { t } = useT();
  const { prefs, update } = useAlertPrefs();
  const [open, setOpen] = useState(false);
  const [hint, setHint] = useState<'blocked' | 'unsupported' | null>(null);
  // Fixed position computed from the trigger, so the panel is never clipped by the viewport edge
  // or by a container (the bell sits near the left edge, inside the sidebar).
  const [pos, setPos] = useState<{ top: number; left: number; width: number } | null>(null);
  const box = useRef<HTMLDivElement>(null);
  const trigger = useRef<HTMLButtonElement>(null);
  const panelId = useId();

  const placePanel = () => {
    const rect = trigger.current?.getBoundingClientRect();
    return rect ? popoverPosition(rect, window.innerWidth, PANEL_WIDTH) : null;
  };

  useEffect(() => {
    if (!open) return;
    const onDown = (e: MouseEvent) => !box.current?.contains(e.target as Node) && setOpen(false);
    const onKey = (e: KeyboardEvent) => {
      if (e.key !== 'Escape') return;
      setOpen(false);
      trigger.current?.focus();
    };
    const onResize = () => setPos(placePanel());
    document.addEventListener('mousedown', onDown);
    document.addEventListener('keydown', onKey);
    window.addEventListener('resize', onResize);
    return () => {
      document.removeEventListener('mousedown', onDown);
      document.removeEventListener('keydown', onKey);
      window.removeEventListener('resize', onResize);
    };
  }, [open]);

  const toggle = () => {
    if (!open) setPos(placePanel());
    setOpen((v) => !v);
  };

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
    <div
      ref={box}
      className="relative"
      // Keyboard focus moving out of the popover closes it (a null target is the window losing focus: stay open).
      onBlur={(e) => open && e.relatedTarget && !box.current?.contains(e.relatedTarget as Node) && setOpen(false)}
    >
      <Button
        ref={trigger}
        variant="ghost"
        size="icon"
        className={className}
        onClick={toggle}
        aria-label={t('alerts.open')}
        aria-expanded={open}
        aria-controls={panelId}
        title={t('alerts.open')}
      >
        <Icon className="h-4 w-4" />
      </Button>
      {open && (
        <div
          id={panelId}
          style={pos ?? undefined}
          className={cn('fixed z-30 rounded-lg border bg-card p-3 shadow-lg', !pos && 'invisible')}
        >
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
