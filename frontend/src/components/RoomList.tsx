'use client';

import { Lock, MessageSquare } from 'lucide-react';
import Avatar from '@/components/Avatar';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

interface RoomItem {
  id: string;
  name: string;
  unread_count?: number;
  is_direct?: boolean;
  is_private?: boolean;
}

function SectionHeader({ label, action }: { label: string; action?: React.ReactNode }) {
  return (
    <div className="relative flex h-8 items-center justify-between px-4 pt-3">
      <h2 className="text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{label}</h2>
      {action}
    </div>
  );
}

export default function RoomList({
  rooms,
  onSelectRoom,
  selectedRoom,
  action,
}: {
  rooms: RoomItem[];
  onSelectRoom: (id: string) => void;
  selectedRoom: string | null;
  /** Control shown in the "Rooms" header (the new-room button). */
  action?: React.ReactNode;
}) {
  const { t } = useT();
  const channels = rooms.filter((r) => !r.is_direct);
  const directs = rooms.filter((r) => r.is_direct);

  const renderRoom = (room: RoomItem) => {
    const selected = selectedRoom === room.id;
    const unread = !!room.unread_count && !selected;
    const Prefix = room.is_private && !room.is_direct ? Lock : null;
    return (
      <li key={room.id}>
        <button
          type="button"
          onClick={() => onSelectRoom(room.id)}
          aria-current={selected ? 'true' : undefined}
          className={cn(
            'flex h-8 w-full items-center gap-2 rounded-md px-2 text-left text-sm transition-colors',
            selected
              ? 'bg-soft font-semibold text-primary'
              : unread
                ? 'font-semibold hover:bg-accent'
                : 'text-muted-foreground hover:bg-accent hover:text-foreground'
          )}
        >
          {room.is_direct ? (
            <Avatar name={room.name} size="xs" />
          ) : Prefix ? (
            <Prefix className="h-3.5 w-3.5 shrink-0" aria-hidden="true" />
          ) : (
            <span className="w-3.5 shrink-0 text-center font-mono font-medium text-muted-foreground" aria-hidden="true">
              #
            </span>
          )}
          <span className="min-w-0 flex-1 truncate">{room.name}</span>
          {unread && (
            <span
              className="flex h-[18px] min-w-[18px] shrink-0 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-semibold text-primary-foreground"
              aria-label={t('rooms.unread', { count: room.unread_count! })}
            >
              {room.unread_count! > 99 ? '99+' : room.unread_count}
            </span>
          )}
        </button>
      </li>
    );
  };

  return (
    <div className="pb-4">
      <SectionHeader label={t('rooms.section')} action={action} />
      {channels.length === 0 ? (
        <div className="flex flex-col items-center gap-2 px-6 py-8 text-center text-muted-foreground">
          <MessageSquare className="h-5 w-5" aria-hidden="true" />
          <p className="text-sm">{t('rooms.empty')}</p>
        </div>
      ) : (
        <ul className="space-y-px px-2">{channels.map(renderRoom)}</ul>
      )}
      {directs.length > 0 && (
        <>
          <SectionHeader label={t('rooms.direct')} />
          <ul className="space-y-px px-2">{directs.map(renderRoom)}</ul>
        </>
      )}
    </div>
  );
}
