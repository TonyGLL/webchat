import { describe, expect, it, vi } from 'vitest';
import { createSocketEventBus } from './socketEvents';

const frame = (type: string, payload: unknown) => JSON.stringify({ type, payload });

describe('createSocketEventBus', () => {
  it('delivers the payload only to subscribers of that type', () => {
    const bus = createSocketEventBus();
    const presence = vi.fn();
    const typing = vi.fn();
    bus.subscribe('presence', presence);
    bus.subscribe('typing', typing);
    bus.emit(frame('presence', { user_id: 1, online: true }));
    expect(presence).toHaveBeenCalledWith({ user_id: 1, online: true });
    expect(typing).not.toHaveBeenCalled();
  });

  it('stops delivering after unsubscribe', () => {
    const bus = createSocketEventBus();
    const handler = vi.fn();
    const off = bus.subscribe('presence', handler);
    off();
    bus.emit(frame('presence', { user_id: 1, online: true }));
    expect(handler).not.toHaveBeenCalled();
  });

  it('parses a frame once however many subscribers there are', () => {
    const bus = createSocketEventBus();
    bus.subscribe('presence', vi.fn());
    bus.subscribe('presence', vi.fn());
    const parse = vi.spyOn(JSON, 'parse');
    bus.emit(frame('presence', { user_id: 2, online: false }));
    expect(parse).toHaveBeenCalledTimes(1);
    parse.mockRestore();
  });

  it('ignores malformed, binary and unknown frames', () => {
    const bus = createSocketEventBus();
    const handler = vi.fn();
    bus.subscribe('presence', handler);
    expect(() => {
      bus.emit('not json');
      bus.emit(new ArrayBuffer(2));
      bus.emit(frame('nope', {}));
      bus.emit(JSON.stringify({ type: 'presence' }));
    }).not.toThrow();
    expect(handler).not.toHaveBeenCalled();
  });

  it('survives a handler that throws or unsubscribes while emitting', () => {
    const bus = createSocketEventBus();
    const error = vi.spyOn(console, 'error').mockImplementation(() => {});
    const last = vi.fn();
    bus.subscribe('typing', () => {
      throw new Error('boom');
    });
    const off = bus.subscribe('typing', () => off());
    bus.subscribe('typing', last);
    bus.emit(frame('typing', { room_id: 'r', user_id: 3 }));
    bus.emit(frame('typing', { room_id: 'r', user_id: 3 }));
    expect(last).toHaveBeenCalledTimes(2);
    error.mockRestore();
  });
});
