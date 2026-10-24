import { describe, expect, it } from 'vitest';
import { countIncomingMessage } from './unread';

const rooms = [
  { id: 'a', unread_count: 0 },
  { id: 'b', unread_count: 2 },
];

describe('countIncomingMessage', () => {
  it('counts a message from someone else in a room that is not open', () => {
    const next = countIncomingMessage(rooms, { room_id: 'b', author_id: 9 }, 'a', 1);
    expect(next).toEqual([
      { id: 'a', unread_count: 0 },
      { id: 'b', unread_count: 3 },
    ]);
  });

  it('ignores the open room and the user\'s own messages', () => {
    expect(countIncomingMessage(rooms, { room_id: 'a', author_id: 9 }, 'a', 1)).toBe(rooms);
    expect(countIncomingMessage(rooms, { room_id: 'b', author_id: 1 }, 'a', 1)).toBe(rooms);
  });
});
