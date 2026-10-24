import { useCallback, useEffect, useRef, useState } from 'react';
import { RoomSummary, roomService } from '@/services/roomService';

/** Pages fetched in a row while a page holds only rooms the user already belongs to. */
const MAX_EMPTY_PAGES = 5;

/** Rooms the user is not in yet: public ones can be joined, private ones need an invitation. */
export const useDiscoverRooms = (refreshKey: unknown) => {
  const [all, setAll] = useState<RoomSummary[]>([]);
  const [hasMore, setHasMore] = useState(false);
  const [loadingMore, setLoadingMore] = useState(false);
  // Latest raw list, so loadMore does not depend on (and re-create with) the state.
  const allRef = useRef<RoomSummary[]>([]);

  const fetchPages = useCallback(async (reset: boolean) => {
    let acc = reset ? [] : allRef.current;
    let more = true;
    for (let i = 0; more && i < MAX_EMPTY_PAGES; i++) {
      const after = acc.length > 0 ? acc[acc.length - 1].id : undefined;
      const page = await roomService.discoverRooms(after);
      acc = [...acc, ...page.rooms];
      more = page.hasMore;
      if (page.rooms.some((r) => !r.is_member)) break;
    }
    allRef.current = acc;
    setAll(acc);
    setHasMore(more);
  }, []);

  const refresh = useCallback(async () => {
    try {
      await fetchPages(true);
    } catch (error) {
      console.error('Failed to discover rooms', error);
    }
  }, [fetchPages]);

  const loadMore = useCallback(async () => {
    setLoadingMore(true);
    try {
      await fetchPages(false);
    } catch (error) {
      console.error('Failed to load more rooms', error);
    } finally {
      setLoadingMore(false);
    }
  }, [fetchPages]);

  // `refreshKey` changes when the user's own rooms do (joined, created, left).
  useEffect(() => {
    refresh();
  }, [refresh, refreshKey]);

  return { rooms: all.filter((r) => !r.is_member), hasMore, loadingMore, loadMore, refresh };
};
