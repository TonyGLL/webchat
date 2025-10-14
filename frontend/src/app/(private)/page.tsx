'use client';

import { useState } from 'react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import { useSocket } from '@/hooks/useSocket';
import { useAuth } from '@/hooks/useAuth';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';

export default function ChatPage() {
  const { rooms, loading, createRoom } = useRoom();
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null);
  const { tokens } = useAuth();
  const socket = useSocket(tokens?.accessToken || null);

  return (
    <PrivateRoute>
      <div className="flex h-screen bg-white">
        <div className="w-1/4 border-r border-gray-200 flex flex-col">
          <div className="p-4 border-b border-gray-200">
            <h1 className="text-xl font-bold">Rooms</h1>
          </div>
          <div className="flex-1">
            {loading ? <Spinner /> : <RoomList rooms={rooms} onSelectRoom={setSelectedRoom} />}
          </div>
          <CreateRoomForm createRoom={createRoom} />
        </div>
        <div className="w-3/4">
          <MessageView roomId={selectedRoom} socket={socket} />
        </div>
      </div>
    </PrivateRoute>
  );
}