import { cn } from '@/lib/utils';

export default function RoomList({
  rooms,
  onSelectRoom,
  selectedRoom,
}: {
  rooms: any[];
  onSelectRoom: (roomId: string) => void;
  selectedRoom: string | null;
}) {
  return (
    <div className="h-full overflow-y-auto">
      {rooms.map(room => (
        <div
          key={room.id}
          onClick={() => onSelectRoom(room.id)}
          className={cn(
            'p-4 cursor-pointer hover:bg-accent border-b border-border',
            selectedRoom === room.id ? 'bg-accent' : ''
          )}
        >
          {room.name}
        </div>
      ))}
    </div>
  );
}