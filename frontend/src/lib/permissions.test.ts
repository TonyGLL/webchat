import { describe, expect, it } from 'vitest';
import { assignableRoles, canDeleteAnyMessage, canModerate, canPin, RANK, rankOf } from './permissions';

describe('rankOf', () => {
  it('combines room and global roles', () => {
    expect(rankOf(null, 'user')).toBe(RANK.none);
    expect(rankOf('member', 'user')).toBe(RANK.member);
    expect(rankOf('owner', 'user')).toBe(RANK.owner);
    expect(rankOf(null, 'moderator')).toBe(RANK.moderator);
    expect(rankOf('admin', 'moderator')).toBe(RANK.admin);
    expect(rankOf('owner', 'admin')).toBe(RANK.super);
  });
});

describe('capabilities', () => {
  it('matches the backend matrix', () => {
    expect(canDeleteAnyMessage(RANK.member)).toBe(false);
    expect(canDeleteAnyMessage(RANK.moderator)).toBe(true);
    expect(canPin(RANK.moderator)).toBe(false);
    expect(canPin(RANK.admin)).toBe(true);
  });

  it('moderation needs strictly more rank than the target', () => {
    expect(canModerate(RANK.moderator, RANK.member)).toBe(true);
    expect(canModerate(RANK.moderator, RANK.moderator)).toBe(false);
    expect(canModerate(RANK.admin, RANK.owner)).toBe(false);
    expect(canModerate(RANK.super, RANK.owner)).toBe(true);
    expect(canModerate(RANK.member, RANK.none)).toBe(false);
  });

  it('lists the roles an actor may assign', () => {
    expect(assignableRoles(RANK.moderator, RANK.member)).toEqual([]);
    expect(assignableRoles(RANK.admin, RANK.member)).toEqual(['member', 'moderator']);
    expect(assignableRoles(RANK.admin, RANK.admin)).toEqual([]);
    expect(assignableRoles(RANK.owner, RANK.admin)).toEqual(['member', 'moderator', 'admin']);
  });
});
