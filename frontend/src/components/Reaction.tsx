import data from '@emoji-mart/data'
import Picker from '@emoji-mart/react'
import esPickerStrings from '@emoji-mart/data/i18n/es.json'
import { useState } from 'react';
import { SmilePlus } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

interface ReactionProps {
  messageId: string;
  reactions: { user_id: number; emoji: string }[];
  onAddReaction: (messageId: string, emoji: string) => void;
  onRemoveReaction: (messageId: string, emoji: string) => void;
  currentUserId: number;
}

export default function Reaction({ messageId, reactions, onAddReaction, onRemoveReaction, currentUserId }: ReactionProps) {
  const { t, locale } = useT();
  const [showPicker, setShowPicker] = useState(false);

  const handleEmojiSelect = (emoji: { native: string }) => {
    const existingReaction = reactions.find((r) => r.emoji === emoji.native && r.user_id === currentUserId);
    if (existingReaction) {
      onRemoveReaction(messageId, emoji.native);
    } else {
      onAddReaction(messageId, emoji.native);
    }
    setShowPicker(false);
  };

  const groupReactions = (reactions: { user_id: number; emoji: string }[]) => {
    return reactions.reduce((acc, reaction) => {
      const existing = acc.find((item) => item.emoji === reaction.emoji);
      if (existing) {
        existing.count++;
        existing.users.push(reaction.user_id);
      } else {
        acc.push({ emoji: reaction.emoji, count: 1, users: [reaction.user_id] });
      }
      return acc;
    }, [] as { emoji: string; count: number; users: number[] }[]);
  };

  const groupedReactions = groupReactions(reactions);

  return (
    <div className="relative flex flex-wrap items-center gap-1">
      {groupedReactions.map((reaction) => {
        const mine = reaction.users.includes(currentUserId);
        return (
          <button
            key={reaction.emoji}
            type="button"
            onClick={() =>
              mine ? onRemoveReaction(messageId, reaction.emoji) : onAddReaction(messageId, reaction.emoji)
            }
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
        onClick={() => setShowPicker(!showPicker)}
        className="rounded-full p-1 text-muted-foreground transition-colors hover:bg-accent hover:text-foreground"
        aria-label={t('chat.addReaction')}
      >
        <SmilePlus className="h-4 w-4" />
      </button>
      {showPicker && (
        <div className="absolute bottom-full right-0 z-20 mb-2">
          <Picker data={data} onEmojiSelect={handleEmojiSelect} theme={document.documentElement.classList.contains('dark') ? 'dark' : 'light'} previewPosition="none" locale={locale} i18n={locale === 'es' ? esPickerStrings : undefined} />
        </div>
      )}
    </div>
  );
}
