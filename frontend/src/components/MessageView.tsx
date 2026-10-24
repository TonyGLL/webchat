import { useState, useEffect, useMemo, useRef } from 'react';
import { ArrowLeft, MessageSquare, Send } from 'lucide-react';
import { useMessages } from '@/hooks/useMessages';
import Spinner from './Spinner';
import { useUser } from '@/hooks/useUser';
import dayjs from 'dayjs';
import isToday from 'dayjs/plugin/isToday';
import Reaction from './Reaction';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
dayjs.extend(isToday);

interface MessageViewProps {
  roomId: string | null;
  roomName?: string;
  onBack: () => void;
}

const formatTimestamp = (date: Date) => {
  const options: Intl.DateTimeFormatOptions = { hour: 'numeric', minute: 'numeric', hour12: true };

  if (!dayjs(date).isToday()) {
    options.day = 'numeric';
    options.month = 'short';
    options.year = 'numeric';
  }

  return new Date(date).toLocaleTimeString('en-US', options);
};

export default function MessageView({ roomId, roomName, onBack }: MessageViewProps) {
  const { messages, loading, createMessage, addReaction, removeReaction } = useMessages(roomId);
  const { profile } = useUser();
  const [content, setContent] = useState('');
  const messagesEndRef = useRef<HTMLDivElement>(null);
  // The API returns newest first; the chat shows oldest first.
  const ordered = useMemo(() => messages?.slice().reverse() ?? [], [messages]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [ordered]);

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId && content.trim()) {
      createMessage({ room_id: roomId, content });
      setContent('');
    }
  };

  if (!roomId) {
    return (
      <div className="flex h-full flex-col items-center justify-center gap-2 text-center text-muted-foreground">
        <MessageSquare className="h-6 w-6" />
        <p className="text-sm">Select a room to start messaging.</p>
      </div>
    );
  }

  return (
    <div className="flex h-full flex-col">
      <header className="flex h-14 shrink-0 items-center gap-2 border-b px-2 md:px-4">
        <Button variant="ghost" size="icon" className="md:hidden" onClick={onBack} aria-label="Back to rooms">
          <ArrowLeft className="h-4 w-4" />
        </Button>
        <h2 className="truncate text-sm font-medium">{roomName ?? 'Room'}</h2>
      </header>

      <div className="flex-1 overflow-y-auto px-4 py-6">
        {loading ? (
          <div className="py-8">
            <Spinner />
          </div>
        ) : (
          <ul className="mx-auto flex max-w-2xl flex-col gap-3">
            {ordered.map((msg) => {
              const isCurrentUser = msg.author_id === profile?.id;
              return (
                <li key={msg.id} className={cn('flex flex-col gap-1', isCurrentUser ? 'items-end' : 'items-start')}>
                  <div
                    className={cn(
                      'max-w-[80%] rounded-2xl px-3.5 py-2',
                      isCurrentUser
                        ? 'rounded-br-md bg-primary text-primary-foreground'
                        : 'rounded-bl-md bg-muted'
                    )}
                  >
                    <p className="whitespace-pre-wrap break-words text-sm">{msg.content}</p>
                    <p className="mt-0.5 text-right text-[11px] opacity-60">{formatTimestamp(msg.created_at)}</p>
                  </div>
                  {msg.reactions && profile && (
                    <Reaction
                      messageId={msg.id}
                      reactions={msg.reactions}
                      onAddReaction={addReaction}
                      onRemoveReaction={removeReaction}
                      currentUserId={profile.id}
                    />
                  )}
                </li>
              );
            })}
          </ul>
        )}
        <div ref={messagesEndRef} />
      </div>

      <form onSubmit={handleSendMessage} className="flex shrink-0 items-center gap-2 border-t p-3">
        <Input
          value={content}
          onChange={(e) => setContent(e.target.value)}
          placeholder="Write a message"
          aria-label="Message"
          className="flex-1"
        />
        <Button type="submit" size="icon" disabled={!content.trim()} aria-label="Send">
          <Send className="h-4 w-4" />
        </Button>
      </form>
    </div>
  );
}
