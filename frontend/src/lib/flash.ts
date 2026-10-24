/** One-shot notice carried across a client-side redirect ("account created" on the login page). */
const KEY = 'flash:registered';

export function setRegisteredFlash(storage: Pick<Storage, 'setItem'>): void {
  try {
    storage.setItem(KEY, '1');
  } catch {
    // storage blocked: the notice is simply skipped
  }
}

/** True once after `setRegisteredFlash`; the flag is consumed. */
export function takeRegisteredFlash(storage: Pick<Storage, 'getItem' | 'removeItem'>): boolean {
  try {
    const set = storage.getItem(KEY) === '1';
    if (set) storage.removeItem(KEY);
    return set;
  } catch {
    return false;
  }
}
