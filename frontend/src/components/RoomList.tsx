export default function RoomList({ rooms, onSelectRoom }: { rooms: any[], onSelectRoom: (roomId: string) => void }) {
  return (
    <div className="h-full overflow-y-auto">
      {rooms.map((room) => (
        <div
          key={room.id}
          onClick={() => onSelectRoom(room.id)}
          className="p-4 cursor-pointer hover:bg-gray-100 border-b border-gray-200"
        >
          {room.name}
        </div>
      ))}
    </div>
  );
}