'use client';

import { useState } from 'react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';

export default function DashboardPage() {
  const { rooms, loading, createRoom } = useRoom();
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null);

  return (
    <PrivateRoute>
      <div className="grid grid-cols-[260px_1fr] h-screen bg-background">
        <aside className="flex flex-col border-r border-border">
          <div className="p-4 border-b border-border">
            <h1 className="text-xl font-bold text-foreground">Rooms</h1>
          </div>
          <div className="flex-1 p-4">
            {loading ? (
              <Spinner />
            ) : (
              <RoomList rooms={rooms} onSelectRoom={setSelectedRoom} selectedRoom={selectedRoom} />
            )}
          </div>
          <div className="p-4 border-t border-border">
            <CreateRoomForm createRoom={createRoom} />
          </div>
        </aside>
        <main className="flex flex-col h-screen">
          <MessageView roomId={selectedRoom} />
        </main>
      </div>
    </PrivateRoute>
  );
}