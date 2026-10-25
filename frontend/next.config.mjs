import { PHASE_PRODUCTION_BUILD } from 'next/constants.js';
import { assertProductionApiUrl, securityHeaders } from './security-headers.mjs';

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Self-contained server (server.js + only the node_modules it needs) for the Docker image.
  output: 'standalone',
  poweredByHeader: false,
  // Nothing uses next/image: no optimizer endpoint and no sharp/libvips in the standalone output.
  images: { unoptimized: true },
  // Next's image optimizer is traced into the standalone output even when unused: drop sharp
  // and its 18 MB libvips binary (pnpm keeps them under node_modules/.pnpm).
  outputFileTracingExcludes: {
    '*': ['node_modules/.pnpm/@img+sharp*/**', 'node_modules/.pnpm/sharp@*/**', 'node_modules/@img/**', 'node_modules/sharp/**'],
  },
  async redirects() {
    // Signed-out and signed-in visitors both start at /login, which sends a restored session to
    // the dashboard; doing it here keeps the app JS out of `/`.
    return [{ source: '/', destination: '/login', permanent: false }];
  },
  async headers() {
    // NEXT_PUBLIC_API_URL is read at build time, same as the browser bundle that uses it.
    const headers = securityHeaders({
      apiUrl: process.env.NEXT_PUBLIC_API_URL,
      isProduction: process.env.NODE_ENV === 'production',
    });
    return [{ source: '/:path*', headers }];
  },
};

export default (phase) => {
  if (phase === PHASE_PRODUCTION_BUILD) assertProductionApiUrl(process.env.NEXT_PUBLIC_API_URL);
  return nextConfig;
};
