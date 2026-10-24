'use client';

import Picker from '@emoji-mart/react';
import { useEffect, useState } from 'react';
import { useT } from '@/lib/i18n/useT';

// The dataset (~500 KB) is only fetched when the picker is opened, never on first paint.
const loadData = async () => (await import('@emoji-mart/data')).default;

/** Full emoji picker in a floating layer: it never gets clipped by the scrolling message list. */
export default function EmojiPicker({ onSelect, onClose }: { onSelect: (emoji: string) => void; onClose: () => void }) {
  const { locale } = useT();
  const [esStrings, setEsStrings] = useState<object | null>(null);

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
      <div className="w-[min(22rem,calc(100vw-2rem))] overflow-hidden rounded-lg shadow-xl" onClick={(e) => e.stopPropagation()}>
        {(locale !== 'es' || esStrings) && (
          <Picker
            dynamicWidth
            data={loadData}
            onEmojiSelect={(emoji: { native: string }) => onSelect(emoji.native)}
            theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
            previewPosition="none"
            locale={locale}
            i18n={locale === 'es' ? esStrings : undefined}
          />
        )}
      </div>
    </div>
  );
}
