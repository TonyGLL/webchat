'use client';

import { SocketProvider } from "@/contexts/SocketContext";

export default function PrivateLayout({ children }: { children: React.ReactNode }) {
  return <SocketProvider>{children}</SocketProvider>;
}
