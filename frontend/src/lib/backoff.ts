const BASE_MS = 1000;
const MAX_MS = 30000;

/** Delay before reconnect attempt number `attempt` (0-based): exponential, capped, with jitter. */
export function backoffDelay(attempt: number, random: () => number = Math.random): number {
  const exponential = Math.min(MAX_MS, BASE_MS * 2 ** attempt);
  // 50–100 % of the delay, so many clients do not reconnect at the same instant
  return Math.round(exponential * (0.5 + random() / 2));
}
