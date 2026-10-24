import { MessageSquare, User } from 'lucide-react';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

export default function RoomList({
  rooms,
  onSelectRoom,
  selectedRoom,
}: {
  rooms: { id: string; name: string; unread_count?: number; is_direct?: boolean }[];
  onSelectRoom: (id: string) => void;
  selectedRoom: string | null;
}) {
  const { t } = useT();
  if (rooms.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 px-6 py-12 text-center text-muted-foreground">
        <MessageSquare className="h-5 w-5" />
        <p className="text-sm">{t('rooms.empty')}</p>
      </div>
    );
  }

  return (
    <ul className="space-y-0.5 p-2">
      {rooms.map((room) => (
        <li key={room.id}>
          <button
            type="button"
            onClick={() => onSelectRoom(room.id)}
            aria-current={selectedRoom === room.id ? 'true' : undefined}
            className={cn(
              'flex w-full items-center gap-3 rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-accent',
              selectedRoom === room.id && 'bg-accent'
            )}
          >
            <span className="flex h-9 w-9 shrink-0 items-center justify-center rounded-full bg-muted text-sm font-medium">
              {room.is_direct ? <User className="h-4 w-4" /> : room.name.charAt(0).toUpperCase()}
            </span>
            <span className="min-w-0 flex-1 truncate text-sm font-medium">{room.name}</span>
            {!!room.unread_count && selectedRoom !== room.id && (
              <span
                className="flex h-5 min-w-5 shrink-0 items-center justify-center rounded-full bg-primary px-1.5 text-[11px] font-medium text-primary-foreground"
                aria-label={t('rooms.unread', { count: room.unread_count })}
              >
                {room.unread_count > 99 ? '99+' : room.unread_count}
              </span>
            )}
          </button>
        </li>
      ))}
    </ul>
  );
}
