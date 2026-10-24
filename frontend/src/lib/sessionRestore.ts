export type RestoreOutcome = 'restored' | 'rejected' | 'unavailable';

/** The server answered and said this session is no longer valid (as opposed to a network failure). */
export function isSessionRejected(error: unknown): boolean {
  const status = (error as { response?: { status?: number } } | null)?.response?.status;
  return status === 401 || status === 403;
}

const wait = (ms: number) => new Promise<void>((resolve) => setTimeout(resolve, ms));

/**
 * Asks for a new access token, retrying transient failures (offline laptop just woken up,
 * backend restarting). Only an explicit rejection means the session is over: any other failure
 * leaves the session alone so the user is not signed out by a hiccup.
 */
export async function restoreSession(
  refresh: () => Promise<unknown>,
  { attempts = 3, delayMs = 800, sleep = wait }: { attempts?: number; delayMs?: number; sleep?: (ms: number) => Promise<void> } = {}
): Promise<RestoreOutcome> {
  for (let attempt = 1; attempt <= attempts; attempt++) {
    try {
      await refresh();
      return 'restored';
    } catch (error) {
      if (isSessionRejected(error)) return 'rejected';
      if (attempt < attempts) await sleep(delayMs * attempt);
    }
  }
  return 'unavailable';
}
