'use client';

import { Search } from 'lucide-react';
import { useState } from 'react';
import Avatar from '@/components/Avatar';
import { Input } from '@/components/ui/input';
import { useDirectory } from '@/hooks/useDirectory';
import { DirectoryUser } from '@/services/userService';
import { cn } from '@/lib/utils';
import { useT } from '@/lib/i18n/useT';

export const fullName = (u: { name: string; last_name: string | null }) => [u.name, u.last_name].filter(Boolean).join(' ');

export function PresenceDot({ online, className }: { online: boolean; className?: string }) {
  const { t } = useT();
  return (
    <span
      role="img"
      aria-label={online ? t('chat.online') : t('chat.offline')}
      className={cn('h-2 w-2 shrink-0 rounded-full', online ? 'bg-success' : 'border border-presence-offline', className)}
    />
  );
}

/** Everybody on the server with live presence, always visible: click a person to write to them. */
export default function PeoplePanel({ onOpenDirect }: { onOpenDirect: (userId: number) => void }) {
  const { t } = useT();
  const [query, setQuery] = useState('');
  const { people, loading } = useDirectory(query);
  const online = people.filter((p) => p.online);
  const offline = people.filter((p) => !p.online);

  const group = (label: string, list: DirectoryUser[]) =>
    list.length > 0 && (
      <section>
        <h3 className="px-4 pb-1 pt-3 text-[11px] font-semibold uppercase tracking-wider text-muted-foreground">{label}</h3>
        <ul className="space-y-px px-2">
          {list.map((p) => (
            <li key={p.id}>
              <button
                type="button"
                onClick={() => onOpenDirect(p.id)}
                title={t('people.message', { name: fullName(p) })}
                className="flex h-9 w-full items-center gap-2.5 rounded-md px-2 text-left text-sm transition-colors hover:bg-accent"
              >
                <span className="relative shrink-0">
                  <Avatar name={fullName(p)} size="sm" />
                  <PresenceDot online={p.online} className="absolute -bottom-0.5 -right-0.5 h-2.5 w-2.5 border-2 border-card" />
                </span>
                <span className={cn('min-w-0 flex-1 truncate', !p.online && 'text-muted-foreground')}>{fullName(p)}</span>
                <span className="truncate font-mono text-[11px] text-muted-foreground">@{p.username}</span>
              </button>
            </li>
          ))}
        </ul>
      </section>
    );

  return (
    <div className="flex h-full flex-col">
      <div className="flex h-14 shrink-0 items-center border-b px-4">
        <h2 className="text-[13px] font-semibold">{t('people.title')}</h2>
      </div>
      <div className="relative px-3 pb-1 pt-3">
        <Search className="pointer-events-none absolute left-6 top-[1.35rem] h-4 w-4 text-muted-foreground" />
        <Input
          value={query}
          onChange={(e) => setQuery(e.target.value)}
          placeholder={t('people.search')}
          aria-label={t('people.search')}
          className="h-9 bg-background pl-8"
        />
      </div>
      <div className="flex-1 overflow-y-auto pb-4">
        {!loading && people.length === 0 && <p className="px-4 py-6 text-center text-sm text-muted-foreground">{t('people.empty')}</p>}
        {group(t('people.onlineGroup', { count: online.length }), online)}
        {group(t('people.offlineGroup', { count: offline.length }), offline)}
      </div>
    </div>
  );
}
