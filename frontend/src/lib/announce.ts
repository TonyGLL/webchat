import type { Translate } from '@/lib/i18n';

/** Minimum time between two screen-reader announcements about incoming messages. */
export const ANNOUNCE_WINDOW_MS = 1000;

/**
 * Messages that arrived after `previousNewestId`, from a list kept newest first. `known` is false
 * when that id is not in the list (room switch, reload after a reconnect): the caller should just
 * remember the new newest id and announce nothing. Older pages loaded at the end of the list never
 * count as new because only the items in front of the previous newest are looked at.
 */
export function incomingSince<T extends { id: string }>(newestFirst: T[], previousNewestId: string | null): { known: boolean; fresh: T[] } {
  if (previousNewestId === null) return { known: false, fresh: [] };
  const index = newestFirst.findIndex((m) => m.id === previousNewestId);
  if (index < 0) return { known: false, fresh: [] };
  return { known: true, fresh: newestFirst.slice(0, index) };
}

/** One sentence for the live region: the author when there is a single message, a count otherwise. Empty when nothing arrived. */
export function announcementText(authors: string[], t: Translate): string {
  if (authors.length === 0) return '';
  if (authors.length === 1) return t('chat.newMessageFrom', { name: authors[0] });
  return t('chat.newMessages', { count: authors.length });
}
