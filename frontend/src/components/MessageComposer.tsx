import { memo, useState } from 'react';
import { Send, Smile } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useT } from '@/lib/i18n/useT';
import EmojiPicker from './LazyEmojiPicker';

interface MessageComposerProps {
  /** Called with the raw text when the user sends a non-empty message. */
  onSend: (content: string) => void;
  /** Called while the user types (the caller throttles it). */
  onTyping: () => void;
}

/** Owns the draft text, so typing re-renders only this form and never the message list. */
function MessageComposer({ onSend, onTyping }: MessageComposerProps) {
  const { t } = useT();
  const [content, setContent] = useState('');
  const [showEmoji, setShowEmoji] = useState(false);

  const handleSubmit = (e: React.FormEvent) => {
    e.preventDefault();
    if (content.trim()) {
      onSend(content);
      setContent('');
    }
  };

  return (
    <>
      <form onSubmit={handleSubmit} className="shrink-0 px-3 pb-4 md:px-6 md:pb-5">
        <div className="flex h-12 items-center gap-2 rounded-lg border bg-card pl-4 pr-2 focus-within:ring-2 focus-within:ring-ring">
          <Input
            value={content}
            onChange={(e) => {
              setContent(e.target.value);
              if (e.target.value) onTyping();
            }}
            placeholder={t('chat.writeMessage')}
            aria-label={t('chat.message')}
            className="h-auto flex-1 border-0 bg-transparent p-0 shadow-none focus-visible:ring-0 focus-visible:ring-offset-0"
          />
          <Button type="button" variant="ghost" size="icon" className="h-8 w-8 text-muted-foreground" onClick={() => setShowEmoji(true)} aria-label={t('chat.emoji')} title={t('chat.emoji')}>
            <Smile className="h-4 w-4" />
          </Button>
          <Button type="submit" size="icon" className="h-8 w-8" disabled={!content.trim()} aria-label={t('chat.send')}>
            <Send className="h-4 w-4" />
          </Button>
        </div>
      </form>
      {showEmoji && (
        <EmojiPicker
          onSelect={(emoji) => {
            setContent((c) => c + emoji);
            setShowEmoji(false);
          }}
          onClose={() => setShowEmoji(false)}
        />
      )}
    </>
  );
}

export default memo(MessageComposer);
