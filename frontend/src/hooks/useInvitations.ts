import { useCallback, useState } from 'react';
import { Invitation, roomService } from '@/services/roomService';
import { usePolling } from './usePolling';
import { logError } from '@/lib/logError';

const POLL_MS = 15_000;

/** Invitations addressed to the current user, refreshed while the app is open and visible. */
export const useInvitations = (onAccepted?: (roomId: string) => void) => {
  const [invitations, setInvitations] = useState<Invitation[]>([]);
  const [error, setError] = useState(false);

  const refresh = useCallback(async () => {
    try {
      setInvitations(await roomService.getInvitations());
    } catch (e) {
      logError('Failed to load invitations', e);
    }
  }, []);

  usePolling(refresh, POLL_MS);

  const respond = async (invitation: Invitation, accept: boolean) => {
    setError(false);
    try {
      if (accept) await roomService.acceptInvitation(invitation.id);
      else await roomService.declineInvitation(invitation.id);
      setInvitations((prev) => prev.filter((i) => i.id !== invitation.id));
      if (accept) onAccepted?.(invitation.room_id);
    } catch {
      setError(true);
      refresh();
    }
  };

  return { invitations, error, respond };
};
