import { useEffect, useRef } from 'react';
import { decideAlert, previewText } from '@/lib/alerts';
import { setUnreadCount } from '@/lib/documentTitle';
import { playMessageSound } from '@/lib/sound';
import { useT } from '@/lib/i18n/useT';
import type { Room } from '@/services/roomService';
import { useAlertPrefs } from './useAlertPrefs';
import { useSocketEvent } from './useSocketEvents';
import { useUser } from './useUser';

const permissionNow = (): NotificationPermission | 'unsupported' =>
  typeof Notification === 'undefined' ? 'unsupported' : Notification.permission;

/**
 * Sound, desktop notifications and an unread counter in the tab title for incoming messages.
 * Everything is decided by `decideAlert`; this hook only wires it to the socket and the browser.
 */
export const useMessageAlerts = (rooms: Room[], activeRoomId: string | null, onOpenRoom: (roomId: string) => void) => {
  const { profile } = useUser();
  const { prefs } = useAlertPrefs();
  const { t } = useT();

  // Latest values for the long-lived socket listener, so it is not re-subscribed on every change.
  const latest = useRef({ rooms, activeRoomId, prefs, myId: profile?.id, onOpenRoom });
  useEffect(() => {
    latest.current = { rooms, activeRoomId, prefs, myId: profile?.id, onOpenRoom };
  });

  useSocketEvent('new_message', ({ room_id, author_id, content }) => {
    const current = latest.current;

    const decision = decideAlert({
      authorId: author_id,
      myId: current.myId,
      roomId: room_id,
      activeRoomId: current.activeRoomId,
      pageActive: document.visibilityState === 'visible' && document.hasFocus(),
      prefs: current.prefs,
      permission: permissionNow(),
    });
    const room = current.rooms.find((r) => r.id === room_id);

    if (decision.sound) playMessageSound(room?.is_direct ? 'direct' : 'room', room_id);
    if (decision.desktop) {
      try {
        const notification = new Notification(room?.name ?? t('common.appName'), {
          body: previewText(content),
          tag: room_id, // several messages from one room replace each other
        });
        notification.onclick = () => {
          window.focus();
          current.onOpenRoom(room_id);
          notification.close();
        };
      } catch {
        // some browsers only allow notifications from a service worker: the sound still played
      }
    }
  });

  // Unread total in the tab title, so a background tab shows there is something new.
  const unread = rooms.reduce((sum, r) => sum + (r.unread_count || 0), 0);
  useEffect(() => {
    setUnreadCount(unread);
    return () => setUnreadCount(0);
  }, [unread]);
};
