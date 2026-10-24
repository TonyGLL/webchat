'use client';

import data from '@emoji-mart/data';
import Picker from '@emoji-mart/react';
import esPickerStrings from '@emoji-mart/data/i18n/es.json';
import { useEffect } from 'react';
import { useT } from '@/lib/i18n/useT';

/** Full emoji picker in a floating layer: it never gets clipped by the scrolling message list. */
export default function EmojiPicker({ onSelect, onClose }: { onSelect: (emoji: string) => void; onClose: () => void }) {
  const { locale } = useT();

  useEffect(() => {
    const onKey = (e: KeyboardEvent) => e.key === 'Escape' && onClose();
    window.addEventListener('keydown', onKey);
    return () => window.removeEventListener('keydown', onKey);
  }, [onClose]);

  return (
    <div className="fixed inset-0 z-50 flex items-end justify-center bg-black/30 p-4 sm:items-center" onClick={onClose}>
      <div onClick={(e) => e.stopPropagation()}>
        <Picker
          data={data}
          onEmojiSelect={(emoji: { native: string }) => onSelect(emoji.native)}
          theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'}
          previewPosition="none"
          locale={locale}
          i18n={locale === 'es' ? esPickerStrings : undefined}
        />
      </div>
    </div>
  );
}
