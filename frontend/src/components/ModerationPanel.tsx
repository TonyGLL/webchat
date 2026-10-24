import { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { RANK } from '@/lib/permissions';
import { AuditEntry, Ban, moderationService, Report } from '@/services/moderationService';
import { RoomMember } from '@/services/roomService';
import { cn } from '@/lib/utils';

type Tab = 'reports' | 'bans' | 'audit';

interface ModerationPanelProps {
  roomId: string;
  rank: number;
  members: RoomMember[];
}

const when = (iso: string) => new Date(iso).toLocaleString('en-US', { dateStyle: 'short', timeStyle: 'short' });

export default function ModerationPanel({ roomId, rank, members }: ModerationPanelProps) {
  const [tab, setTab] = useState<Tab>('reports');
  const [reports, setReports] = useState<Report[]>([]);
  const [bans, setBans] = useState<Ban[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const canSeeAudit = rank >= RANK.admin;

  const nameOf = useCallback(
    (id: number | null | undefined) => (id == null ? 'deleted user' : (members.find((m) => m.user_id === id)?.name ?? `user #${id}`)),
    [members]
  );

  const load = useCallback(async () => {
    setError(null);
    try {
      if (tab === 'reports') setReports(await moderationService.getReports(roomId));
      else if (tab === 'bans') setBans(await moderationService.getBans(roomId));
      else setAudit(await moderationService.getAudit(roomId));
    } catch {
      setError('Could not load this list.');
    }
  }, [tab, roomId]);

  useEffect(() => {
    load();
  }, [load]);

  const act = async (action: () => Promise<void>) => {
    setError(null);
    try {
      await action();
      await load();
    } catch {
      setError('That action is not allowed.');
    }
  };

  return (
    <section className="max-h-64 shrink-0 overflow-y-auto border-b bg-muted/40 px-4 py-2 text-xs" aria-label="Moderation">
      <div className="mb-2 flex gap-1" role="tablist">
        {(['reports', 'bans', ...(canSeeAudit ? ['audit'] : [])] as Tab[]).map((t) => (
          <button
            key={t}
            role="tab"
            aria-selected={tab === t}
            type="button"
            onClick={() => setTab(t)}
            className={cn('rounded px-2 py-1 capitalize', tab === t ? 'bg-primary text-primary-foreground' : 'hover:bg-accent')}
          >
            {t}
          </button>
        ))}
      </div>
      {error && (
        <p className="mb-1 text-destructive" role="alert">
          {error}
        </p>
      )}

      {tab === 'reports' &&
        (reports.length === 0 ? (
          <p className="text-muted-foreground">No pending reports.</p>
        ) : (
          <ul className="space-y-2">
            {reports.map((r) => (
              <li key={r.id} className="rounded border bg-background p-2">
                <p className="text-muted-foreground">
                  {nameOf(r.reported_by_id)} reported {nameOf(r.message_author_id)} · {when(r.created_at)}
                </p>
                <p className="my-1 break-words font-medium">“{r.message_content}”</p>
                <p className="mb-1 break-words">Reason: {r.reason}</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => act(() => moderationService.resolveReport(r.id, 'resolve', true))}>
                    Delete message
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => act(() => moderationService.resolveReport(r.id, 'resolve'))}>
                    Mark resolved
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => act(() => moderationService.resolveReport(r.id, 'ignore'))}>
                    Ignore
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        ))}

      {tab === 'bans' &&
        (bans.length === 0 ? (
          <p className="text-muted-foreground">Nobody is banned.</p>
        ) : (
          <ul className="space-y-1">
            {bans.map((b) => (
              <li key={b.user_id} className="flex items-center gap-2">
                <span className="min-w-0 flex-1 truncate">
                  {nameOf(b.user_id)} · {b.expires_at ? `until ${when(b.expires_at)}` : 'permanent'}
                  {b.reason && <span className="text-muted-foreground"> · {b.reason}</span>}
                </span>
                <button type="button" className="underline hover:text-foreground" onClick={() => act(() => moderationService.unban(roomId, b.user_id))}>
                  Unban
                </button>
              </li>
            ))}
          </ul>
        ))}

      {tab === 'audit' &&
        (audit.length === 0 ? (
          <p className="text-muted-foreground">Nothing recorded yet.</p>
        ) : (
          <ul className="space-y-1">
            {audit.map((e) => (
              <li key={e.id} className="truncate">
                <span className="text-muted-foreground">{when(e.created_at)}</span> {nameOf(e.actor_id)} · {e.action}
                {e.target_user_id != null && <> → {nameOf(e.target_user_id)}</>}
              </li>
            ))}
          </ul>
        ))}
    </section>
  );
}
