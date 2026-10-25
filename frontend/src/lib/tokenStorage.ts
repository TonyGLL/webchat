export interface AuthTokens {
  accessToken: string;
}

const SESSION_HINT_KEY = 'hasSession';

type Listener = (tokens: AuthTokens | null) => void;
const listeners = new Set<Listener>();

// The access token lives only in memory: an XSS cannot read it from storage, and a page
// reload gets a new one through the httpOnly refresh cookie (see AuthContext).
let current: AuthTokens | null = null;

// Bumped on every clear(): an operation that started before a sign-out can tell it is stale.
let epoch = 0;

function notify(tokens: AuthTokens | null) {
  listeners.forEach((listener) => listener(tokens));
}

function writeHint(value: boolean) {
  try {
    if (value) window.localStorage.setItem(SESSION_HINT_KEY, '1');
    else window.localStorage.removeItem(SESSION_HINT_KEY);
  } catch {
    // storage unavailable: the hint is only an optimisation
  }
}

/**
 * The only place that knows where the access token lives. The refresh token is an httpOnly
 * cookie set by the backend, so JavaScript never sees it (H-004).
 */
export const tokenStorage = {
  get(): AuthTokens | null {
    return current;
  },

  set(tokens: AuthTokens) {
    current = tokens;
    writeHint(true);
    notify(tokens);
  },

  clear() {
    epoch++;
    current = null;
    writeHint(false);
    notify(null);
  },

  /** Changes whenever the session is cleared; compare before and after an async step. */
  epoch(): number {
    return epoch;
  },

  /**
   * Signs this tab out when another tab removes the session hint (logout or rejected refresh
   * there). Returns the function that stops listening.
   */
  watchRemoteSignOut(): () => void {
    const onStorage = (event: StorageEvent) => {
      const removed = event.key === SESSION_HINT_KEY ? event.newValue === null : event.key === null; // null key: storage cleared
      if (removed && current) tokenStorage.clear();
    };
    window.addEventListener('storage', onStorage);
    return () => window.removeEventListener('storage', onStorage);
  },

  /**
   * Non-secret flag telling whether this browser probably has a session (a refresh cookie),
   * so visitors do not trigger a pointless refresh request on every page load.
   */
  mayHaveSession(): boolean {
    try {
      return window.localStorage.getItem(SESSION_HINT_KEY) === '1';
    } catch {
      return false;
    }
  },

  /** Subscribe to token changes (login, silent refresh, logout). Returns an unsubscribe function. */
  subscribe(listener: Listener) {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
};
