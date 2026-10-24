/** Rooms with the unread counter, as far as this helper cares. */
interface CountedRoom {
  id: string;
  unread_count: number;
}

/**
 * Updates unread counters for a message that just arrived over the websocket.
 * Messages from the user themselves, or in the room being viewed, are not unread.
 */
export function countIncomingMessage<T extends CountedRoom>(
  rooms: T[],
  message: { room_id: string; author_id: number },
  activeRoomId: string | null,
  myUserId: number | undefined
): T[] {
  if (message.room_id === activeRoomId || message.author_id === myUserId) return rooms;
  return rooms.map((room) =>
    room.id === message.room_id ? { ...room, unread_count: room.unread_count + 1 } : room
  );
}
