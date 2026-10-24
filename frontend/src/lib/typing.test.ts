import { describe, expect, it } from 'vitest';
import { typingLabel } from './typing';

describe('typingLabel', () => {
  it('describes zero, one, two and many typists', () => {
    expect(typingLabel([])).toBe('');
    expect(typingLabel(['Ana'])).toBe('Ana is typing…');
    expect(typingLabel(['Ana', 'Luis'])).toBe('Ana and Luis are typing…');
    expect(typingLabel(['Ana', 'Luis', 'Eva'])).toBe('Several people are typing…');
  });
});
