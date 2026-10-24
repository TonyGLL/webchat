export interface AlertPrefs {
  /** Short sound when a message arrives that you are not already looking at. */
  sound: boolean;
  /** Operating-system notification when the page is in the background. Needs the browser's permission. */
  desktop: boolean;
}

export const ALERT_PREFS_KEY = 'alertPrefs';
export const DEFAULT_ALERT_PREFS: AlertPrefs = { sound: true, desktop: false };

/** Reads stored preferences defensively: anything unexpected falls back to the defaults. */
export function parseAlertPrefs(raw: string | null | undefined): AlertPrefs {
  if (!raw) return DEFAULT_ALERT_PREFS;
  try {
    const value = JSON.parse(raw) as Partial<AlertPrefs>;
    return {
      sound: typeof value.sound === 'boolean' ? value.sound : DEFAULT_ALERT_PREFS.sound,
      desktop: typeof value.desktop === 'boolean' ? value.desktop : DEFAULT_ALERT_PREFS.desktop,
    };
  } catch {
    return DEFAULT_ALERT_PREFS;
  }
}

export interface AlertContext {
  authorId: number;
  myId: number | undefined;
  roomId: string;
  activeRoomId: string | null;
  /** The tab is visible AND the window has focus: the person is looking at the app. */
  pageActive: boolean;
  prefs: AlertPrefs;
  permission: NotificationPermission | 'unsupported';
}

/**
 * What to do when a message arrives. Never for your own messages, never for the conversation you
 * are reading right now, and operating-system notifications only when the page is in the
 * background (a sound is enough while you are looking at another room).
 */
export function decideAlert({ authorId, myId, roomId, activeRoomId, pageActive, prefs, permission }: AlertContext): {
  sound: boolean;
  desktop: boolean;
} {
  if (myId === undefined || authorId === myId) return { sound: false, desktop: false };
  if (pageActive && roomId === activeRoomId) return { sound: false, desktop: false };
  return {
    sound: prefs.sound,
    desktop: prefs.desktop && permission === 'granted' && !pageActive,
  };
}

/** "(3) WebChat" while there are unread messages, plain "WebChat" otherwise. */
export function formatTitle(base: string, unread: number): string {
  if (unread <= 0) return base;
  return `(${unread > 99 ? '99+' : unread}) ${base}`;
}

/** One line of a message for a notification body. */
export function previewText(content: string, max = 120): string {
  const flat = content.replace(/\s+/g, ' ').trim();
  return flat.length > max ? `${flat.slice(0, max - 1)}…` : flat;
}
