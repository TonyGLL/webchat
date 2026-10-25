import type { MessageKey } from '@/lib/i18n/en';
import { formatTitle } from '@/lib/alerts';

/** i18n key of the page title for a pathname, or null when the route has no specific title. */
export function titleKeyForPath(pathname: string): MessageKey | null {
  const segment = pathname.split('/').filter(Boolean)[0];
  switch (segment) {
    case 'login': return 'title.login';
    case 'register': return 'title.register';
    case 'forgot-password': return 'title.forgot';
    case 'reset-password': return 'title.reset';
    case 'dashboard': return 'title.dashboard';
    case 'profile': return 'title.profile';
    default: return null;
  }
}

/** "Page · WebChat", or just the app name when there is no page title. */
export function composeTitle(page: string | null, appName: string): string {
  return page ? `${page} · ${appName}` : appName;
}

// The tab title is the page title plus an optional unread-count prefix. Both pieces are kept
// here so whoever sets one never clobbers the other.
let base = '';
let unread = 0;

function apply() {
  if (typeof document !== 'undefined' && base) document.title = formatTitle(base, unread);
}

/** Sets the page part of the tab title (called by RouteTitle on route/language change). */
export function setBaseTitle(title: string) {
  base = title;
  apply();
}

/** Sets the unread-messages prefix, e.g. "(3) Chats · WebChat". 0 removes it. */
export function setUnreadCount(count: number) {
  unread = count;
  apply();
}
