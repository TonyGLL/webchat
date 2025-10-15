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
      <div className="grid md:grid-cols-[350px_1fr] h-screen bg-whatsapp-deep-sea-green text-white">
        <aside className="flex flex-col bg-whatsapp-surfie-green border-r border-gray-700">
          <UserProfile />
          <div className="p-4 border-b border-gray-700">
            <h2 className="text-xl font-semibold">Chats</h2>
          </div>
          <div className="flex-1 overflow-y-auto">
            {loading ? (
              <div className="flex justify-center items-center h-full">
                <Spinner />
              </div>
            ) : (
              <RoomList rooms={rooms} onSelectRoom={setSelectedRoom} selectedRoom={selectedRoom} />
            )}
          </div>
          <div className="p-4 border-t border-gray-700">
            <CreateRoomForm createRoom={createRoom} />
          </div>
        </aside>
        <main className="flex flex-col h-screen bg-whatsapp-pearl-bush bg-opacity-10">
          {/* Add this style for the background image */}
          <style jsx global>{`
            main {
              background-image: url('/whatsapp-bg.png');
              background-repeat: repeat;
            }
          `}</style>
          <MessageView roomId={selectedRoom} />
        </main>
      </div>
    </PrivateRoute>
  );
}