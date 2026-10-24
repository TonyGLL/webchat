import { useState, useEffect, useMemo, useRef } from 'react';
import { ArrowLeft, LogOut, Smile, UserPlus, Shield, Users, Flag, MessageSquare, Pin, PinOff, Reply, Send, Trash2, X } from 'lucide-react';
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
import Avatar from './Avatar';
import EmojiPicker from './EmojiPicker';
import InviteUsers from './InviteUsers';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { cn } from '@/lib/utils';
import { intlLocale, Locale } from '@/lib/i18n';
import { useT } from '@/lib/i18n/useT';
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

const formatTimestamp = (date: string | Date, locale: Locale) => {
  const options: Intl.DateTimeFormatOptions = { hour: 'numeric', minute: 'numeric' };

  if (!dayjs(date).isToday()) {
    options.day = 'numeric';
    options.month = 'short';
    options.year = 'numeric';
  }

  return new Date(date).toLocaleTimeString(intlLocale(locale), options);
};

export default function MessageView({ roomId, roomName, myRole, isDirect = false, onOpenDirect, onBack, onLeave }: MessageViewProps) {
  const { t, locale } = useT();
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
      setMemberError(t('chat.actionNotAllowed'));
    }
  };
  const reportMessage = async (messageId: string) => {
    const reason = window.prompt(t('chat.reportPrompt'));
    if (!reason?.trim()) return;
    try {
      await moderationService.report(messageId, reason.trim());
      setNotice(t('chat.reportSent'));
    } catch {
      setNotice(t('chat.reportFailed'));
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
    if (!window.confirm(t('chat.leaveConfirm', { room: roomName ?? t('chat.thisRoom') }))) return;
    try {
      await onLeave();
    } catch {
      setLeaveError(t('chat.leaveError'));
    }
  };

  const byId = useMemo(() => new Map(messages.map((m) => [m.id, m])), [messages]);
  const [showInvite, setShowInvite] = useState(false);
  const [showEmoji, setShowEmoji] = useState(false);
  const memberById = useMemo(() => new Map(members.map((m) => [m.user_id, m])), [members]);
  const fullName = (m: { name: string; last_name?: string | null }) => [m.name, m.last_name].filter(Boolean).join(' ');
  const authorName = (id: number) => {
    if (id === profile?.id) return fullName(profile);
    const member = memberById.get(id);
    return member ? fullName(member) : t('common.userNumber', { id });
  };
  // In a direct conversation the other member is who the header is about.
  const peer = isDirect ? members.find((m) => m.user_id !== profile?.id) : undefined;
  const roleLabel = (role?: RoomRole) =>
    role && role !== 'member' ? (
      <span className={cn('text-[10px] font-semibold uppercase tracking-wide', role === 'moderator' ? 'text-muted-foreground' : 'text-primary')}>
        {t(`role.${role}`)}
      </span>
    ) : null;

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
        <p className="text-sm">{t('chat.selectRoom')}</p>
      </div>
    );
  }

  const iconButton = 'h-8 w-8 text-muted-foreground';
  const actionClass = 'flex items-center gap-1 hover:text-foreground';

  const membersList = (
    <ul className="space-y-3 text-xs">
      {members.map((m) => {
        const others = m.user_id !== profile?.id;
        const assignable = assignableRoles(myRank, rankOf(m.role, null));
        return (
          <li key={m.user_id} className="space-y-1">
            <div className="flex items-center gap-2 text-sm">
              <span
                className={cn('h-2 w-2 shrink-0 rounded-full', m.online ? 'bg-success' : 'border border-muted-foreground/60')}
                role="img"
                aria-label={m.online ? t('chat.online') : t('chat.offline')}
              />
              <span className={cn('min-w-0 flex-1 truncate', !m.online && 'text-muted-foreground')}>
                {fullName(m)} <span className="text-xs text-muted-foreground">@{m.username}</span>
              </span>
              {assignable.length === 0 || !others ? roleLabel(m.role) : null}
            </div>
            {others && (
              <div className="flex flex-wrap items-center gap-x-3 gap-y-1 pl-4 text-xs text-muted-foreground">
                {assignable.length > 0 && (
                  <select
                    value={m.role}
                    aria-label={t('chat.roleOf', { name: m.name })}
                    className="rounded border bg-background px-1 py-0.5 text-[11px] text-foreground"
                    onChange={(e) => manage(() => changeRole(m.user_id, e.target.value as RoomRole))}
                  >
                    {Array.from(new Set<RoomRole>([m.role, ...assignable])).map((role) => (
                      <option key={role} value={role} disabled={!assignable.includes(role)}>
                        {t(`role.${role}`)}
                      </option>
                    ))}
                  </select>
                )}
                <button type="button" className="hover:text-foreground hover:underline" onClick={() => onOpenDirect(m.user_id).then(() => setShowMembers(false))}>
                  {t('chat.message')}
                </button>
                {canModerate(myRank, rankOf(m.role, null)) && (
                  <>
                    <button
                      type="button"
                      className="hover:text-destructive hover:underline"
                      onClick={() => window.confirm(t('chat.removeConfirm', { name: m.name })) && manage(() => kickMember(m.user_id))}
                    >
                      {t('chat.remove')}
                    </button>
                    <button
                      type="button"
                      className="hover:text-destructive hover:underline"
                      onClick={() => {
                        const reason = window.prompt(t('chat.banPrompt', { name: m.name }));
                        if (reason === null) return;
                        const duration = window.prompt(t('chat.banDurationPrompt')) ?? '';
                        manage(() => banMember(m.user_id, reason, duration));
                      }}
                    >
                      {t('chat.ban')}
                    </button>
                  </>
                )}
              </div>
            )}
          </li>
        );
      })}
      {memberError && (
        <li className="text-destructive" role="alert">
          {memberError}
        </li>
      )}
    </ul>
  );

  return (
    <div className="relative flex h-full">
      <div className="flex min-w-0 flex-1 flex-col">
        <header className="flex h-14 shrink-0 items-center gap-2 border-b bg-card px-2 md:px-6">
          <Button variant="ghost" size="icon" className="md:hidden" onClick={onBack} aria-label={t('chat.backToRooms')}>
            <ArrowLeft className="h-4 w-4" />
          </Button>
          {isDirect && (
            <span className="relative shrink-0">
              <Avatar name={peer ? fullName(peer) : roomName} />
              {peer && (
                <span
                  role="img"
                  aria-label={peer.online ? t('chat.online') : t('chat.offline')}
                  className={cn(
                    'absolute -bottom-0.5 -right-0.5 h-3 w-3 rounded-full border-2 border-card',
                    peer.online ? 'bg-success' : 'bg-muted-foreground/50'
                  )}
                />
              )}
            </span>
          )}
          <div className="min-w-0 flex-1">
            <h2 className="flex items-center gap-1.5 truncate text-[15px] font-semibold leading-tight">
              {!isDirect && (
                <span className="font-mono font-medium text-muted-foreground" aria-hidden="true">
                  #
                </span>
              )}
              <span className="truncate">{(peer ? fullName(peer) : roomName) ?? t('rooms.fallbackName')}</span>
            </h2>
            {!isDirect && members.length > 0 && (
              <p className="truncate text-xs text-muted-foreground">
                {t(members.length === 1 ? 'chat.membersCount.one' : 'chat.membersCount.other', { count: members.length, online: onlineCount })}
              </p>
            )}
            {peer && (
              <p className="truncate text-xs text-muted-foreground">
                <span className={peer.online ? 'font-medium text-success' : undefined}>{peer.online ? t('chat.online') : t('chat.offline')}</span> · @{peer.username}
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
              className={cn(iconButton, showModeration && 'bg-soft text-primary hover:bg-soft')}
              onClick={() => setShowModeration((v) => !v)}
              aria-label={t('chat.moderation')}
              aria-expanded={showModeration}
              title={t('chat.moderation')}
            >
              <Shield className="h-4 w-4" />
            </Button>
          )}
          {!isDirect && myRank >= RANK.admin && (
            <Button
              variant="ghost"
              size="icon"
              className={cn(iconButton, showInvite && 'bg-soft text-primary hover:bg-soft')}
              onClick={() => setShowInvite((v) => !v)}
              aria-label={t('invite.button')}
              aria-expanded={showInvite}
              title={t('invite.button')}
            >
              <UserPlus className="h-4 w-4" />
            </Button>
          )}
          {!isDirect && (
            <Button
              variant="ghost"
              size="icon"
              className={cn(iconButton, showMembers && 'bg-soft text-primary hover:bg-soft')}
              onClick={() => setShowMembers((v) => !v)}
              aria-label={t('chat.members')}
              aria-expanded={showMembers}
              title={t('chat.members')}
            >
              <Users className="h-4 w-4" />
            </Button>
          )}
          <Button variant="ghost" size="icon" className={iconButton} onClick={handleLeave} aria-label={t('chat.leave')} title={t('chat.leave')}>
            <LogOut className="h-4 w-4" />
          </Button>
        </header>

        {showInvite && !isDirect && roomId && (
          <InviteUsers roomId={roomId} roomName={roomName ?? t('rooms.fallbackName')} memberIds={members.map((m) => m.user_id)} onClose={() => setShowInvite(false)} />
        )}

        {showModeration && !isDirect && roomId && <ModerationPanel roomId={roomId} rank={myRank} members={members} />}

        {notice && (
          <p className="shrink-0 border-b bg-muted/40 px-6 py-1 text-xs" role="status">
            {notice}{' '}
            <button type="button" className="underline" onClick={() => setNotice(null)}>
              {t('common.ok')}
            </button>
          </p>
        )}

        {pinned.length > 0 && (
          <div className="shrink-0 border-b bg-soft px-6 py-2 text-[13px]">
            <p className="mb-0.5 flex items-center gap-1.5 text-xs font-semibold text-primary">
              <Pin className="h-3 w-3" /> {t('chat.pinned')}
            </p>
            <ul className="max-h-20 space-y-0.5 overflow-y-auto">
              {pinned.map((p) => (
                <li key={p.id} className="truncate">
                  {p.content}
                </li>
              ))}
            </ul>
          </div>
        )}

        <div ref={scrollRef} className="flex-1 overflow-y-auto overflow-x-hidden px-4 py-5 md:px-6">
          {loading ? (
            <div className="py-8">
              <Spinner />
            </div>
          ) : (
            <ul className="flex flex-col gap-1">
              {!hasMore && ordered.length === 0 && pending.length === 0 && (
                <li className="flex flex-col items-center gap-3 px-6 py-16 text-center text-muted-foreground">
                  {isDirect ? <Avatar name={peer ? fullName(peer) : roomName} size="md" /> : <MessageSquare className="h-6 w-6" aria-hidden="true" />}
                  <p className="text-sm">
                    {isDirect
                      ? t('chat.dmStart', { name: (peer ? fullName(peer) : roomName) ?? t('rooms.fallbackName') })
                      : t('chat.roomStart', { room: roomName ?? t('chat.thisRoom') })}
                  </p>
                </li>
              )}
              {hasMore && (
                <li className="flex justify-center pb-2">
                  <Button variant="ghost" size="sm" onClick={handleLoadMore} disabled={loadingMore}>
                    {loadingMore ? t('common.loading') : t('chat.loadOlder')}
                  </Button>
                </li>
              )}
              {ordered.map((msg) => {
                const isCurrentUser = msg.author_id === profile?.id;
                const author = authorName(msg.author_id);
                const role = memberById.get(msg.author_id)?.role;
                const replied = msg.reply_to_message_id ? byId.get(msg.reply_to_message_id) : undefined;
                return (
                  <li key={msg.id} className={cn('group flex gap-2.5 py-1', isCurrentUser && 'flex-row-reverse')}>
                    <Avatar name={author} tone={isCurrentUser ? 'me' : 'other'} className="mt-0.5" />
                    <div className={cn('flex min-w-0 max-w-[85%] flex-col md:max-w-[72%]', isCurrentUser && 'items-end')}>
                      <div
                        className={cn(
                          'min-w-0 max-w-full rounded-xl border px-3.5 py-2.5 shadow-sm',
                          isCurrentUser ? 'rounded-tr-sm border-primary/25 bg-soft' : 'rounded-tl-sm bg-card'
                        )}
                      >
                        <div className="flex flex-wrap items-baseline gap-x-2">
                          {!isCurrentUser && !isDirect ? (
                            <button
                              type="button"
                              onClick={() => onOpenDirect(msg.author_id)}
                              title={t('chat.messageUser', { name: author })}
                              className="text-sm font-semibold hover:text-primary hover:underline"
                            >
                              {author}
                            </button>
                          ) : (
                            <span className="text-sm font-semibold">{author}</span>
                          )}
                          {roleLabel(role)}
                          <span className="font-mono text-[11px] text-muted-foreground">{formatTimestamp(msg.created_at, locale)}</span>
                          <div className="ml-auto flex gap-3 pl-3 text-[11px] text-muted-foreground md:opacity-0 md:transition-opacity md:focus-within:opacity-100 md:group-hover:opacity-100">
                            <button type="button" className={actionClass} onClick={() => setReplyingTo(msg)}>
                              <Reply className="h-3 w-3" /> {t('chat.reply')}
                            </button>
                            {canPin && (
                              <button type="button" className={actionClass} onClick={() => togglePin(msg.id, !msg.pinned_at)}>
                                {msg.pinned_at ? <PinOff className="h-3 w-3" /> : <Pin className="h-3 w-3" />}
                                {msg.pinned_at ? t('chat.unpin') : t('chat.pin')}
                              </button>
                            )}
                            {!isCurrentUser && !isDirect && (
                              <button type="button" className={actionClass} onClick={() => reportMessage(msg.id)}>
                                <Flag className="h-3 w-3" /> {t('chat.report')}
                              </button>
                            )}
                            {(isCurrentUser || canDeleteAnyMessage(myRank)) && (
                              <button
                                type="button"
                                className="flex items-center gap-1 hover:text-destructive"
                                onClick={() => window.confirm(t('chat.deleteConfirm')) && deleteMessage(msg.id)}
                              >
                                <Trash2 className="h-3 w-3" /> {t('chat.delete')}
                              </button>
                            )}
                          </div>
                        </div>
                        {msg.reply_to_message_id && (
                          <p className="mt-1.5 truncate rounded-md border-l-2 border-primary/40 bg-background/70 px-2.5 py-1 text-[13px] text-muted-foreground">
                            {replied ? `${authorName(replied.author_id)}: ${replied.content}` : t('chat.originalMessage')}
                          </p>
                        )}
                        <p className="mt-1 whitespace-pre-wrap break-words text-sm">{msg.content}</p>
                      </div>
                      {profile && (
                        <div className="mt-1.5">
                          <Reaction
                            messageId={msg.id}
                            reactions={msg.reactions ?? []}
                            onAddReaction={addReaction}
                            onRemoveReaction={removeReaction}
                            currentUserId={profile.id}
                            alignRight={isCurrentUser}
                          />
                        </div>
                      )}
                    </div>
                  </li>
                );
              })}
              {pending.map((item) => (
                <li key={item.tempId} className="flex flex-row-reverse gap-2.5 py-1">
                  <Avatar name={profile ? fullName(profile) : null} tone="me" className="mt-0.5" />
                  <div className="flex min-w-0 max-w-[85%] flex-col items-end md:max-w-[72%]">
                    <div
                      className={cn(
                        'min-w-0 max-w-full rounded-xl rounded-tr-sm border px-3.5 py-2.5',
                        item.status === 'sending' ? 'border-primary/25 bg-soft opacity-60' : 'border-destructive bg-card'
                      )}
                    >
                      <p className="whitespace-pre-wrap break-words text-sm">{item.content}</p>
                    </div>
                    {item.status === 'sending' ? (
                      <p className="mt-1 text-[11px] text-muted-foreground">{t('chat.sending')}</p>
                    ) : (
                      <p className="mt-1 flex gap-2 text-[11px] text-destructive" role="alert">
                        {t('chat.notSent')}
                        <button type="button" className="underline" onClick={() => retryMessage(item.tempId)}>
                          {t('chat.retry')}
                        </button>
                        <button type="button" className="underline" onClick={() => discardMessage(item.tempId)}>
                          {t('chat.discard')}
                        </button>
                      </p>
                    )}
                  </div>
                </li>
              ))}
            </ul>
          )}
          <div ref={messagesEndRef} />
        </div>

        {replyingTo && (
          <div className="flex shrink-0 items-center gap-2 border-t bg-soft px-4 py-2 text-xs md:px-6">
            <Reply className="h-3 w-3 shrink-0 text-primary" />
            <p className="min-w-0 flex-1 truncate">{t('chat.replyingTo', { content: replyingTo.content })}</p>
            <button type="button" onClick={() => setReplyingTo(null)} aria-label={t('chat.cancelReply')}>
              <X className="h-3 w-3" />
            </button>
          </div>
        )}
        <p className="h-5 shrink-0 px-4 text-xs text-muted-foreground md:px-6" aria-live="polite">
          {typingLabel(typingNames, t)}
        </p>
        <form onSubmit={handleSendMessage} className="shrink-0 px-3 pb-4 md:px-6 md:pb-5">
          <div className="flex h-12 items-center gap-2 rounded-lg border bg-card pl-4 pr-2 focus-within:ring-2 focus-within:ring-ring">
            <Input
              value={content}
              onChange={(e) => {
                setContent(e.target.value);
                if (e.target.value) notifyTyping();
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
      </div>

      {showMembers && !isDirect && (
        <aside
          aria-label={t('chat.members')}
          className="absolute inset-y-0 right-0 z-20 w-72 overflow-y-auto border-l bg-card p-4 shadow-lg lg:static lg:z-auto lg:shrink-0 lg:shadow-none"
        >
          <div className="mb-3 flex items-baseline justify-between">
            <h3 className="text-[13px] font-semibold">{t('chat.members')}</h3>
            <span className="font-mono text-[11px] text-muted-foreground">{members.length}</span>
          </div>
          {membersList}
        </aside>
      )}
    </div>
  );
}
