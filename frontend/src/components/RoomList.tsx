import { MessageSquare } from 'lucide-react';
import { cn } from '@/lib/utils';

export default function RoomList({
  rooms,
  onSelectRoom,
  selectedRoom,
}: {
  rooms: { id: string; name: string }[];
  onSelectRoom: (id: string) => void;
  selectedRoom: string | null;
}) {
  if (rooms.length === 0) {
    return (
      <div className="flex flex-col items-center gap-2 px-6 py-12 text-center text-muted-foreground">
        <MessageSquare className="h-5 w-5" />
        <p className="text-sm">No rooms yet. Create your first one with +.</p>
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
              {room.name.charAt(0).toUpperCase()}
            </span>
            <span className="truncate text-sm font-medium">{room.name}</span>
          </button>
        </li>
      ))}
    </ul>
  );
}
