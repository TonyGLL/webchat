import { describe, expect, it } from 'vitest';
import { initials } from './initials';

describe('initials', () => {
  it('uses the first and last word', () => {
    expect(initials('Ana Torres')).toBe('AT');
    expect(initials('  luis   de la  moreno ')).toBe('LM');
  });

  it('falls back to one letter or a placeholder', () => {
    expect(initials('tonygll')).toBe('T');
    expect(initials('')).toBe('?');
    expect(initials(null)).toBe('?');
  });
});
