// Security headers for every response. Kept out of next.config.mjs so it can be unit-tested.

const DEFAULT_API_URL = 'http://localhost:8081/api/v1';

/** Origins the browser may talk to: the API (http/https) and its WebSocket twin (ws/wss). Never throws. */
export function apiConnectSources(rawUrl) {
  let url;
  try {
    url = new URL(rawUrl || DEFAULT_API_URL);
  } catch {
    url = new URL(DEFAULT_API_URL);
  }
  if (url.protocol !== 'http:' && url.protocol !== 'https:') url = new URL(DEFAULT_API_URL);
  const ws = url.protocol === 'https:' ? 'wss:' : 'ws:';
  return [url.origin, `${ws}//${url.host}`];
}

const LOCAL_HOSTS = new Set(['localhost', '127.0.0.1', '[::1]']);

/**
 * A production bundle must not send tokens and cookies in clear text: the API URL has to be
 * https, or a localhost host (the default, also used by CI image builds). Throws otherwise.
 */
export function assertProductionApiUrl(rawUrl) {
  if (!rawUrl) return; // falls back to the localhost default
  let url;
  try {
    url = new URL(rawUrl);
  } catch {
    throw new Error(`NEXT_PUBLIC_API_URL is not a valid absolute URL: "${rawUrl}"`);
  }
  if (url.protocol === 'https:') return;
  if (url.protocol === 'http:' && LOCAL_HOSTS.has(url.hostname)) return;
  throw new Error(`NEXT_PUBLIC_API_URL must be https (or a localhost host) in a production build, got "${rawUrl}"`);
}

export function buildCsp({ apiUrl, isProduction }) {
  // The theme/locale init script in app/layout.tsx is inline, and Next injects inline bootstrap
  // scripts. A nonce would force every page to render dynamically (and needs a proxy), so
  // 'unsafe-inline' stays for scripts. 'unsafe-eval' is only for React Refresh in dev.
  const script = ["'self'", "'unsafe-inline'", ...(isProduction ? [] : ["'unsafe-eval'"])];
  return [
    "default-src 'self'",
    `script-src ${script.join(' ')}`,
    "style-src 'self' 'unsafe-inline'", // emoji-mart injects a <style> into its shadow root
    "img-src 'self' data:",
    "font-src 'self' data:",
    `connect-src ${["'self'", ...apiConnectSources(apiUrl)].join(' ')}`,
    "object-src 'none'",
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join('; ');
}

export function securityHeaders({ apiUrl, isProduction }) {
  const headers = [
    { key: 'Content-Security-Policy', value: buildCsp({ apiUrl, isProduction }) },
    { key: 'X-Frame-Options', value: 'DENY' },
    { key: 'Cross-Origin-Opener-Policy', value: 'same-origin' },
    { key: 'X-Content-Type-Options', value: 'nosniff' },
    // The password-reset token travels in the URL path: never leak it through Referer.
    { key: 'Referrer-Policy', value: 'no-referrer' },
    // Notifications stay allowed (message alerts).
    { key: 'Permissions-Policy', value: 'camera=(), microphone=(), geolocation=()' },
  ];
  // Browsers only honour HSTS over HTTPS; sending it in production is harmless behind plain HTTP.
  if (isProduction) {
    headers.push({ key: 'Strict-Transport-Security', value: 'max-age=63072000; includeSubDomains' });
  }
  return headers;
}
