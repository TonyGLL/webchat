interface ErrorLike {
  message?: unknown;
  response?: { status?: unknown };
  config?: { url?: unknown; method?: unknown };
}

/** Request path without origin, query or fragment (they can carry tokens). */
function endpointPath(url: unknown): string | undefined {
  if (typeof url !== 'string' || !url) return undefined;
  return url.replace(/^[a-z][a-z0-9+.-]*:\/\/[^/?#]+/i, '').split(/[?#]/)[0] || '/';
}

/**
 * Safe one-line summary of an error: message, HTTP status and endpoint path only.
 * Axios errors carry the whole request config (password in `config.data`, the Authorization
 * header), so the error object itself must never reach the console.
 */
export function describeError(error: unknown): string {
  if (error === null || typeof error !== 'object') return typeof error === 'string' ? error : 'unknown error';
  const e = error as ErrorLike;
  const parts: string[] = [typeof e.message === 'string' ? e.message : 'unknown error'];
  const status = e.response?.status;
  if (typeof status === 'number') parts.push(`status ${status}`);
  const path = endpointPath(e.config?.url);
  if (path) parts.push(`${typeof e.config?.method === 'string' ? e.config.method.toUpperCase() + ' ' : ''}${path}`);
  return parts.join(' | ');
}

export function logError(context: string, error: unknown) {
  console.error(`${context}: ${describeError(error)}`);
}
