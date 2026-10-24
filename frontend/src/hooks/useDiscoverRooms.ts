import { useCallback, useEffect, useState } from 'react';
import { RoomSummary, roomService } from '@/services/roomService';

/** Rooms the user is not in yet: public ones can be joined, private ones need an invitation. */
export const useDiscoverRooms = (refreshKey: unknown) => {
  const [rooms, setRooms] = useState<RoomSummary[]>([]);

  const refresh = useCallback(async () => {
    try {
      setRooms(await roomService.discoverRooms());
    } catch (error) {
      console.error('Failed to discover rooms', error);
    }
  }, []);

  // `refreshKey` changes when the user's own rooms do (joined, created, left).
  useEffect(() => {
    refresh();
  }, [refresh, refreshKey]);

  return { rooms: rooms.filter((r) => !r.is_member), refresh };
};
