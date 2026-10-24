import { useState } from 'react';
import { SmilePlus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';
import EmojiPicker from './LazyEmojiPicker';

const QUICK_EMOJIS = ['👍', '❤️', '😂', '🎉', '😮', '🙏'];

interface ReactionProps {
  messageId: string;
  reactions: { user_id: number; emoji: string }[];
  onAddReaction: (messageId: string, emoji: string) => void;
  onRemoveReaction: (messageId: string, emoji: string) => void;
  currentUserId: number;
  /** Own messages sit on the right: open the quick menu towards the left so it stays on screen. */
  alignRight?: boolean;
}

export default function Reaction({ messageId, reactions, onAddReaction, onRemoveReaction, currentUserId, alignRight }: ReactionProps) {
  const { t } = useT();
  const [open, setOpen] = useState(false);
  const [full, setFull] = useState(false);

  const toggle = (emoji: string) => {
    const mine = reactions.some((r) => r.emoji === emoji && r.user_id === currentUserId);
    if (mine) onRemoveReaction(messageId, emoji);
    else onAddReaction(messageId, emoji);
    setOpen(false);
    setFull(false);
  };

  const grouped = reactions.reduce(
    (acc, r) => {
      const existing = acc.find((item) => item.emoji === r.emoji);
      if (existing) {
        existing.count++;
        existing.users.push(r.user_id);
      } else {
        acc.push({ emoji: r.emoji, count: 1, users: [r.user_id] });
      }
      return acc;
    },
    [] as { emoji: string; count: number; users: number[] }[]
  );

  return (
    <div className={cn('relative flex flex-wrap items-center gap-1', alignRight && 'justify-end')}>
      {grouped.map((reaction) => {
        const mine = reaction.users.includes(currentUserId);
        return (
          <button
            key={reaction.emoji}
            type="button"
            onClick={() => toggle(reaction.emoji)}
            className={cn(
              'h-6 rounded-full border bg-card px-2 text-xs transition-colors hover:bg-accent',
              mine && 'border-primary bg-soft font-medium text-primary hover:bg-soft'
            )}
          >
            {reaction.emoji}
            {reaction.count > 1 && <span className="ml-1 text-muted-foreground">{reaction.count}</span>}
          </button>
        );
      })}
      <button
        type="button"
        onClick={() => setOpen((v) => !v)}
        className={cn(
          'rounded-full p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground',
          // With no reactions yet the button only shows on hover (always on touch screens).
          grouped.length === 0 && 'md:opacity-0 md:focus:opacity-100 md:group-hover:opacity-100'
        )}
        aria-label={t('chat.react')}
        aria-expanded={open}
        title={t('chat.react')}
      >
        <SmilePlus className="h-4 w-4" />
      </button>
      {open && (
        <>
          <div className="fixed inset-0 z-10" onClick={() => setOpen(false)} aria-hidden="true" />
          <div
            className={cn(
              'absolute bottom-full z-20 mb-1 flex items-center gap-0.5 rounded-full border bg-card p-1 shadow-lg',
              alignRight ? 'right-0' : 'left-0'
            )}
          >
            {QUICK_EMOJIS.map((emoji) => (
              <button
                key={emoji}
                type="button"
                onClick={() => toggle(emoji)}
                className="flex h-8 w-8 items-center justify-center rounded-full text-lg transition-colors hover:bg-accent"
              >
                {emoji}
              </button>
            ))}
            <button
              type="button"
              onClick={() => setFull(true)}
              className="flex h-8 w-8 items-center justify-center rounded-full text-muted-foreground transition-colors hover:bg-accent"
              aria-label={t('chat.addReaction')}
              title={t('chat.addReaction')}
            >
              <SmilePlus className="h-4 w-4" />
            </button>
          </div>
        </>
      )}
      {full && <EmojiPicker onSelect={toggle} onClose={() => setFull(false)} />}
    </div>
  );
}
