'use client';

import { useEffect, useState } from 'react';
import { Search, X } from 'lucide-react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';
import UserProfile from '@/components/UserProfile';
import SearchResults from '@/components/SearchResults';
import { Input } from '@/components/ui/input';
import { useDebounce } from '@/lib/useDebounce';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { cn } from '@/lib/utils';

export default function DashboardPage() {
  const [selectedRoom, setSelectedRoom] = useState<string | null>(null);
  const { rooms, loading, createRoom, leaveRoom, openDirect } = useRoom(selectedRoom);
  const room = rooms.find((r) => r.id === selectedRoom);

  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search.trim(), 300);
  const searching = search.trim().length >= 2;
  const [results, setResults] = useState<IGetRoomMessages[] | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);

  useEffect(() => {
    if (debouncedSearch.length < 2) {
      setResults(null);
      return;
    }
    let cancelled = false;
    setSearchLoading(true);
    messageService
      .search(debouncedSearch)
      .then((found) => !cancelled && setResults(found))
      .catch(() => !cancelled && setResults([]))
      .finally(() => !cancelled && setSearchLoading(false));
    return () => {
      cancelled = true;
    };
  }, [debouncedSearch]);

  const openFromSearch = (roomId: string) => {
    setSelectedRoom(roomId);
    setSearch('');
  };

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
          <div className="relative px-3 pb-2">
            <Search className="pointer-events-none absolute left-6 top-2.5 h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder="Search messages"
              aria-label="Search messages"
              className="h-9 pl-8 pr-8"
            />
            {search && (
              <button
                type="button"
                onClick={() => setSearch('')}
                aria-label="Clear search"
                className="absolute right-5 top-2.5 text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
          <div className="flex-1 overflow-y-auto">
            {searching ? (
              <SearchResults
                query={search.trim()}
                results={results}
                loading={searchLoading}
                rooms={rooms}
                onSelectRoom={openFromSearch}
              />
            ) : loading ? (
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
            isDirect={room?.is_direct}
            onOpenDirect={async (userId) => setSelectedRoom(await openDirect(userId))}
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
