/** @type {import('next').NextConfig} */
const nextConfig = {
  // Self-contained server (server.js + only the node_modules it needs) for the Docker image.
  output: 'standalone',
};

export default nextConfig;
