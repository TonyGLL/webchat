import { memo } from 'react';
import { Flag, Pin, PinOff, Reply, Trash2 } from 'lucide-react';
import dayjs from 'dayjs';
import isToday from 'dayjs/plugin/isToday';
import { IGetRoomMessages } from '@/services/messageService';
import { RoomRole } from '@/lib/permissions';
import { cn } from '@/lib/utils';
import { intlLocale, Locale } from '@/lib/i18n';
import { useT } from '@/lib/i18n/useT';
import Avatar from './Avatar';
import Reaction from './Reaction';
dayjs.extend(isToday);

export const formatTimestamp = (date: string | Date, locale: Locale) => {
  const options: Intl.DateTimeFormatOptions = { hour: 'numeric', minute: 'numeric' };

  if (!dayjs(date).isToday()) {
    options.day = 'numeric';
    options.month = 'short';
    options.year = 'numeric';
  }

  return new Date(date).toLocaleTimeString(intlLocale(locale), options);
};

export function RoleLabel({ role }: { role?: RoomRole }) {
  const { t } = useT();
  if (!role || role === 'member') return null;
  return (
    <span className={cn('text-[10px] font-semibold uppercase tracking-wide', role === 'moderator' ? 'text-muted-foreground' : 'text-primary')}>
      {t(`role.${role}`)}
    </span>
  );
}

// min-h-6 px-1: at least a 24x24 CSS px target.
const actionClass = 'flex min-h-6 items-center gap-1 px-1 hover:text-foreground';

interface MessageItemProps {
  msg: IGetRoomMessages;
  /** Display name of the author. */
  author: string;
  role?: RoomRole;
  /** "Name: text" of the message this one replies to, or the "original message" fallback; null when it is not a reply. */
  replyPreview: string | null;
  isCurrentUser: boolean;
  isDirect: boolean;
  canPin: boolean;
  canDelete: boolean;
  currentUserId?: number;
  // These must keep a stable identity (see useStableCallback), or the memo is pointless.
  onOpenDirect: (userId: number) => void;
  onReply: (msg: IGetRoomMessages) => void;
  onTogglePin: (messageId: string, pin: boolean) => void;
  onReport: (messageId: string) => void;
  onDelete: (messageId: string) => void;
  onAddReaction: (messageId: string, emoji: string) => void;
  onRemoveReaction: (messageId: string, emoji: string) => void;
}

function MessageItem({
  msg,
  author,
  role,
  replyPreview,
  isCurrentUser,
  isDirect,
  canPin,
  canDelete,
  currentUserId,
  onOpenDirect,
  onReply,
  onTogglePin,
  onReport,
  onDelete,
  onAddReaction,
  onRemoveReaction,
}: MessageItemProps) {
  const { t, locale } = useT();

  return (
    <li
      className={cn(
        'group flex gap-2.5 py-1',
        // Skip layout and paint of rows far off screen. Not while one of its popups is open
        // (the reaction button reports aria-expanded): containment would clip the fixed emoji layer.
        '[content-visibility:auto] [contain-intrinsic-size:auto_96px] has-[[aria-expanded=true]]:[content-visibility:visible]',
        isCurrentUser && 'flex-row-reverse'
      )}
    >
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
                aria-label={t('chat.messageUser', { name: author })}
                className="text-sm font-semibold hover:text-primary hover:underline"
              >
                {author}
              </button>
            ) : (
              <span className="text-sm font-semibold">{author}</span>
            )}
            <RoleLabel role={role} />
            <time
              dateTime={new Date(msg.created_at).toISOString()}
              title={new Date(msg.created_at).toLocaleString(intlLocale(locale), { dateStyle: 'full', timeStyle: 'short' })}
              className="font-mono text-[11px] text-muted-foreground"
            >
              {formatTimestamp(msg.created_at, locale)}
            </time>
            {/* Hidden until hover or focus only where a pointer can hover; touch devices always show it. */}
            <div className="ml-auto flex gap-2 pl-3 text-[11px] text-muted-foreground [@media(hover:hover)]:opacity-0 [@media(hover:hover)]:transition-opacity [@media(hover:hover)]:focus-within:opacity-100 [@media(hover:hover)]:group-hover:opacity-100">
              <button type="button" className={actionClass} onClick={() => onReply(msg)} aria-label={t('chat.replyTo', { name: author })}>
                <Reply className="h-3 w-3" /> {t('chat.reply')}
              </button>
              {canPin && (
                <button
                  type="button"
                  className={actionClass}
                  onClick={() => onTogglePin(msg.id, !msg.pinned_at)}
                  aria-label={t(msg.pinned_at ? 'chat.unpinFrom' : 'chat.pinFrom', { name: author })}
                >
                  {msg.pinned_at ? <PinOff className="h-3 w-3" /> : <Pin className="h-3 w-3" />}
                  {msg.pinned_at ? t('chat.unpin') : t('chat.pin')}
                </button>
              )}
              {!isCurrentUser && !isDirect && (
                <button type="button" className={actionClass} onClick={() => onReport(msg.id)} aria-label={t('chat.reportFrom', { name: author })}>
                  <Flag className="h-3 w-3" /> {t('chat.report')}
                </button>
              )}
              {canDelete && (
                <button
                  type="button"
                  className="flex min-h-6 items-center gap-1 px-1 hover:text-destructive"
                  onClick={() => onDelete(msg.id)}
                  aria-label={t('chat.deleteFrom', { name: author })}
                >
                  <Trash2 className="h-3 w-3" /> {t('chat.delete')}
                </button>
              )}
            </div>
          </div>
          {replyPreview !== null && (
            <p className="mt-1.5 truncate rounded-md border-l-2 border-primary/40 bg-background/70 px-2.5 py-1 text-[13px] text-muted-foreground">
              <span className="sr-only">{t('chat.inReplyTo')} </span>
              {replyPreview}
            </p>
          )}
          <p className="mt-1 whitespace-pre-wrap break-words text-sm">{msg.content}</p>
        </div>
        {currentUserId !== undefined && (
          <div className="mt-1.5">
            <Reaction
              messageId={msg.id}
              reactions={msg.reactions ?? []}
              onAddReaction={onAddReaction}
              onRemoveReaction={onRemoveReaction}
              currentUserId={currentUserId}
              alignRight={isCurrentUser}
            />
          </div>
        )}
      </div>
    </li>
  );
}

export default memo(MessageItem);
