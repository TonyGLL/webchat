import { useState, useEffect } from 'react';
import { roomService } from '@/services/roomService';

export const useRoom = () => {
  const [rooms, setRooms] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  const fetchRooms = async () => {
    try {
      const data = await roomService.getUserRooms();
      setRooms(data);
    } catch (error) {
      console.error('Failed to fetch rooms', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchRooms();
  }, []);

  const createRoom = async (data: { name: string; is_public: boolean, topic: string }) => {
    await roomService.createRoom(data);
    fetchRooms();
  };

  return { rooms, loading, createRoom };
};