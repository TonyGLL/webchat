import { useCallback, useEffect, useState } from 'react';
import { Button } from '@/components/ui/button';
import { RANK } from '@/lib/permissions';
import { AuditEntry, Ban, moderationService, Report } from '@/services/moderationService';
import { RoomMember } from '@/services/roomService';
import { cn } from '@/lib/utils';
import { intlLocale } from '@/lib/i18n';
import { useT } from '@/lib/i18n/useT';

type Tab = 'reports' | 'bans' | 'audit';

interface ModerationPanelProps {
  roomId: string;
  rank: number;
  members: RoomMember[];
}

export default function ModerationPanel({ roomId, rank, members }: ModerationPanelProps) {
  const { t, locale } = useT();
  const when = useCallback(
    (iso: string) => new Date(iso).toLocaleString(intlLocale(locale), { dateStyle: 'short', timeStyle: 'short' }),
    [locale]
  );
  const [tab, setTab] = useState<Tab>('reports');
  const [reports, setReports] = useState<Report[]>([]);
  const [bans, setBans] = useState<Ban[]>([]);
  const [audit, setAudit] = useState<AuditEntry[]>([]);
  const [error, setError] = useState<string | null>(null);
  const canSeeAudit = rank >= RANK.admin;

  const nameOf = useCallback(
    (id: number | null | undefined) => (id == null ? t('common.deletedUser') : (members.find((m) => m.user_id === id)?.name ?? t('common.userNumber', { id }))),
    [members, t]
  );

  const load = useCallback(async () => {
    setError(null);
    try {
      if (tab === 'reports') setReports(await moderationService.getReports(roomId));
      else if (tab === 'bans') setBans(await moderationService.getBans(roomId));
      else setAudit(await moderationService.getAudit(roomId));
    } catch {
      setError(t('mod.loadError'));
    }
  }, [tab, roomId, t]);

  useEffect(() => {
    load();
  }, [load]);

  const act = async (action: () => Promise<void>) => {
    setError(null);
    try {
      await action();
      await load();
    } catch {
      setError(t('chat.actionNotAllowed'));
    }
  };

  return (
    <section className="max-h-64 shrink-0 overflow-y-auto border-b bg-muted/40 px-4 py-2 text-xs" aria-label={t('chat.moderation')}>
      <div className="mb-2 flex gap-1" role="tablist">
        {(['reports', 'bans', ...(canSeeAudit ? ['audit'] : [])] as Tab[]).map((name) => (
          <button
            key={name}
            role="tab"
            aria-selected={tab === name}
            type="button"
            onClick={() => setTab(name)}
            className={cn('rounded px-2 py-1 capitalize', tab === name ? 'bg-primary text-primary-foreground' : 'hover:bg-accent')}
          >
            {t(`mod.${name}`)}
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
          <p className="text-muted-foreground">{t('mod.noReports')}</p>
        ) : (
          <ul className="space-y-2">
            {reports.map((r) => (
              <li key={r.id} className="rounded border bg-background p-2">
                <p className="text-muted-foreground">
                  {t('mod.reported', { reporter: nameOf(r.reported_by_id), author: nameOf(r.message_author_id), when: when(r.created_at) })}
                </p>
                <p className="my-1 break-words font-medium">“{r.message_content}”</p>
                <p className="mb-1 break-words">{t('mod.reason', { reason: r.reason })}</p>
                <div className="flex gap-2">
                  <Button size="sm" variant="outline" onClick={() => act(() => moderationService.resolveReport(r.id, 'resolve', true))}>
                    {t('mod.deleteMessage')}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => act(() => moderationService.resolveReport(r.id, 'resolve'))}>
                    {t('mod.markResolved')}
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => act(() => moderationService.resolveReport(r.id, 'ignore'))}>
                    {t('mod.ignore')}
                  </Button>
                </div>
              </li>
            ))}
          </ul>
        ))}

      {tab === 'bans' &&
        (bans.length === 0 ? (
          <p className="text-muted-foreground">{t('mod.noBans')}</p>
        ) : (
          <ul className="space-y-1">
            {bans.map((b) => (
              <li key={b.user_id} className="flex items-center gap-2">
                <span className="min-w-0 flex-1 truncate">
                  {nameOf(b.user_id)} · {b.expires_at ? t('mod.until', { when: when(b.expires_at) }) : t('mod.permanent')}
                  {b.reason && <span className="text-muted-foreground"> · {b.reason}</span>}
                </span>
                <button type="button" className="underline hover:text-foreground" onClick={() => act(() => moderationService.unban(roomId, b.user_id))}>
                  {t('mod.unban')}
                </button>
              </li>
            ))}
          </ul>
        ))}

      {tab === 'audit' &&
        (audit.length === 0 ? (
          <p className="text-muted-foreground">{t('mod.noAudit')}</p>
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
