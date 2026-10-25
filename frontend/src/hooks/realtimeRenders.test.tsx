import { act, type ReactNode } from 'react';
import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';
import { AuthContext } from '@/contexts/AuthContext';
import { SocketContext } from '@/contexts/SocketContext';
import { createSocketEventBus } from '@/lib/socketEvents';
import { cleanup, flush, renderHookCounting } from '@/test/harness';

const mocks = vi.hoisted(() => ({
  getMessages: vi.fn(),
  getPinned: vi.fn(),
  getMembers: vi.fn(),
  getDirectory: vi.fn(),
  getInvitations: vi.fn(),
  getUserRooms: vi.fn(),
  markRead: vi.fn(),
}));

vi.mock('@/services/messageService', () => ({ messageService: { getMessages: mocks.getMessages, getPinned: mocks.getPinned } }));
vi.mock('@/services/roomService', () => ({
  roomService: { getMembers: mocks.getMembers, getInvitations: mocks.getInvitations, getUserRooms: mocks.getUserRooms, markRead: mocks.markRead },
}));
vi.mock('@/services/userService', () => ({ userService: { getDirectory: mocks.getDirectory } }));
vi.mock('@/services/moderationService', () => ({ moderationService: {} }));

import { useMessages } from './useMessages';
import { useRoomActivity } from './useRoomActivity';
import { useDirectory } from './useDirectory';
import { useInvitations } from './useInvitations';
import { useRoom } from './useRoom';

const user = { id: 1, name: 'Me', last_name: 'X', username: 'me', email: 'm@x', avatar_url: null, role: 'user', created_at: '' };
const auth = { user, tokens: { accessToken: 't' }, loading: false } as never;
const frame = (type: string, payload: unknown) => JSON.stringify({ type, payload });
const messages = (n: number) =>
  Array.from({ length: n }, (_, i) => ({ id: `m${n - i}`, content: `hello ${n - i}`, author_id: 2, room_id: 'r1', created_at: '2026-10-24T10:00:00Z', reactions: [] }));

function setup() {
  const events = createSocketEventBus();
  const wrapper = ({ children }: { children: ReactNode }) => (
    <AuthContext.Provider value={auth}>
      <SocketContext.Provider value={{ socket: null, events }}>{children}</SocketContext.Provider>
    </AuthContext.Provider>
  );
  return { events, wrapper };
}

beforeEach(() => {
  vi.clearAllMocks();
  mocks.getMessages.mockResolvedValue(messages(50));
  mocks.getPinned.mockResolvedValue([]);
  mocks.getMembers.mockResolvedValue([{ user_id: 2, username: 'bob', name: 'Bob', last_name: null, avatar_url: null, role: 'member', online: false }]);
  mocks.getDirectory.mockResolvedValue([{ id: 2, username: 'bob', name: 'Bob', last_name: null, avatar_url: null, online: false }]);
  mocks.getInvitations.mockResolvedValue([]);
  mocks.getUserRooms.mockResolvedValue([{ id: 'r1', name: 'general', unread_count: 0, my_role: 'member' }]);
  mocks.markRead.mockResolvedValue(undefined);
});
afterEach(async () => {
  vi.useRealTimers();
  vi.restoreAllMocks();
  await cleanup();
});

describe('events that change nothing do not re-render', () => {
  const noops: [string, unknown][] = [
    ['reaction_added', { message_id: 'other-room-msg', user_id: 5, emoji: '👍' }],
    ['reaction_removed', { message_id: 'other-room-msg', user_id: 5, emoji: '👍' }],
    ['update_message', { message_id: 'other-room-msg', content: 'x' }],
    ['delete_message', { message_id: 'other-room-msg' }],
    ['pin_message', { message_id: 'other-room-msg', pinned: true }],
  ];

  it.each(noops)('useMessages ignores %s of a message it does not hold', async (type, payload) => {
    const { events, wrapper } = setup();
    const hook = await renderHookCounting(() => useMessages('r1'), wrapper);
    await flush();
    expect(hook.state.current.messages).toHaveLength(50);
    await act(async () => events.emit(frame(type, payload))); // React may render once more before it bails out
    const base = hook.state.renders;
    const before = hook.state.current.messages;
    for (let i = 0; i < 3; i++) await act(async () => events.emit(frame(type, payload)));
    expect(hook.state.renders - base).toBe(0);
    expect(hook.state.current.messages).toBe(before);
  });

  it('foreign pin_message events coalesce into one later refetch of the pinned list, without a re-render when unchanged', async () => {
    const { events, wrapper } = setup();
    const hook = await renderHookCounting(() => useMessages('r1'), wrapper);
    await flush();
    vi.useFakeTimers();
    mocks.getPinned.mockClear();
    const pinnedBefore = hook.state.current.pinned;
    await act(async () => {
      for (let i = 0; i < 5; i++) events.emit(frame('pin_message', { message_id: `x${i}`, pinned: true }));
    });
    expect(mocks.getPinned).not.toHaveBeenCalled();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(2000);
    });
    expect(mocks.getPinned).toHaveBeenCalledTimes(1);
    expect(hook.state.current.pinned).toBe(pinnedBefore); // same list: kept, not replaced
  });

  it('useRoomActivity ignores presence of a non-member and a repeated status', async () => {
    const { events, wrapper } = setup();
    const hook = await renderHookCounting(() => useRoomActivity('r1'), wrapper);
    await flush();
    await act(async () => events.emit(frame('presence', { user_id: 999, online: true }))); // warm-up (see above)
    const base = hook.state.renders;
    await act(async () => events.emit(frame('presence', { user_id: 999, online: true })));
    await act(async () => events.emit(frame('presence', { user_id: 2, online: false })));
    expect(hook.state.renders - base).toBe(0);
    await act(async () => events.emit(frame('presence', { user_id: 2, online: true })));
    expect(hook.state.renders - base).toBe(1);
  });

  it('useDirectory ignores presence that changes nobody', async () => {
    const { events, wrapper } = setup();
    const hook = await renderHookCounting(() => useDirectory(''), wrapper);
    await flush();
    await act(async () => events.emit(frame('presence', { user_id: 999, online: true }))); // warm-up (see above)
    const base = hook.state.renders;
    await act(async () => events.emit(frame('presence', { user_id: 2, online: false })));
    await act(async () => events.emit(frame('presence', { user_id: 999, online: true })));
    expect(hook.state.renders - base).toBe(0);
  });

  it('useInvitations keeps its state when a poll returns the same invitations', async () => {
    const same = [{ id: 'i1', room_id: 'r', room_name: 'n', inviter_id: 3, inviter_name: 'Z', invitee_id: 1, created_at: '' }];
    mocks.getInvitations.mockImplementation(async () => structuredClone(same));
    const { wrapper } = setup();
    const hook = await renderHookCounting(() => useInvitations(), wrapper);
    await flush();
    const first = hook.state.current.invitations;
    const poll = async (minutesLater: number) => {
      vi.spyOn(Date, 'now').mockReturnValue(Date.now() + minutesLater * 60_000); // past the polling gap
      await act(async () => {
        document.dispatchEvent(new Event('visibilitychange'));
      });
      await flush();
    };
    await poll(1); // warm-up (React may render once more before it bails out)
    const base = hook.state.renders;
    const calls = mocks.getInvitations.mock.calls.length;
    await poll(2);
    expect(mocks.getInvitations.mock.calls.length).toBeGreaterThan(calls);
    expect(hook.state.current.invitations).toBe(first);
    expect(hook.state.renders - base).toBe(0);
  });
});

