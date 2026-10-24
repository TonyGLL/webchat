import { IGetRoomMessages } from '@/services/messageService';
import { Room } from '@/services/roomService';
import Spinner from './Spinner';

interface SearchResultsProps {
  query: string;
  results: IGetRoomMessages[] | null;
  loading: boolean;
  rooms: Room[];
  onSelectRoom: (roomId: string) => void;
}

export default function SearchResults({ query, results, loading, rooms, onSelectRoom }: SearchResultsProps) {
  if (loading && !results) {
    return (
      <div className="py-8">
        <Spinner />
      </div>
    );
  }
  if (!results || results.length === 0) {
    return <p className="px-6 py-8 text-center text-sm text-muted-foreground">No messages match “{query}”.</p>;
  }

  const roomName = (id: string) => rooms.find((r) => r.id === id)?.name ?? 'Room';
  return (
    <ul className="space-y-0.5 p-2" aria-label="Search results">
      {results.map((m) => (
        <li key={m.id}>
          <button
            type="button"
            onClick={() => onSelectRoom(m.room_id)}
            className="w-full rounded-lg px-3 py-2.5 text-left transition-colors hover:bg-accent"
          >
            <div className="flex items-baseline justify-between gap-2">
              <span className="truncate text-xs font-medium">{roomName(m.room_id)}</span>
              <span className="shrink-0 text-[11px] text-muted-foreground">
                {new Date(m.created_at).toLocaleDateString('en-US', { month: 'short', day: 'numeric' })}
              </span>
            </div>
            <p className="mt-0.5 line-clamp-2 break-words text-sm text-muted-foreground">{m.content}</p>
          </button>
        </li>
      ))}
    </ul>
  );
}
