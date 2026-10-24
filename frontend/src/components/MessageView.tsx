import { useState, useEffect, useMemo, useRef } from 'react';
import { ArrowLeft, LogOut, Shield, Users, Flag, MessageSquare, Pin, PinOff, Reply, Send, Trash2, X } from 'lucide-react';
import { IGetRoomMessages } from '@/services/messageService';
import { useMessages } from '@/hooks/useMessages';
import { useRoomActivity } from '@/hooks/useRoomActivity';
import { typingLabel } from '@/lib/typing';
import ModerationPanel from './ModerationPanel';
import { moderationService } from '@/services/moderationService';
import { RANK, assignableRoles, canDeleteAnyMessage, canModerate, canPin as rankCanPin, rankOf, RoomRole } from '@/lib/permissions';
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
  /** Owners and admins can pin messages. */
  /** The user's role in this room; drives what moderation controls are shown. */
  myRole?: RoomRole;
  /** One-to-one conversation: no member list. */
  isDirect?: boolean;
  /** Starts (or reopens) a one-to-one conversation with a member. */
  onOpenDirect: (userId: number) => Promise<void>;
  onBack: () => void;
  /** Leaves the current room (the owner cannot). Resolves when the server accepted it. */
  onLeave: () => Promise<void>;
}

const formatTimestamp = (date: string | Date) => {
  const options: Intl.DateTimeFormatOptions = { hour: 'numeric', minute: 'numeric', hour12: true };

  if (!dayjs(date).isToday()) {
    options.day = 'numeric';
    options.month = 'short';
    options.year = 'numeric';
  }

  return new Date(date).toLocaleTimeString('en-US', options);
};

