'use client';

import { X } from 'lucide-react';
import { useState } from 'react';
import Avatar from '@/components/Avatar';
import { fullName, PresenceDot } from '@/components/PeoplePanel';
import { Button } from '@/components/ui/button';
import { Input } from '@/components/ui/input';
import { useDirectory } from '@/hooks/useDirectory';
import { roomService } from '@/services/roomService';
import { useT } from '@/lib/i18n/useT';

/** Picker to send personal invitations to a room; shown to the people who may invite. */
export default function InviteUsers({
  roomId,
  roomName,
  memberIds,
  onClose,
}: {
  roomId: string;
  roomName: string;
  memberIds: number[];
  onClose: () => void;
}) {
  const { t } = useT();
  const [query, setQuery] = useState('');
  const { people } = useDirectory(query);
  const [sent, setSent] = useState<number[]>([]);
  const [error, setError] = useState(false);
  const candidates = people.filter((p) => !memberIds.includes(p.id));

  const invite = async (userId: number) => {
    setError(false);
    try {
      await roomService.inviteUser(roomId, userId);
      setSent((prev) => [...prev, userId]);
    } catch {
      setError(true);
    }
  };

  return (
    <div className="shrink-0 border-b bg-card px-4 py-3 md:px-6">
      <div className="mb-2 flex items-center justify-between">
        <h3 className="text-[13px] font-semibold">{t('invite.title', { room: roomName })}</h3>
        <button type="button" onClick={onClose} aria-label={t('common.cancel')} className="text-muted-foreground hover:text-foreground">
          <X className="h-4 w-4" />
        </button>
      </div>
      <Input value={query} onChange={(e) => setQuery(e.target.value)} placeholder={t('people.search')} aria-label={t('people.search')} className="mb-2 h-9" />
      <ul className="max-h-48 space-y-px overflow-y-auto">
        {candidates.length === 0 && <li className="py-2 text-sm text-muted-foreground">{t('people.empty')}</li>}
        {candidates.map((p) => (
          <li key={p.id} className="flex items-center gap-2.5 rounded-md px-1 py-1 text-sm">
            <Avatar name={fullName(p)} size="sm" />
            <span className="min-w-0 flex-1 truncate">
              {fullName(p)} <span className="font-mono text-[11px] text-muted-foreground">@{p.username}</span>
            </span>
            <PresenceDot online={p.online} />
            {sent.includes(p.id) ? (
              <span className="w-16 text-right text-xs text-success">{t('invite.sent')}</span>
            ) : (
              <Button size="sm" variant="outline" className="h-7 w-16 px-2 text-xs" onClick={() => invite(p.id)}>
                {t('invite.send')}
              </Button>
            )}
          </li>
        ))}
      </ul>
      {error && (
        <p className="pt-2 text-xs text-destructive" role="alert">
          {t('invite.failed')}
        </p>
      )}
    </div>
  );
}
