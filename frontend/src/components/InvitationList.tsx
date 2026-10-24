'use client';

import { Mail } from 'lucide-react';
import { Button } from '@/components/ui/button';
import { Invitation } from '@/services/roomService';
import { useT } from '@/lib/i18n/useT';

/** Invitations other people sent to this user, waiting for an answer. */
export default function InvitationList({
  invitations,
  error,
  onRespond,
}: {
  invitations: Invitation[];
  error: boolean;
  onRespond: (invitation: Invitation, accept: boolean) => void;
}) {
  const { t } = useT();
  if (invitations.length === 0) return null;

  return (
    <section className="border-b bg-soft px-3 py-3" aria-label={t('invitations.title')}>
      <h2 className="mb-2 flex items-center gap-1.5 px-1 text-[11px] font-semibold uppercase tracking-wider text-primary">
        <Mail className="h-3 w-3" aria-hidden="true" />
        {t('invitations.title')} · {invitations.length}
      </h2>
      <ul className="space-y-2">
        {invitations.map((inv) => (
          <li key={inv.id} className="rounded-lg border bg-card p-2.5">
            <p className="text-[13px]">{t('invitations.from', { name: inv.inviter_name, room: inv.room_name })}</p>
            <div className="mt-2 flex gap-2">
              <Button size="sm" className="h-7 flex-1 text-xs" onClick={() => onRespond(inv, true)}>
                {t('invitations.accept')}
              </Button>
              <Button size="sm" variant="outline" className="h-7 flex-1 text-xs" onClick={() => onRespond(inv, false)}>
                {t('invitations.decline')}
              </Button>
            </div>
          </li>
        ))}
      </ul>
      {error && (
        <p className="px-1 pt-2 text-xs text-destructive" role="alert">
          {t('invitations.failed')}
        </p>
      )}
    </section>
  );
}