export default function MessageView({ roomId, roomName, myRole, isDirect = false, onOpenDirect, onBack, onLeave }: MessageViewProps) {
  const {
    messages,
    pinned,
    togglePin,
    pending,
    retryMessage,
    discardMessage,
    loading,
    loadingMore,
    hasMore,
    loadMore,
    createMessage,
    deleteMessage,
    addReaction,
    removeReaction,
  } = useMessages(roomId);
  const { profile } = useUser();
  const [content, setContent] = useState('');
  const [replyingTo, setReplyingTo] = useState<IGetRoomMessages | null>(null);
  const [showMembers, setShowMembers] = useState(false);
  const { members, typingNames, notifyTyping, changeRole, kickMember, banMember } = useRoomActivity(roomId);
  const myRank = rankOf(myRole, profile?.role);
  const canPin = rankCanPin(myRank);
  const [showModeration, setShowModeration] = useState(false);
  const [notice, setNotice] = useState<string | null>(null);
  const [memberError, setMemberError] = useState<string | null>(null);
  const manage = async (action: () => Promise<void>) => {
    setMemberError(null);
    try {
      await action();
    } catch {
      setMemberError('That action is not allowed.');
    }
  };
  const reportMessage = async (messageId: string) => {
    const reason = window.prompt('Why are you reporting this message?');
    if (!reason?.trim()) return;
    try {
      await moderationService.report(messageId, reason.trim());
      setNotice('Report sent. Thank you.');
    } catch {
      setNotice('Could not send the report (maybe you already reported it).');
    }
  };
  const onlineCount = members.filter((m) => m.online).length;
  const [leaveError, setLeaveError] = useState<string | null>(null);
  const messagesEndRef = useRef<HTMLDivElement>(null);
  const scrollRef = useRef<HTMLDivElement>(null);
  const newestIdRef = useRef<string | null>(null);
  const heightBeforeLoadRef = useRef<number | null>(null);
  // The API returns newest first; the chat shows oldest first.
  const ordered = useMemo(() => messages?.slice().reverse() ?? [], [messages]);

  // Follow the conversation only when a newer message arrives; when older ones are
  // prepended, keep the reader where they were.
  useEffect(() => {
    const container = scrollRef.current;
    const newestId = pending[0]?.tempId ?? ordered[ordered.length - 1]?.id ?? null;

    if (heightBeforeLoadRef.current !== null && container) {
      container.scrollTop = container.scrollHeight - heightBeforeLoadRef.current;
      heightBeforeLoadRef.current = null;
    } else if (newestId !== newestIdRef.current) {
      messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
    }
    newestIdRef.current = newestId;
  }, [ordered, pending]);

  const handleLeave = async () => {
    if (!window.confirm(`Leave ${roomName ?? 'this room'}?`)) return;
    try {
      await onLeave();
    } catch {
      setLeaveError("You can't leave this room (the owner must stay).");
    }
  };

  const byId = useMemo(() => new Map(messages.map((m) => [m.id, m])), [messages]);

  const handleLoadMore = async () => {
    heightBeforeLoadRef.current = scrollRef.current?.scrollHeight ?? null;
    await loadMore();
  };

  const handleSendMessage = (e: React.FormEvent) => {
    e.preventDefault();
    if (roomId && content.trim()) {
      createMessage({ room_id: roomId, content, reply_to_message_id: replyingTo?.id });
      setContent('');
      setReplyingTo(null);
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
        <div className="min-w-0 flex-1">
          <h2 className="truncate text-sm font-medium">{roomName ?? 'Room'}</h2>
          {!isDirect && members.length > 0 && (
            <p className="truncate text-[11px] text-muted-foreground">
              {members.length} {members.length === 1 ? 'member' : 'members'} · {onlineCount} online
            </p>
          )}
        </div>
        {leaveError && (
          <p className="text-xs text-destructive" role="alert">
            {leaveError}
          </p>
        )}
        {!isDirect && myRank >= RANK.moderator && (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setShowModeration((v) => !v)}
            aria-label="Moderation"
            aria-expanded={showModeration}
            title="Moderation"
          >
            <Shield className="h-4 w-4" />
          </Button>
        )}
        {!isDirect && (
          <Button
            variant="ghost"
            size="icon"
            onClick={() => setShowMembers((v) => !v)}
          aria-label="Members"
          aria-expanded={showMembers}
          title="Members"
          >
            <Users className="h-4 w-4" />
          </Button>
        )}
        <Button variant="ghost" size="icon" onClick={handleLeave} aria-label="Leave room" title="Leave room">
          <LogOut className="h-4 w-4" />
        </Button>
      </header>

      {showModeration && !isDirect && roomId && <ModerationPanel roomId={roomId} rank={myRank} members={members} />}

      {notice && (
        <p className="shrink-0 border-b bg-muted/40 px-4 py-1 text-xs" role="status">
          {notice} <button type="button" className="underline" onClick={() => setNotice(null)}>OK</button>
        </p>
      )}

      {showMembers && !isDirect && (
        <ul className="max-h-40 shrink-0 space-y-1 overflow-y-auto border-b bg-muted/40 px-4 py-2 text-xs">
          {members.map((m) => (
            <li key={m.user_id} className="flex items-center gap-2">
              <span
                className={cn('h-2 w-2 shrink-0 rounded-full', m.online ? 'bg-green-500' : 'bg-muted-foreground/40')}
                aria-label={m.online ? 'Online' : 'Offline'}
              />
              <span className="truncate">
                {m.name} <span className="text-muted-foreground">@{m.username}</span>
              </span>
              <span className="ml-auto flex items-center gap-2">
                {assignableRoles(myRank, rankOf(m.role, null)).length > 0 && m.user_id !== profile?.id ? (
                  <select
                    value={m.role}
                    aria-label={`Role of ${m.name}`}
                    className="rounded border bg-background px-1 py-0.5 text-[11px]"
                    onChange={(e) => manage(() => changeRole(m.user_id, e.target.value as RoomRole))}
                  >
                    {Array.from(new Set<RoomRole>([m.role, ...assignableRoles(myRank, rankOf(m.role, null))])).map((role) => (
                      <option key={role} value={role} disabled={!assignableRoles(myRank, rankOf(m.role, null)).includes(role)}>
                        {role}
                      </option>
                    ))}
                  </select>
                ) : (
                  m.role !== 'member' && <span className="text-[10px] uppercase text-muted-foreground">{m.role}</span>
                )}
                {m.user_id !== profile?.id && (
                  <button
                    type="button"
                    className="underline hover:text-foreground"
                    onClick={() => onOpenDirect(m.user_id).then(() => setShowMembers(false))}
                  >
                    Message
                  </button>
                )}
                {m.user_id !== profile?.id && canModerate(myRank, rankOf(m.role, null)) && (
                  <button
                    type="button"
                    className="underline hover:text-destructive"
                    onClick={() => window.confirm(`Remove ${m.name} from the room?`) && manage(() => kickMember(m.user_id))}
                  >
                    Remove
                  </button>
                )}
                {m.user_id !== profile?.id && canModerate(myRank, rankOf(m.role, null)) && (
                  <button
                    type="button"
                    className="underline hover:text-destructive"
                    onClick={() => {
                      const reason = window.prompt(`Ban ${m.name}. Reason?`);
                      if (reason === null) return;
                      const duration = window.prompt('For how long? Examples: 24h, 168h. Leave empty for permanent.') ?? '';
                      manage(() => banMember(m.user_id, reason, duration));
                    }}
                  >
                    Ban
                  </button>
                )}
              </span>
            </li>
          ))}
          {memberError && (
            <li className="text-destructive" role="alert">
              {memberError}
            </li>
          )}
        </ul>
      )}

      {pinned.length > 0 && (
        <div className="shrink-0 border-b bg-muted/40 px-4 py-2 text-xs">
          <p className="mb-1 flex items-center gap-1 font-medium text-muted-foreground">
            <Pin className="h-3 w-3" /> Pinned
          </p>
          <ul className="max-h-20 space-y-1 overflow-y-auto">
            {pinned.map((p) => (
              <li key={p.id} className="truncate">
                {p.content}
              </li>
            ))}
          </ul>
        </div>
      )}

      <div ref={scrollRef} className="flex-1 overflow-y-auto px-4 py-6">
        {loading ? (
          <div className="py-8">
            <Spinner />
          </div>
        ) : (
          <ul className="mx-auto flex max-w-2xl flex-col gap-3">
            {hasMore && (
              <li className="flex justify-center">
                <Button variant="ghost" size="sm" onClick={handleLoadMore} disabled={loadingMore}>
                  {loadingMore ? 'Loading…' : 'Load older messages'}
                </Button>
              </li>
            )}
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
                    {msg.reply_to_message_id && (
                      <p className="mb-1 truncate border-l-2 border-current pl-2 text-xs opacity-70">
                        {byId.get(msg.reply_to_message_id)?.content ?? 'Original message'}
                      </p>
                    )}
                    <p className="whitespace-pre-wrap break-words text-sm">{msg.content}</p>
                    <p className="mt-0.5 text-right text-[11px] opacity-60">{formatTimestamp(msg.created_at)}</p>
                  </div>
                  <div className="flex gap-3 text-[11px] text-muted-foreground">
                    <button type="button" className="flex items-center gap-1 hover:text-foreground" onClick={() => setReplyingTo(msg)}>
                      <Reply className="h-3 w-3" /> Reply
                    </button>
                    {canPin && (
                      <button
                        type="button"
                        className="flex items-center gap-1 hover:text-foreground"
                        onClick={() => togglePin(msg.id, !msg.pinned_at)}
                      >
                        {msg.pinned_at ? <PinOff className="h-3 w-3" /> : <Pin className="h-3 w-3" />}
                        {msg.pinned_at ? 'Unpin' : 'Pin'}
                      </button>
                    )}
                    {!isCurrentUser && !isDirect && (
                      <button type="button" className="flex items-center gap-1 hover:text-foreground" onClick={() => reportMessage(msg.id)}>
                        <Flag className="h-3 w-3" /> Report
                      </button>
                    )}
                    {(isCurrentUser || canDeleteAnyMessage(myRank)) && (
                      <button
                        type="button"
                        className="flex items-center gap-1 hover:text-destructive"
                        onClick={() => window.confirm('Delete this message?') && deleteMessage(msg.id)}
                      >
                        <Trash2 className="h-3 w-3" /> Delete
                      </button>
                    )}
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
            {pending.map((item) => (
              <li key={item.tempId} className="flex flex-col items-end gap-1">
                <div
                  className={cn(
                    'max-w-[80%] rounded-2xl rounded-br-md bg-primary px-3.5 py-2 text-primary-foreground',
                    item.status === 'sending' ? 'opacity-60' : 'opacity-100 ring-2 ring-destructive'
                  )}
                >
                  <p className="whitespace-pre-wrap break-words text-sm">{item.content}</p>
                </div>
                {item.status === 'sending' ? (
                  <p className="text-[11px] text-muted-foreground">Sending…</p>
                ) : (
                  <p className="flex gap-2 text-[11px] text-destructive" role="alert">
                    Not sent.
                    <button type="button" className="underline" onClick={() => retryMessage(item.tempId)}>
                      Retry
                    </button>
                    <button type="button" className="underline" onClick={() => discardMessage(item.tempId)}>
                      Discard
                    </button>
                  </p>
                )}
              </li>
            ))}
          </ul>
        )}
        <div ref={messagesEndRef} />
      </div>

      {replyingTo && (
        <div className="flex shrink-0 items-center gap-2 border-t bg-muted/50 px-3 py-2 text-xs">
          <Reply className="h-3 w-3 shrink-0" />
          <p className="min-w-0 flex-1 truncate">Replying to: {replyingTo.content}</p>
          <button type="button" onClick={() => setReplyingTo(null)} aria-label="Cancel reply">
            <X className="h-3 w-3" />
          </button>
        </div>
      )}
      <p className="h-4 shrink-0 px-4 text-[11px] text-muted-foreground" aria-live="polite">
        {typingLabel(typingNames)}
      </p>
      <form onSubmit={handleSendMessage} className="flex shrink-0 items-center gap-2 border-t p-3">
        <Input
          value={content}
          onChange={(e) => {
            setContent(e.target.value);
            if (e.target.value) notifyTyping();
          }}
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
