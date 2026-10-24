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
import { useMediaQuery } from '@/hooks/useMediaQuery';
import { useMessageAlerts } from '@/hooks/useMessageAlerts';
import { useUser } from '@/hooks/useUser';
import dynamic from 'next/dynamic';
import UserProfile from '@/components/UserProfile';
import SearchResults from '@/components/SearchResults';
import { Input } from '@/components/ui/input';
import { useDebounce } from '@/lib/useDebounce';
import { IGetRoomMessages, messageService } from '@/services/messageService';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

// Only shown on a user's first visit (or when reopened): keep it out of the first-load bundle.
const OnboardingDialog = dynamic(() => import('@/components/OnboardingDialog'), { ssr: false });

export default function DashboardPage() {
  const { t } = useT();
  const { profile } = useUser();
  // One people panel (and so one directory fetch loop): in the right column from `lg` up, in the sidebar below.
  const wide = useMediaQuery('(min-width: 1024px)');
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

  // Below md only one pane shows; with no room open the list pane has to be the page's main landmark.
  const isWide = useMediaQuery('(min-width: 768px)');
  const listIsMain = !isWide && !selectedRoom;

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
        <h1 className="sr-only">{t('dashboard.heading')}</h1>
        <aside
          aria-label={t('dashboard.sidebar')}
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
                className="absolute right-4 top-4 inline-flex h-7 w-7 items-center justify-center rounded-md text-muted-foreground hover:text-foreground"
              >
                <X className="h-4 w-4" />
              </button>
            )}
          </div>
          <InvitationList invitations={invitations} error={invitationError} onRespond={respond} />
          <div className="flex-1 overflow-y-auto" role={listIsMain ? 'main' : undefined}>
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
                {!wide && (
                  <div className="border-t">
                    <PeoplePanel onOpenDirect={openPerson} />
                  </div>
                )}
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
        {wide && (
          <aside aria-label={t('people.title')} className="w-[248px] shrink-0 border-l bg-card">
            <PeoplePanel onOpenDirect={openPerson} />
          </aside>
        )}
      </div>
    </PrivateRoute>
  );
}
