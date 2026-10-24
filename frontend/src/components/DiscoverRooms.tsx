'use client';

import { Lock } from 'lucide-react';
import { useState } from 'react';
import { Button } from '@/components/ui/button';
import { RoomSummary } from '@/services/roomService';
import { useT } from '@/lib/i18n/useT';

/** Rooms the user has not joined. Public ones open with a click; private ones are only listed. */
export default function DiscoverRooms({ rooms, onJoin }: { rooms: RoomSummary[]; onJoin: (roomId: string) => Promise<void> }) {
  const { t } = useT();
  const [busy, setBusy] = useState<string | null>(null);
  const [error, setError] = useState(false);

  if (rooms.length === 0) return null;

  const join = async (id: string) => {
    setBusy(id);
    setError(false);
    try {
      await onJoin(id);
    } catch {
      setError(true);
    } finally {
      setBusy(null);
    }
  };

  return (
    <section className="pb-4">
      <h2 className="px-4 pb-1 pt-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{t('discover.title')}</h2>
      <ul className="space-y-px px-2">
        {rooms.map((room) => (
          <li key={room.id} className="flex min-h-9 items-center gap-2 rounded-md px-2 py-1 text-sm">
            {room.is_private ? (
              <Lock className="h-3.5 w-3.5 shrink-0 text-muted-foreground" aria-hidden="true" />
            ) : (
              <span className="w-3.5 shrink-0 text-center font-mono font-medium text-muted-foreground" aria-hidden="true">
                #
              </span>
            )}
            <div className="min-w-0 flex-1">
              <p className="truncate">{room.name}</p>
              <p className="truncate text-[11px] text-muted-foreground">
                {room.is_private
                  ? t('discover.privateNote')
                  : t(room.member_count === 1 ? 'discover.members.one' : 'discover.members.other', { count: room.member_count })}
              </p>
            </div>
            {!room.is_private && (
              <Button size="sm" variant="outline" className="h-7 px-2.5 text-xs" disabled={busy === room.id} onClick={() => join(room.id)}>
                {t('discover.join')}
              </Button>
            )}
          </li>
        ))}
      </ul>
      {error && (
        <p className="px-4 pt-1 text-xs text-destructive" role="alert">
          {t('discover.joinFailed')}
        </p>
      )}
    </section>
  );
}
