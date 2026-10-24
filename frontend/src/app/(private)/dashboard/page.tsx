'use client';

import { useEffect, useState } from 'react';
import { Search, X } from 'lucide-react';
import PrivateRoute from '@/components/PrivateRoute';
import { useRoom } from '@/hooks/useRoom';
import Spinner from '@/components/Spinner';
import RoomList from '@/components/RoomList';
import CreateRoomForm from '@/components/CreateRoomForm';
import MessageView from '@/components/MessageView';
import PeoplePanel from '@/components/PeoplePanel';
import DiscoverRooms from '@/components/DiscoverRooms';
import InvitationList from '@/components/InvitationList';
import { useDiscoverRooms } from '@/hooks/useDiscoverRooms';
import { useInvitations } from '@/hooks/useInvitations';
import { useOnboarding } from '@/hooks/useOnboarding';
import { useMessageAlerts } from '@/hooks/useMessageAlerts';
import { useUser } from '@/hooks/useUser';
import OnboardingDialog from '@/components/OnboardingDialog';
import UserProfile from '@/components/UserProfile';
import SearchResults from '@/components/SearchResults';
import { Input } from '@/components/ui/input';
import { useDebounce } from '@/lib/useDebounce';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

export default function DashboardPage() {
  const { t } = useT();
  const { profile } = useUser();
  const tour = useOnboarding(profile?.id);
  const [chosenRoom, setSelectedRoom] = useState<string | null>(null);
  const { rooms, loading, createRoom, leaveRoom, openDirect, joinRoom, refreshRooms } = useRoom(chosenRoom);
  const { rooms: discoverable, hasMore: discoverHasMore, loadingMore: discoverLoadingMore, loadMore: loadMoreDiscover, refresh: refreshDiscover } = useDiscoverRooms(rooms.length);
  const { invitations, error: invitationError, respond } = useInvitations(async (roomId) => {
    await refreshRooms();
    setSelectedRoom(roomId);
  });
  const openPerson = async (userId: number) => setSelectedRoom(await openDirect(userId));
  // A room we were removed from (or left elsewhere) disappears from the list: close its chat.
  const selectedRoom = loading || rooms.some((r) => r.id === chosenRoom) ? chosenRoom : null;
  const room = rooms.find((r) => r.id === selectedRoom);
  useMessageAlerts(rooms, selectedRoom, setSelectedRoom);

  const [search, setSearch] = useState('');
  const debouncedSearch = useDebounce(search.trim(), 300);
  const searching = search.trim().length >= 2;
  const [results, setResults] = useState<IGetRoomMessages[] | null>(null);
  const [searchLoading, setSearchLoading] = useState(false);

  useEffect(() => {
    if (debouncedSearch.length < 2) return; // results are hidden while the query is too short
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
      {tour.open && <OnboardingDialog name={profile?.name ?? profile?.username ?? ''} onFinish={tour.finish} />}
      <div className="flex h-dvh bg-background">
        <aside
          className={cn(
            'w-full flex-col border-r bg-card md:flex md:w-[272px] md:shrink-0',
            selectedRoom ? 'hidden' : 'flex'
          )}
        >
          <UserProfile onHelp={tour.reopen} />
          <div className="relative px-3 pb-1 pt-3">
            <Search className="pointer-events-none absolute left-6 top-[1.35rem] h-4 w-4 text-muted-foreground" />
            <Input
              value={search}
              onChange={(e) => setSearch(e.target.value)}
              placeholder={t('dashboard.searchPlaceholder')}
              aria-label={t('dashboard.searchPlaceholder')}
              className="h-9 bg-background pl-8 pr-8"
            />
            {search && (
              <button
                type="button"
                onClick={() => setSearch('')}
                aria-label={t('dashboard.clearSearch')}
                className="absolute right-5 top-[1.35rem] text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
          <InvitationList invitations={invitations} error={invitationError} onRespond={respond} />
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
              <>
                <RoomList
                  rooms={rooms}
                  onSelectRoom={setSelectedRoom}
                  selectedRoom={selectedRoom}
                  action={<CreateRoomForm createRoom={createRoom} />}
                />
                <DiscoverRooms
                  rooms={discoverable}
                  hasMore={discoverHasMore}
                  loadingMore={discoverLoadingMore}
                  onLoadMore={loadMoreDiscover}
                  onJoin={async (id) => {
                    setSelectedRoom(await joinRoom(id));
                    refreshDiscover();
                  }}
                />
                {/* Small screens have no side column: people live here instead. */}
                <div className="border-t lg:hidden">
                  <PeoplePanel onOpenDirect={openPerson} />
                </div>
              </>
            )}
          </div>
        </aside>
        <main className={cn('min-w-0 flex-1 flex-col md:flex', selectedRoom ? 'flex' : 'hidden')}>
          {/* key: every room starts with fresh state (draft, reply, pins, members, pending sends) */}
          <MessageView
            key={selectedRoom ?? 'none'}
            roomId={selectedRoom}
            roomName={room?.name}
            isDirect={room?.is_direct}
            onOpenDirect={async (userId) => setSelectedRoom(await openDirect(userId))}
            myRole={room?.my_role}
            onBack={() => setSelectedRoom(null)}
            onLeave={async () => {
              if (!selectedRoom) return;
              await leaveRoom(selectedRoom);
              setSelectedRoom(null);
            }}
          />
        </main>
        {/* Always visible on wide screens, whichever room (or none) is open. */}
        <aside className="hidden w-[248px] shrink-0 border-l bg-card lg:block">
          <PeoplePanel onOpenDirect={openPerson} />
        </aside>
      </div>
    </PrivateRoute>
  );
}