describe('useRoom', () => {
  const message = (id: string) => frame('new_message', { id, content: 'x', author_id: 2, room_id: 'r1', created_at: '2026-10-24T10:00:00Z' });

  it('marks read once, trailing, for a burst of incoming messages in the open room', async () => {
    const { events, wrapper } = setup();
    await renderHookCounting(() => useRoom('r1'), wrapper);
    await flush();
    expect(mocks.markRead).toHaveBeenCalledTimes(1); // opening the room
    vi.useFakeTimers();
    await act(async () => {
      for (let i = 0; i < 50; i++) events.emit(message(`n${i}`));
    });
    expect(mocks.markRead).toHaveBeenCalledTimes(1);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });
    expect(mocks.markRead).toHaveBeenCalledTimes(2);
    await act(async () => {
      await vi.advanceTimersByTimeAsync(20_000);
    });
    expect(mocks.markRead).toHaveBeenCalledTimes(2);
  });

  it('does not mark read while the tab is hidden, and does when it comes back', async () => {
    const { events, wrapper } = setup();
    await renderHookCounting(() => useRoom('r1'), wrapper);
    await flush();
    vi.useFakeTimers();
    const visibility = vi.spyOn(document, 'visibilityState', 'get').mockReturnValue('hidden');
    await act(async () => events.emit(message('n1')));
    await act(async () => {
      await vi.advanceTimersByTimeAsync(5000);
    });
    expect(mocks.markRead).toHaveBeenCalledTimes(1);
    visibility.mockReturnValue('visible');
    await act(async () => {
      document.dispatchEvent(new Event('visibilitychange'));
      await vi.advanceTimersByTimeAsync(5000);
    });
    expect(mocks.markRead).toHaveBeenCalledTimes(2);
    visibility.mockRestore();
  });

  it('bumps membershipVersion only after the first load', async () => {
    const { events, wrapper } = setup();
    const hook = await renderHookCounting(() => useRoom(null), wrapper);
    await flush();
    expect(hook.state.current.membershipVersion).toBe(0);
    await act(async () => events.emit(frame('member_removed', { room_id: 'r1', user_id: 1 })));
    expect(hook.state.current.membershipVersion).toBe(1);
  });
});

describe('reconnect', () => {
  it('refetches after a drop and reopen, after a random delay, and not on the first open', async () => {
    const events = createSocketEventBus();
    let socket: WebSocket | null = null;
    const wrapper = ({ children }: { children: ReactNode }) => (
      <AuthContext.Provider value={auth}>
        <SocketContext.Provider value={{ socket, events }}>{children}</SocketContext.Provider>
      </AuthContext.Provider>
    );
    const hook = await renderHookCounting(() => useDirectory(''), wrapper);
    await flush();
    expect(mocks.getDirectory).toHaveBeenCalledTimes(1);
    vi.useFakeTimers();
    socket = {} as WebSocket; // first open
    await hook.refresh();
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mocks.getDirectory).toHaveBeenCalledTimes(1);
    socket = null; // drop
    await hook.refresh();
    socket = {} as WebSocket; // reopen
    await hook.refresh();
    expect(mocks.getDirectory).toHaveBeenCalledTimes(1); // delayed, not immediate
    await act(async () => {
      await vi.advanceTimersByTimeAsync(3000);
    });
    expect(mocks.getDirectory).toHaveBeenCalledTimes(2);
  });
});
