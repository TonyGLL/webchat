'use client';

import { useState } from 'react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';
import UserProfile from '@/components/UserProfile';

export default function DashboardPage() {
  const { rooms, loading, createRoom } = useRoom();
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null);

  return (
    <PrivateRoute>
      <div className="grid md:grid-cols-[300px_1fr] h-screen bg-background">
        <aside className="flex flex-col bg-gray-100 dark:bg-gray-800 border-r border-gray-200 dark:border-gray-700">
          <UserProfile />
          <div className="p-4 border-b border-gray-200 dark:border-gray-700">
            <h2 className="text-lg font-semibold text-gray-800 dark:text-gray-200">Chats</h2>
          </div>
          <div className="flex-1 p-2 overflow-y-auto">
            {loading ? (
              <div className="flex justify-center items-center h-full">
                <Spinner />
              </div>
            ) : (
              <RoomList rooms={rooms} onSelectRoom={setSelectedRoom} selectedRoom={selectedRoom} />
            )}
          </div>
          <div className="p-4 border-t border-gray-200 dark:border-gray-700">
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