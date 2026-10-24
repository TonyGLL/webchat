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
    "frame-ancestors 'none'",
    "base-uri 'self'",
    "form-action 'self'",
  ].join('; ');
}

export function securityHeaders({ apiUrl, isProduction }) {
  const headers = [
    { key: 'Content-Security-Policy', value: buildCsp({ apiUrl, isProduction }) },
    { key: 'X-Frame-Options', value: 'DENY' },
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
