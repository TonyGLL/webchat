import type { IGetRoomMessages, IReaction } from '@/services/messageService';
import type { RoomRole } from '@/lib/permissions';

/** Events the backend pushes over the websocket, keyed by `type` (see each module application package in the backend). */
export interface SocketEventMap {
  new_message: IGetRoomMessages;
  pin_message: { message_id: string; pinned: boolean };
  delete_message: { message_id: string };
  update_message: { message_id: string; content: string };
  reaction_added: IReaction & { message_id: string };
  reaction_removed: IReaction & { message_id: string };
  presence: { user_id: number; online: boolean };
  role_changed: { room_id: string; user_id: number; role: RoomRole };
  member_removed: { room_id: string; user_id: number };
  typing: { room_id: string; user_id: number };
}

export type SocketEventType = keyof SocketEventMap;
export type SocketEventHandler<T extends SocketEventType> = (payload: SocketEventMap[T]) => void;

export interface SocketEventBus {
  /** Listens to one event type; returns the function that stops listening. */
  subscribe<T extends SocketEventType>(type: T, handler: SocketEventHandler<T>): () => void;
  /** Parses a raw frame once and hands its payload to the subscribers of its type. Bad frames are ignored. */
  emit(raw: unknown): void;
}

export function createSocketEventBus(): SocketEventBus {
  const handlers = new Map<string, Set<(payload: never) => void>>();

  return {
    subscribe(type, handler) {
      let set = handlers.get(type);
      if (!set) handlers.set(type, (set = new Set()));
      const entry = handler as (payload: never) => void;
      set.add(entry);
      return () => {
        set.delete(entry);
      };
    },
    emit(raw) {
      if (typeof raw !== 'string') return;
      let frame: { type?: unknown; payload?: unknown } | null;
      try {
        frame = JSON.parse(raw);
      } catch {
        return;
      }
      if (typeof frame?.type !== 'string' || frame.payload == null) return;
      const set = handlers.get(frame.type);
      if (!set) return;
      // Copy: a handler may unsubscribe while we iterate. One failing handler must not starve the rest.
      for (const handler of Array.from(set)) {
        try {
          (handler as (payload: unknown) => void)(frame.payload);
        } catch (error) {
          console.error('Socket event handler failed', error);
        }
      }
    },
  };
}
