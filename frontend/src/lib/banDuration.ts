const UNIT = '(?:ns|us|µs|ms|s|m|h)';

/**
 * Whether `value` is empty (permanent ban) or a positive Go duration such as `24h`, `90m` or `1h30m`,
 * the format the backend parses. The server still enforces the upper limit.
 */
export function isValidBanDuration(value: string): boolean {
  const v = value.trim();
  if (v === '') return true;
  if (!new RegExp(`^(?:\\d+(?:\\.\\d+)?${UNIT})+$`).test(v)) return false;
  return /[1-9]/.test(v);
}
