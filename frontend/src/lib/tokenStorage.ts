export interface AuthTokens {
  accessToken: string;
  refreshToken: string;
}

const STORAGE_KEY = 'authTokens';

type Listener = (tokens: AuthTokens | null) => void;
const listeners = new Set<Listener>();

function notify(tokens: AuthTokens | null) {
  listeners.forEach((listener) => listener(tokens));
}

/**
 * The only place that knows where auth tokens live. Today that is localStorage;
 * moving to httpOnly cookies later (H-004) only requires changing this file.
 */
export const tokenStorage = {
  get(): AuthTokens | null {
    if (typeof window === 'undefined') return null;
    try {
      const raw = window.localStorage.getItem(STORAGE_KEY);
      return raw ? (JSON.parse(raw) as AuthTokens) : null;
    } catch {
      return null;
    }
  },

  set(tokens: AuthTokens) {
    window.localStorage.setItem(STORAGE_KEY, JSON.stringify(tokens));
    notify(tokens);
  },

  clear() {
    window.localStorage.removeItem(STORAGE_KEY);
    notify(null);
  },

  /** Subscribe to token changes (login, silent refresh, logout). Returns an unsubscribe function. */
  subscribe(listener: Listener) {
    listeners.add(listener);
    return () => {
      listeners.delete(listener);
    };
  },
};
