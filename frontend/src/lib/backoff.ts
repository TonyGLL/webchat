const BASE_MS = 1000;
const MAX_MS = 30000;
/** Close code of a server that is going away (restart, rolling deploy): every client reconnects at once. */
export const GOING_AWAY = 1012;
const GOING_AWAY_SPREAD_MS = 5000;

/** Delay before reconnect attempt number `attempt` (0-based): exponential, capped, with full jitter. */
export function backoffDelay(attempt: number, random: () => number = Math.random): number {
  const exponential = Math.min(MAX_MS, BASE_MS * 2 ** attempt);
  // Anywhere between 0 and the delay, so a herd of clients spreads out instead of retrying together.
  return Math.round(exponential * random());
}

/** Delay before reconnecting after a close with `code`: a server shutdown spreads clients over 0-5 s. */
export function reconnectDelay(attempt: number, code: number | undefined, random: () => number = Math.random): number {
  if (code === GOING_AWAY) return Math.round(GOING_AWAY_SPREAD_MS * random());
  return backoffDelay(attempt, random);
}
