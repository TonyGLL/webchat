'use client';

import Picker from '@emoji-mart/react';
import { useEffect, useRef, useState } from 'react';
import { useT } from '@/lib/i18n/useT';

// The dataset (~500 KB) is only fetched when the picker is opened, never on first paint.
const loadData = async () => (await import('@emoji-mart/data')).default;

/** Full emoji picker in a floating layer: it never gets clipped by the scrolling message list. */
export default function EmojiPicker({ onSelect, onClose }: { onSelect: (emoji: string) => void; onClose: () => void }) {
  const { locale } = useT();
  const { t } = useT();
  const [esStrings, setEsStrings] = useState<object | null>(null);
  const dialog = useRef<HTMLDivElement>(null);

  // Records the opener (to give focus back on close) and takes focus right away; the picker's own
  // autoFocus then moves it to the search field once it has mounted.
  useEffect(() => {
    const opener = document.activeElement instanceof HTMLElement ? document.activeElement : null;
    if (!dialog.current?.contains(document.activeElement)) dialog.current?.focus();
    return () => {
      if (opener?.isConnected) opener.focus();
    };
  }, []);

  // The picker lives in a shadow root, so useFocusTrap cannot see its buttons: two guards at the ends
  // of the dialog catch Tab leaving it and wrap focus to the other end.
  const wrap = (toEnd: boolean) => {
    const root = dialog.current?.querySelector('em-emoji-picker')?.shadowRoot;
    const items = Array.from(root?.querySelectorAll<HTMLElement>('input, button, [tabindex="0"]') ?? []);
    (toEnd ? items[items.length - 1] : items[0] ?? dialog.current)?.focus();
  };

  useEffect(() => {
    if (locale !== 'es') return;
    let cancelled = false;
    import('@emoji-mart/data/i18n/es.json').then((m) => !cancelled && setEsStrings(m.default));
    return () => {
      cancelled = true;
    };
  }, [locale]);

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/30 p-4 sm:items-center" onClick={onClose}>
      {/* dynamicWidth: the picker fills this box instead of its fixed 352px, so its emoji grid never overflows sideways. */}
      <div
        ref={dialog}
        role="dialog"
        aria-modal="true"
        aria-label={t('chat.emojiPicker')}
        tabIndex={-1}
        className="w-[min(22rem,calc(100vw-2rem))] overflow-hidden rounded-lg shadow-xl focus-visible:outline-none"
        onClick={(e) => e.stopPropagation()}
      >
        <span tabIndex={0} onFocus={() => wrap(true)} />
        {(locale !== 'es' || esStrings) && (
          <Picker
            autoFocus
            dynamicWidth
            data={loadData}
            onEmojiSelect={(emoji: { native: string }) => onSelect(emoji.native)}
            theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
            previewPosition="none"
            locale={locale}
            i18n={locale === 'es' ? esStrings : undefined}
          />
        )}
        <span tabIndex={0} onFocus={() => wrap(false)} />
      </div>
    </div>
  );
}
