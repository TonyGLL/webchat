import { securityHeaders } from './security-headers.mjs';

/** @type {import('next').NextConfig} */
const nextConfig = {
  // Self-contained server (server.js + only the node_modules it needs) for the Docker image.
  output: 'standalone',
  poweredByHeader: false,
  async headers() {
    // NEXT_PUBLIC_API_URL is read at build time, same as the browser bundle that uses it.
    const headers = securityHeaders({
      apiUrl: process.env.NEXT_PUBLIC_API_URL,
      isProduction: process.env.NODE_ENV === 'production',
    });
    return [{ source: '/:path*', headers }];
  },
};

export default nextConfig;
