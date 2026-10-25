import { act, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AuthContext } from '@/contexts/AuthContext';
import { SocketContext } from '@/contexts/SocketContext';
import { createSocketEventBus } from '@/lib/socketEvents';
import { cleanup, flush, mount } from '@/test/harness';

const mocks = vi.hoisted(() => ({
  getUserRooms: vi.fn(),
  discoverRooms: vi.fn(),
  getInvitations: vi.fn(),
  getDirectory: vi.fn(),
  getMembers: vi.fn(),
  markRead: vi.fn(),
}));

vi.mock('next/navigation', () => ({ useRouter: () => ({ replace: vi.fn(), push: vi.fn() }) }));
vi.mock('@/services/roomService', () => ({
  roomService: {
    getUserRooms: mocks.getUserRooms,
    discoverRooms: mocks.discoverRooms,
    getInvitations: mocks.getInvitations,
    getMembers: mocks.getMembers,
    markRead: mocks.markRead,
  },
}));
vi.mock('@/services/userService', () => ({ userService: { getDirectory: mocks.getDirectory } }));
vi.mock('@/services/messageService', () => ({ messageService: { getMessages: vi.fn(), getPinned: vi.fn(), search: vi.fn() } }));
vi.mock('@/services/moderationService', () => ({ moderationService: {} }));
vi.mock('@/components/UserProfile', () => ({ default: () => null }));
vi.mock('@/hooks/useOnboarding', () => ({ useOnboarding: () => ({ open: false, finish: vi.fn(), reopen: vi.fn() }) }));

import DashboardPage from './page';

const user = { id: 1, name: 'Me', last_name: 'X', username: 'me', email: 'm@x', avatar_url: null, role: 'user', created_at: '' };
const requests = () => ({
  rooms: mocks.getUserRooms.mock.calls.length,
  discover: mocks.discoverRooms.mock.calls.length,
  invitations: mocks.getInvitations.mock.calls.length,
  directory: mocks.getDirectory.mock.calls.length,
});

beforeEach(() => {
  vi.clearAllMocks();
  mocks.getUserRooms.mockResolvedValue([
    { id: 'r1', name: 'general', unread_count: 0, my_role: 'member' },
    { id: 'r2', name: 'random', unread_count: 0, my_role: 'member' },
  ]);
  mocks.discoverRooms.mockResolvedValue({ rooms: [], hasMore: false });
  mocks.getInvitations.mockResolvedValue([]);
  mocks.getDirectory.mockResolvedValue([]);
  mocks.getMembers.mockResolvedValue([]);
  mocks.markRead.mockResolvedValue(undefined);
  window.matchMedia = (() => ({ matches: false, addEventListener() {}, removeEventListener() {} })) as never;
});
afterEach(cleanup);

describe('cold dashboard load', () => {
  it('fires nothing before the session exists, then exactly one request per resource', async () => {
    const events = createSocketEventBus();
    let auth = { user: null, tokens: null, loading: true } as never;
    let socket: WebSocket | null = null;
    const wrapper = ({ children }: { children: ReactNode }) => (
      <AuthContext.Provider value={auth}>
        <SocketContext.Provider value={{ socket, events }}>{children}</SocketContext.Provider>
      </AuthContext.Provider>
    );
    const view = await mount(<DashboardPage />, wrapper);
    await flush();
    expect(requests()).toEqual({ rooms: 0, discover: 0, invitations: 0, directory: 0 });

    // The session is restored (one refresh): the dashboard mounts and loads each resource once.
    auth = { user, tokens: { accessToken: 't' }, loading: false } as never;
    await view.rerender(<DashboardPage />);
    await flush();
    await flush();
    expect(requests()).toEqual({ rooms: 1, discover: 1, invitations: 1, directory: 1 });

    // The socket opens: nothing is refetched (only a reconnect after a drop reloads).
    socket = {} as WebSocket;
    await view.rerender(<DashboardPage />);
    await act(async () => new Promise((resolve) => setTimeout(resolve, 2500)));
    expect(requests()).toEqual({ rooms: 1, discover: 1, invitations: 1, directory: 1 });
  });
});
