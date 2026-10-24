'use client';

import { useState } from 'react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';
import UserProfile from '@/components/UserProfile';
import { cn } from '@/lib/utils';

export default function DashboardPage() {
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null);
  const { rooms, loading, createRoom, leaveRoom } = useRoom(selectedRoom);
  const room = rooms.find((r) => r.id === selectedRoom);

  // On small screens only one pane is visible: the room list, or the open chat.
  return (
    <PrivateRoute>
      <div className="flex h-dvh bg-background">
        <aside
          className={cn(
            'w-full flex-col border-r md:flex md:w-80 md:shrink-0',
            selectedRoom ? 'hidden' : 'flex'
          )}
        >
          <UserProfile />
          <div className="relative flex items-center justify-between px-4 pb-2 pt-4">
            <h2 className="text-sm font-medium text-muted-foreground">Chats</h2>
            <CreateRoomForm createRoom={createRoom} />
          </div>
          <div className="flex-1 overflow-y-auto">
            {loading ? (
              <div className="py-8">
                <Spinner />
              </div>
            ) : (
              <RoomList rooms={rooms} onSelectRoom={setSelectedRoom} selectedRoom={selectedRoom} />
            )}
          </div>
        </aside>
        <main className={cn('min-w-0 flex-1 flex-col md:flex', selectedRoom ? 'flex' : 'hidden')}>
          <MessageView
            roomId={selectedRoom}
            roomName={room?.name}
            canPin={room?.my_role === 'owner' || room?.my_role === 'admin'}
            onBack={() => setSelectedRoom(null)}
            onLeave={async () => {
              if (!selectedRoom) return;
              await leaveRoom(selectedRoom);
              setSelectedRoom(null);
            }}
          />
        </main>
      </div>
    </PrivateRoute>
  );
}
