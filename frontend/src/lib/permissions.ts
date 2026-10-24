/**
 * Mirror of backend/internal/shared/permissions. The backend is the authority (it re-checks
 * every request); this only decides which buttons to show.
 */
export type RoomRole = 'owner' | 'admin' | 'moderator' | 'member';
export type GlobalRole = 'user' | 'moderator' | 'admin';

export const RANK = { none: 0, member: 1, moderator: 2, admin: 3, owner: 4, super: 5 } as const;

const ROOM_RANK: Record<RoomRole, number> = {
  member: RANK.member,
  moderator: RANK.moderator,
  admin: RANK.admin,
  owner: RANK.owner,
};

/** Effective power in a room: the higher of the room role and what the global role grants. */
export function rankOf(room: RoomRole | null | undefined, global: GlobalRole | null | undefined): number {
  if (global === 'admin') return RANK.super;
  let rank: number = room ? (ROOM_RANK[room] ?? RANK.none) : RANK.none;
  if (global === 'moderator' && rank < RANK.moderator) rank = RANK.moderator;
  return rank;
}

export const canDeleteAnyMessage = (rank: number) => rank >= RANK.moderator;
export const canPin = (rank: number) => rank >= RANK.admin;
export const canModerate = (actorRank: number, targetRank: number) =>
  actorRank >= RANK.moderator && actorRank > targetRank;

/** Roles the actor may give to a member of rank `targetRank` (empty = cannot change them). */
export function assignableRoles(actorRank: number, targetRank: number): RoomRole[] {
  if (actorRank < RANK.admin || actorRank <= targetRank) return [];
  return actorRank >= RANK.owner ? ['member', 'moderator', 'admin'] : ['member', 'moderator'];
}
