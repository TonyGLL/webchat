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
  return (
    <div className="space-y-1">
      {rooms.map((room) => (
        <div
          key={room.id}
          onClick={() => onSelectRoom(room.id)}
          className={cn(
            'flex items-center p-3 rounded-lg cursor-pointer transition-colors',
            selectedRoom === room.id
              ? 'bg-blue-500 text-white'
              : 'hover:bg-gray-200 dark:hover:bg-gray-700'
          )}
        >
          <div className="w-10 h-10 rounded-full bg-gray-300 dark:bg-gray-600 flex items-center justify-center font-bold text-gray-600 dark:text-gray-300 mr-3">
            {room.name.charAt(0).toUpperCase()}
          </div>
          <p className="font-semibold text-gray-800 dark:text-gray-200">{room.name}</p>
        </div>
      ))}
    </div>
  );
}